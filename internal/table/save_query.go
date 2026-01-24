package table

import (
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/editor"
	"github.com/eduardofuncao/pam/internal/styles"
)

type saveQueryCompleteMsg struct {
	success bool
	query   db.Query
}

func (m Model) saveQuery() (tea.Model, tea.Cmd) {
	// Get the current query SQL (may have been edited)
	sqlToSave := m.currentQuery.SQL
	if m.editedQuery != "" {
		sqlToSave = m.editedQuery
	}

	// Check if this is a named query that should be overwritten
	if m.isNamedQuery() {
		// Overwrite the existing query
		queryToSave := db.Query{
			Name: m.currentQuery.Name,

			SQL: sqlToSave,
			Id:  m.currentQuery.Id,
		}

		if m.saveQueryCallback != nil {
			savedQuery, err := m.saveQueryCallback(queryToSave)
			if err != nil {
				m.statusMessage = styles.Error.Render("✗ Save failed: " + err.Error())
			} else {
				m.statusMessage = styles.Success.Render("✓ Saved: " + m.currentQuery.Name)
				// Update the model with the saved query info (in case ID changed)
				m.currentQuery = savedQuery
			}
		}
		return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return blinkMsg{}
		})
	}

	// For unnamed queries, use editor to get the name
	editorCmd := editor.GetEditorCommand()
	instructions := "# Enter the name for this query\n# Lines starting with # will be ignored\n# Save and exit to confirm, or exit without saving to cancel\n"

	tmpFile, err := editor.CreateTempFile("pam-query-name-", instructions)
	if err != nil {
		m.statusMessage = lipgloss.NewStyle().Foreground(lipgloss.Color("red")).Render("Error creating temp file")
		return m, nil
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	tmpFile.Close()

	cmd := exec.Command(editorCmd, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return saveQueryCompleteMsg{success: false, query: db.Query{}}
		}

		content, readErr := editor.ReadTempFile(tmpPath)

		if readErr != nil {
			return saveQueryCompleteMsg{success: false, query: db.Query{}}
		}

		// Extract the name (first non-comment line)
		lines := strings.Split(string(content), "\n")
		var name string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				name = line
				break
			}
		}

		if name == "" {
			return saveQueryCompleteMsg{success: false, query: db.Query{}}
		}

		// Save with the new name
		queryToSave := db.Query{
			Name: name,
			SQL:  sqlToSave,
			Id:   -1, // New query
		}

		var savedQuery db.Query
		if m.saveQueryCallback != nil {
			var saveErr error
			savedQuery, saveErr = m.saveQueryCallback(queryToSave)
			if saveErr != nil {
				return saveQueryCompleteMsg{success: false, query: db.Query{}}
			}
		}

		return saveQueryCompleteMsg{success: true, query: savedQuery}
	})
}

func (m Model) handleSaveQueryComplete(msg saveQueryCompleteMsg) (tea.Model, tea.Cmd) {
	if msg.success {
		m.statusMessage = styles.Success.Render("✓ Saved as: " + msg.query.Name)
		// Update the model with the saved query info
		m.currentQuery = msg.query
	} else {
		if msg.query.Name != "" {
			m.statusMessage = styles.Error.Render("✗ Error: " + msg.query.Name)
		} else {
			m.statusMessage = styles.Error.Render("✗ Save cancelled")
		}
	}
	return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return blinkMsg{}
	})
}

func (m Model) isNamedQuery() bool {
	// Check if the query name indicates it's a saved query
	// Named queries are NOT: <inline>, <edited>, <runtime>, info commands
	name := m.currentQuery.Name
	if name == "<inline>" || name == "<edited>" || name == "<runtime>" {
		return false
	}
	// Info commands start with "info "
	if len(name) >= 4 && name[:4] == "info " {
		return false
	}
	// If it has a valid ID and name, it's a saved query
	return m.currentQuery.Id > 0 && name != ""
}
