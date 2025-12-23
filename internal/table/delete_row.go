package table

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

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
-- Ensure the WHERE clause is present and correct before saving.  
-- To cancel, delete all content and save.    
--
`
	content := header + deleteStmt

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return m, nil
	}
	tmpFile.Close()

	// Store the row index before entering async operation
	rowToDelete := m.selectedRow

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

		return deleteCompleteMsg{
			sql:      string(editedSQL),
			rowIndex: rowToDelete,
		}
	})
}

// Message sent when delete editor completes
type deleteCompleteMsg struct {
	sql      string
	rowIndex int
}

func (m Model) handleDeleteComplete(msg deleteCompleteMsg) (tea.Model, tea.Cmd) {
	// Validate the delete statement
	if err := validateDeleteStatement(msg.sql); err != nil {
		printError("Delete validation failed: %v", err)
	}

	// Store the cleaned SQL for display
	m.lastExecutedQuery = m.cleanSQLForDisplay(msg.sql)

	// Execute the delete
	if err := m.executeDelete(msg.sql); err != nil {
		printError("Could not execute delete: %v", err)
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

	// Force a full re-render with screen clear
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

	deleteRegex := regexp.MustCompile(`(?i)^\s*DELETE\s+FROM`)
	if ! deleteRegex.MatchString(cleanSQL) {
		return fmt.Errorf("not a DELETE statement")
	}

	whereRegex := regexp.MustCompile(`(?i)\bWHERE\b`)
	if !whereRegex.MatchString(cleanSQL) {
		return fmt.Errorf("DELETE statement must include a WHERE clause for safety")
	}

	return nil
}
