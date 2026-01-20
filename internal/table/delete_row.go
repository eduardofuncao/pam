package table

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) deleteRow() (tea.Model, tea.Cmd) {
	if m.selectedRow < 0 || m.selectedRow >= m.numRows() {
		return m, nil
	}

	if m.primaryKeyCol == "" {
		return m, nil
	}

	deleteStmt := m.buildDeleteStatement()

	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
	}

	tmpFile, err := os.CreateTemp("", "pam-delete-*.sql")
	if err != nil {
		return m, nil
	}
	tmpPath := tmpFile.Name()

	header := `-- DELETE Statement
-- WARNING: This will permanently delete data!
-- To cancel, exit without saving (e.g., :q! in vim)
--
`
	content := header + deleteStmt

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return m, nil
	}
	tmpFile.Close()

	// Get file modification time before editor
	beforeModTime, err := os.Stat(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return m, nil
	}

	rowToDelete := m.selectedRow

	// Build command with cursor at WHERE clause
	cmd := buildEditorCommand(editorCmd, tmpPath, content, CursorAtWhereClause)

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		// Check if file was modified
		afterModTime, statErr := os.Stat(tmpPath)
		if statErr != nil {
			os.Remove(tmpPath)
			return nil
		}

		// If file wasn't modified, user cancelled (exited without saving)
		if afterModTime.ModTime().Equal(beforeModTime.ModTime()) || afterModTime.ModTime().Before(beforeModTime.ModTime()) {
			os.Remove(tmpPath)
			// Return a message that will show "cancelled" status
			return deleteCompleteMsg{
				sql:       "",
				rowIndex:  rowToDelete,
				cancelled: true,
			}
		}

		editedSQL, readErr := os.ReadFile(tmpPath)
		os.Remove(tmpPath)

		if err != nil || readErr != nil {
			return nil
		}

		return deleteCompleteMsg{
			sql:       string(editedSQL),
			rowIndex:  rowToDelete,
			cancelled: false,
		}
	})
}

// Message sent when delete editor completes
type deleteCompleteMsg struct {
	sql       string
	rowIndex  int
	cancelled bool
}

func (m Model) handleDeleteComplete(msg deleteCompleteMsg) (tea.Model, tea.Cmd) {
	// If user cancelled (exited without saving)
	if msg.cancelled {
		m.statusMessage = "Delete cancelled"
		return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return blinkMsg{}
		})
	}

	if err := validateDeleteStatement(msg.sql); err != nil {
		printError("Delete validation failed: %v", err)
		return m, nil
	}

	m.lastExecutedQuery = m.cleanSQLForDisplay(msg.sql)

	if err := m.executeDelete(msg.sql); err != nil {
		printError("Could not execute delete: %v", err)
		return m, nil
	}

	// Successfully deleted - update the model data
	m.data = append(m.data[:msg.rowIndex], m.data[msg.rowIndex+1:]...)
	if m.selectedRow >= m.numRows() && m.numRows() > 0 {
		m.selectedRow = m.numRows() - 1
	}
	if m.offsetY >= m.numRows() && m.numRows() > 0 {
		m.offsetY = m.numRows() - 1
	}

	m.blinkDeletedRow = true
	m.deletedRow = m.selectedRow

	return m, tea.Batch(
		tea.ClearScreen,
		m.blinkCmd(),
	)
}

func (m Model) buildDeleteStatement() string {
	pkValue := ""
	if m.primaryKeyCol != "" {
		for i, col := range m.columns {
			if col == m.primaryKeyCol {
				pkValue = m.data[m.selectedRow][i]
				break
			}
		}
	}

	return m.dbConnection.BuildDeleteStatement(
		m.tableName,
		m.primaryKeyCol,
		pkValue,
	)
}

func (m Model) executeDelete(sql string) error {
	var result strings.Builder
	for line := range strings.SplitSeq(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") && trimmed != "" {
			result. WriteString(trimmed)
			result.WriteString(" ")
		}
	}

	cleanSQL := strings.TrimSpace(result.String())
	cleanSQL = strings.TrimSuffix(cleanSQL, ";")

	if cleanSQL == "" {
		return fmt.Errorf("no SQL to execute")
	}

	return m.dbConnection. Exec(cleanSQL)
}

  func validateDeleteStatement(sql string) error {
      var result strings.Builder
      for line := range strings.SplitSeq(sql, "\n") {
          trimmed := strings.TrimSpace(line)
          if !strings.HasPrefix(trimmed, "--") && trimmed != "" {
              result.WriteString(trimmed)
              result.WriteString(" ")
          }
      }
      cleanSQL := strings.TrimSpace(result.String())

      if cleanSQL == "" {
          return fmt.Errorf("empty SQL statement")
      }

      upperSQL := strings.ToUpper(cleanSQL)

      // Check for ClickHouse ALTER TABLE DELETE or standard DELETE FROM
      isClickHouse := strings.Contains(upperSQL, "ALTER TABLE") && strings.Contains(upperSQL, "DELETE")
      isStandardDelete := strings.HasPrefix(upperSQL, "DELETE")

      if !isClickHouse && !isStandardDelete {
          return fmt.Errorf("not a valid DELETE statement (expected DELETE FROM or ALTER TABLE DELETE)")
      }

      // For ClickHouse: ALTER TABLE ... DELETE ...
      if isClickHouse {
          // Check for DELETE keyword after ALTER TABLE
          alterDeleteRegex := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+\S+\s+DELETE`)
          if !alterDeleteRegex.MatchString(cleanSQL) {
              return fmt.Errorf("ClickHouse ALTER TABLE DELETE must include DELETE clause")
          }
      } else {
          // For standard SQL: DELETE FROM ...
          deleteFromRegex := regexp.MustCompile(`(?i)DELETE\s+FROM\s+`)
          if !deleteFromRegex.MatchString(cleanSQL) {
              return fmt.Errorf("DELETE statement must include FROM clause")
          }
      }

      // Both syntaxes require WHERE clause
      whereRegex := regexp.MustCompile(`(?i)\bWHERE\b`)
      if !whereRegex.MatchString(cleanSQL) {
          return fmt.Errorf("DELETE statement must include a WHERE clause for safety")
      }

      return nil
  }
