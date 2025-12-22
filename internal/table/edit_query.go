package table

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) editAndRerunQuery() (tea.Model, tea.Cmd) {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
	}

	tmpFile, err := os.CreateTemp("", "pam-edit-query-*.sql")
	if err != nil {
		return m, nil
	}
	tmpPath := tmpFile.Name()

	// Write current query to temp file
	if _, err := tmpFile.WriteString(m.currentQuery. SQL); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return m, nil
	}
	tmpFile.Close()

	// Use tea.ExecProcess to run the editor
	cmd := exec. Command(editorCmd, tmpPath)
	
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		// Read the edited file BEFORE removing it
		editedData, readErr := os.ReadFile(tmpPath)
		
		// Now remove the temp file
		os.Remove(tmpPath)
		
		if err != nil || readErr != nil {
			return nil
		}

		editedSQL := strings.TrimSpace(string(editedData))
		if editedSQL == "" {
			return nil
		}

		return queryEditCompleteMsg{
			sql:  editedSQL,
		}
	})
}

// Message sent when query editor completes
type queryEditCompleteMsg struct {
	sql string
}

// Handle the query edit complete message
func (m Model) handleQueryEditComplete(msg queryEditCompleteMsg) (tea.Model, tea.Cmd) {
	// Store the edited query and signal that we should re-run
	m.editedQuery = msg.sql
	m.shouldRerunQuery = true

	// Quit the TUI - the calling code will handle re-execution
	return m, tea.Quit
}
