package table

import (
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eduardofuncao/pam/internal/styles"
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

	if _, err := tmpFile.WriteString(m.currentQuery. SQL); err != nil {
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

	cmd := buildEditorCommand(editorCmd, tmpPath, m.currentQuery. SQL, CursorAtEndOfFile)

	return m, tea. ExecProcess(cmd, func(err error) tea.Msg {
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
			return queryEditCompleteMsg{
				sql:       "",
				cancelled: true,
			}
		}
		editedData, readErr := os.ReadFile(tmpPath)
		os. Remove(tmpPath)
		
		if err != nil || readErr != nil {
			return nil
		}

		editedSQL := strings.TrimSpace(string(editedData))
		if editedSQL == "" {
			return nil
		}

		return queryEditCompleteMsg{
			sql:       editedSQL,
			cancelled: false,
		}
	})
}

type queryEditCompleteMsg struct {
	sql       string
	cancelled bool
}

func (m Model) handleQueryEditComplete(msg queryEditCompleteMsg) (tea.Model, tea.Cmd) {
	// If user cancelled (exited without saving)
	if msg.cancelled {
		m.statusMessage = styles.Error.Render("✗ Edit canceled")
		return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return blinkMsg{}
		})
	}

	m.editedQuery = msg.sql
	m.shouldRerunQuery = true

	return m, tea.Quit
}
