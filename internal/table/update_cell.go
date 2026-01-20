package table

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateCell() (tea.Model, tea.Cmd) {
	if m.selectedRow < 0 || m.selectedRow >= m.numRows() {
		return m, nil
	}
	if m.selectedCol < 0 || m.selectedCol >= m.numCols() {
		return m, nil
	}

	updateStmt := m.buildUpdateStatement()
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
	}

	tmpFile, err := os.CreateTemp("", "pam-update-*.sql")
	if err != nil {
		return m, nil
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write([]byte(updateStmt)); err != nil {
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

	cmd := buildEditorCommand(editorCmd, tmpPath, updateStmt, CursorAtUpdateValue)

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
			return editorCompleteMsg{
				sql:      "",
				colIndex: m.selectedCol,
				cancelled: true,
			}
		}

		editedSQL, readErr := os.ReadFile(tmpPath)
		os.Remove(tmpPath)

		if err != nil || readErr != nil {
			return nil
		}
		return editorCompleteMsg{
			sql:      string(editedSQL),
			colIndex: m.selectedCol,
			cancelled: false,
		}
	})
}

type editorCompleteMsg struct {
	sql       string
	colIndex  int
	cancelled bool
}

func (m Model) handleEditorComplete(msg editorCompleteMsg) (tea.Model, tea.Cmd) {
	// If user cancelled (exited without saving)
	if msg.cancelled {
		m.statusMessage = "Update cancelled"
		return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return blinkMsg{}
		})
	}

	if err := validateUpdateStatement(msg.sql); err != nil {
		printError("Update validation failed:  %v", err)
		return m, nil
	}

	newValue := m.extractNewValue(msg.sql, m.columns[msg.colIndex])

	m.lastExecutedQuery = m.cleanSQLForDisplay(msg.sql)

	if err := m.executeUpdate(msg.sql); err != nil {
		printError("Could not execute update: %v", err)
		return m, nil
	}

	m.data[m.selectedRow][msg.colIndex] = newValue

	m.blinkUpdatedCell = true
	m.updatedRow = m.selectedRow
	m.updatedCol = msg.colIndex

	return m, tea.Batch(
		tea.ClearScreen,
		m.blinkCmd(),
	)
}

func (m Model) blinkCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return blinkMsg{}
	})
}

func (m Model) buildUpdateStatement() string {
	columnName := m.columns[m.selectedCol]
	currentValue := m.data[m.selectedRow][m.selectedCol]

	pkValue := ""
	if m.primaryKeyCol != "" {
		for i, col := range m.columns {
			if col == m.primaryKeyCol {
				pkValue = m.data[m.selectedRow][i]
				break
			}
		}
	}

	return m.dbConnection.BuildUpdateStatement(
		m.tableName,
		columnName,
		currentValue,
		m.primaryKeyCol,
		pkValue,
	)
}

func (m Model) executeUpdate(sql string) error {
	var result strings.Builder
	for line := range strings.SplitSeq(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") && trimmed != "" {
			result.WriteString(trimmed)
			result.WriteString(" ")
		}
	}

	cleanSQL := strings.TrimSpace(result.String())
	cleanSQL = strings.TrimSuffix(cleanSQL, ";")

	if cleanSQL == "" {
		return fmt.Errorf("no SQL to execute")
	}

	return m.dbConnection.Exec(cleanSQL)
}

func validateUpdateStatement(sql string) error {
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

	// Check for ClickHouse ALTER TABLE UPDATE or standard UPDATE
	isClickHouse := strings.Contains(upperSQL, "ALTER TABLE") && strings.Contains(upperSQL, "UPDATE")
	isStandardUpdate := strings.HasPrefix(upperSQL, "UPDATE")

	if !isClickHouse && !isStandardUpdate {
		return fmt.Errorf("not a valid UPDATE statement (expected UPDATE or ALTER TABLE UPDATE)")
	}

	// For ClickHouse: ALTER TABLE ... UPDATE ... WHERE ...
	if isClickHouse {
		// Check for UPDATE keyword after ALTER TABLE
		updateRegex := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+\S+\s+UPDATE\s+`)
		if !updateRegex.MatchString(cleanSQL) {
			return fmt.Errorf("ClickHouse ALTER TABLE UPDATE must include UPDATE clause")
		}
	} else {
		// For standard SQL: UPDATE ... SET ...
		setRegex := regexp.MustCompile(`(?i)\bSET\b`)
		if !setRegex.MatchString(cleanSQL) {
			return fmt.Errorf("UPDATE statement must include a SET clause")
		}
	}

	// Both syntaxes require WHERE clause
	whereRegex := regexp.MustCompile(`(?i)\bWHERE\b`)
	if !whereRegex.MatchString(cleanSQL) {
		return fmt.Errorf("UPDATE statement must include a WHERE clause for safety")
	}

	return nil
}

func (m Model) cleanSQLForDisplay(sql string) string {
	var result strings.Builder
	for line := range strings.SplitSeq(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") && trimmed != "" {
			result.WriteString(trimmed)
			result.WriteString(" ")
		}
	}

	cleanSQL := strings.TrimSpace(result.String())
	return cleanSQL
}

  func (m Model) extractNewValue(sql string, columnName string) string {
      var result strings.Builder
      for line := range strings.SplitSeq(sql, "\n") {
          trimmed := strings.TrimSpace(line)
          if !strings.HasPrefix(trimmed, "--") && trimmed != "" {
              result.WriteString(trimmed)
              result.WriteString(" ")
          }
      }
      cleanSQL := strings.TrimSpace(result.String())

      // First try standard SQL: SET column_name = 'value'
      setPattern := fmt.Sprintf(`SET\s+%s\s*=\s*('([^']*)'|"([^"]*)"|([^,\s;]+))`, regexp.QuoteMeta(columnName))
      setRe := regexp.MustCompile(`(?i)` + setPattern)

      matches := setRe.FindStringSubmatch(cleanSQL)
      if len(matches) > 0 {
          if matches[2] != "" {
              return matches[2]
          } else if matches[3] != "" {
              return matches[3]
          } else if matches[4] != "" {
              return matches[4]
          }
      }

      // Try ClickHouse: UPDATE column_name = 'value' (no SET keyword)
      updatePattern := fmt.Sprintf(`UPDATE\s+%s\s*=\s*('([^']*)'|"([^"]*)"|([^,\s;]+))`,
  regexp.QuoteMeta(columnName))
      updateRe := regexp.MustCompile(`(?i)` + updatePattern)

      matches = updateRe.FindStringSubmatch(cleanSQL)
      if len(matches) > 0 {
          if matches[2] != "" {
              return matches[2]
          } else if matches[3] != "" {
              return matches[3]
          } else if matches[4] != "" {
              return matches[4]
          }
      }

      return "<unknown>"
  }
