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
	if m.selectedRow < 0 || m. selectedRow >= m.numRows() {
		return m, nil
	}
	if m.selectedCol < 0 || m. selectedCol >= m.numCols() {
		return m, nil
	}

	// Capture the old value
	// oldValue := m.data[m.selectedRow][m.selectedCol]

	updateStmt := m.buildUpdateStatement()

	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
	}

	tmpFile, err := os.CreateTemp("", "pam-update-*.sql")
	if err != nil {
		fmt. Fprintf(os.Stderr, "Error creating temp file: %v\n", err)
		return m, tea.Quit
	}
	tmpPath := tmpFile. Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write([]byte(updateStmt)); err != nil {
		tmpFile.Close()
		fmt.Fprintf(os.Stderr, "Error writing temp file: %v\n", err)
		return m, tea.Quit
	}
	tmpFile.Close()

	tea.ExitAltScreen()

	cmd := exec.Command(editorCmd, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd. Stderr = os.Stderr
	
	if err := cmd. Run(); err != nil {
		fmt.Fprintf(os. Stderr, "Error running editor: %v\n", err)
		tea.EnterAltScreen()
		return m, tea.Quit
	}

	editedSQL, err := os.ReadFile(tmpPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading edited file: %v\n", err)
		tea.EnterAltScreen()
		return m, tea.Quit
	}

	newValue := extractNewValue(string(editedSQL), m.columns[m.selectedCol])

	tea.EnterAltScreen()

	if err := m.executeUpdate(string(editedSQL)); err != nil {
		return m, nil
	}
	
	m.data[m.selectedRow][m.selectedCol] = newValue
	
	m.blinkUpdatedCell = true
	m.updatedRow = m.selectedRow
	m.updatedCol = m.selectedCol
	
	return m, m.blinkCmd()
}

func (m Model) blinkCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return blinkMsg{}
	})
}

func (m Model) buildUpdateStatement() string {
	columnName := m.columns[m.selectedCol]
	currentValue := m. data[m.selectedRow][m.selectedCol]
	
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
	for line := range strings. SplitSeq(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") && trimmed != "" {
			result.WriteString(trimmed)
			result.WriteString(" ")
		}
	}
	
	cleanSQL := strings.TrimSpace(result.String())
	cleanSQL = strings.TrimSuffix(cleanSQL, ";")
	
	if cleanSQL == "" {
		return fmt. Errorf("no SQL to execute")
	}
	
	return m.dbConnection.Exec(cleanSQL)
}

// extractNewValue parses the SQL UPDATE statement to extract the new value
func extractNewValue(sql string, columnName string) string {
	// Remove comments and consolidate to single line
	var result strings.Builder
	for line := range strings.SplitSeq(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") && trimmed != "" {
			result.WriteString(trimmed)
			result.WriteString(" ")
		}
	}
	cleanSQL := strings.TrimSpace(result. String())
	
	// Try to match: SET column_name = 'value' or SET column_name = value
	// This regex handles both quoted and unquoted values
	pattern := fmt.Sprintf(`SET\s+%s\s*=\s*('([^']*)'|"([^"]*)"|([^,\s;]+))`, regexp.QuoteMeta(columnName))
	re := regexp.MustCompile(`(?i)` + pattern)
	
	matches := re. FindStringSubmatch(cleanSQL)
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
