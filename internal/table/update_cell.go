package table

import (
	"fmt"
	"os"
	"os/exec"
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

	// Use tea.ExecProcess to run the editor
	cmd := exec.Command(editorCmd, tmpPath)

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		// Read the edited file BEFORE removing it
		editedSQL, readErr := os.ReadFile(tmpPath)

		// Now remove the temp file
		os.Remove(tmpPath)

		if err != nil || readErr != nil {
			return nil
		}
		return editorCompleteMsg{
			sql:      string(editedSQL),
			colIndex: m.selectedCol,
		}
	})
}

// Message sent when editor completes
type editorCompleteMsg struct {
	sql      string
	colIndex int
}

func (m Model) handleEditorComplete(msg editorCompleteMsg) (tea.Model, tea.Cmd) {
	newValue := m.extractNewValue(msg.sql, m.columns[msg.colIndex])

	// Store the cleaned SQL for display
	m.lastExecutedQuery = m.cleanSQLForDisplay(msg.sql)

	// Execute the update
	if err := m.executeUpdate(msg.sql); err != nil {
		printError("Could not execute update: %v", err)
	}

	// Successfully updated - update the model data in-place
	m.data[m.selectedRow][msg.colIndex] = newValue

	m.blinkUpdatedCell = true
	m.updatedRow = m.selectedRow
	m.updatedCol = msg.colIndex

	// Force a full re-render with screen clear
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

// cleanSQLForDisplay removes comments and formats SQL for display
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

// extractNewValue parses the SQL UPDATE statement to extract the new value
func (m Model) extractNewValue(sql string, columnName string) string {
	// Remove comments and consolidate to single line
	var result strings.Builder
	for line := range strings.SplitSeq(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") && trimmed != "" {
			result.WriteString(trimmed)
			result.WriteString(" ")
		}
	}
	cleanSQL := strings.TrimSpace(result.String())

	// Try to match:   SET column_name = 'value' or SET column_name = value
	// This regex handles both quoted and unquoted values
	pattern := fmt.Sprintf(`SET\s+%s\s*=\s*('([^']*)'|"([^"]*)"|([^,\s;]+))`, regexp.QuoteMeta(columnName))
	re := regexp.MustCompile(`(?i)` + pattern)

	matches := re.FindStringSubmatch(cleanSQL)
	if len(matches) > 0 {
		// matches[2] = single quoted value
		// matches[3] = double quoted value
		// matches[4] = unquoted value
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
