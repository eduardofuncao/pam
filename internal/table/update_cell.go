package table

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
		return m, tea.Quit
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write([]byte(updateStmt)); err != nil {
		tmpFile.Close()
		return m, tea.Quit
	}
	tmpFile.Close()

	cmd := exec.Command(editorCmd, tmpPath)
	cmd. Stdin = os.Stdin
	cmd. Stdout = os.Stdout
	cmd.Stderr = os. Stderr
	
	if err := cmd.Run(); err != nil {
		return m, tea.Quit
	}

	editedSQL, err := os.ReadFile(tmpPath)
	if err != nil {
		return m, tea.Quit
	}

	if err := m.executeUpdate(string(editedSQL)); err != nil {
		fmt. Fprintf(os.Stderr, "Error executing update: %v\n", err)
	} else {
		fmt.Println("\nUpdate executed successfully")
	}

	return m, tea.Quit
}

func (m Model) buildUpdateStatement() string {
	columnName := m.columns[m.selectedCol]
	currentValue := m.data[m. selectedRow][m.selectedCol]
	
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
	lines := strings.Split(sql, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if ! strings.HasPrefix(trimmed, "--") && trimmed != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	cleanSQL := strings.TrimSpace(strings.Join(cleanLines, "\n"))
	
	if cleanSQL == "" {
		return fmt.Errorf("no SQL to execute")
	}
	
	return m.dbConnection.Exec(cleanSQL)
}
