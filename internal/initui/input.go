package initui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/styles"
)

const (
	fieldName = iota
	fieldType
	fieldConnString
)

type InitInputModel struct {
	name           string
	dbType         string
	connString     string
	cursorIndex    int  // Which field is focused
	nameCursor     int  // Cursor position within name field
	connCursor     int  // Cursor position within conn-string field
	dbTypes        []string
	aborted        bool
}

func NewInitInputModel(name, dbType, connString string) InitInputModel {
	types := db.GetSupportedDBTypes()

	// Auto-infer db type if conn string provided but type is empty
	if dbType == "" && connString != "" {
		dbType = db.InferDBType(connString)
	}

	return InitInputModel{
		name:        name,
		dbType:      dbType,
		connString:  connString,
		cursorIndex: 0,
		nameCursor:  len(name),
		connCursor:  len(connString),
		dbTypes:     types,
		aborted:     false,
	}
}

func (m InitInputModel) Init() tea.Cmd {
	return nil
}

func (m InitInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit

		case "enter":
			// Check if all required fields are filled
			if m.name == "" || m.dbType == "" || m.connString == "" {
				// Move to next empty field
				m.moveToNextEmptyField()
				return m, nil
			}
			// All fields filled, submit and quit
			return m, tea.Quit

		case "down":
			if m.cursorIndex < fieldConnString {
				m.cursorIndex++
				// Update cursor position for the new field
				if m.cursorIndex == fieldName {
					m.nameCursor = len(m.name)
				} else if m.cursorIndex == fieldConnString {
					m.connCursor = len(m.connString)
				}
			}

		case "up":
			if m.cursorIndex > fieldName {
				m.cursorIndex--
				// Update cursor position for the new field
				if m.cursorIndex == fieldName {
					m.nameCursor = len(m.name)
				} else if m.cursorIndex == fieldConnString {
					m.connCursor = len(m.connString)
				}
			}

		case "right":
			// Cycle through db types when on type field
			if m.cursorIndex == fieldType {
				m.cycleDbType(1)
			} else if m.cursorIndex == fieldName {
				// Move cursor right within name field
				if m.nameCursor < len(m.name) {
					m.nameCursor++
				}
			} else if m.cursorIndex == fieldConnString {
				// Move cursor right within conn-string field
				if m.connCursor < len(m.connString) {
					m.connCursor++
				}
			}

		case "left":
			// Cycle through db types when on type field
			if m.cursorIndex == fieldType {
				m.cycleDbType(-1)
			} else if m.cursorIndex == fieldName {
				// Move cursor left within name field
				if m.nameCursor > 0 {
					m.nameCursor--
				}
			} else if m.cursorIndex == fieldConnString {
				// Move cursor left within conn-string field
				if m.connCursor > 0 {
					m.connCursor--
				}
			}

		case "backspace":
			m.handleBackspace()

		default:
			// Handle regular character input (including pasted text)
			m.handleInput(msg.String())
		}
	}

	return m, nil
}

func (m *InitInputModel) handleBackspace() {
	switch m.cursorIndex {
	case fieldName:
		if m.nameCursor > 0 {
			m.name = m.name[:m.nameCursor-1] + m.name[m.nameCursor:]
			m.nameCursor--
		}
	case fieldConnString:
		if m.connCursor > 0 {
			m.connString = m.connString[:m.connCursor-1] + m.connString[m.connCursor:]
			m.connCursor--
		}
	}
}

func (m *InitInputModel) handleInput(ch string) {
	// Ignore empty strings
	if ch == "" {
		return
	}

	// Handle bracketed paste mode - some terminals wrap pasted content
	// If the pasted content starts with [ and ends with ], strip them
	if len(ch) > 1 && strings.HasPrefix(ch, "[") && strings.HasSuffix(ch, "]") {
		ch = ch[1 : len(ch)-1]
	}

	switch m.cursorIndex {
	case fieldName:
		// Insert at cursor position
		m.name = m.name[:m.nameCursor] + ch + m.name[m.nameCursor:]
		m.nameCursor += len(ch)
	case fieldConnString:
		// Insert at cursor position
		m.connString = m.connString[:m.connCursor] + ch + m.connString[m.connCursor:]
		m.connCursor += len(ch)
	case fieldType:
		// Only accept single characters for type selection (numbers or letters)
		// Multi-character paste is ignored for type field
		if len(ch) == 1 {
			// Allow typing number to select db type
			if ch >= "1" && ch <= "9" {
				idx := int(ch[0] - '1')
				if idx < len(m.dbTypes) {
					m.dbType = m.dbTypes[idx]
				}
			}
		}
	}
}

func (m *InitInputModel) cycleDbType(dir int) {
	if len(m.dbTypes) == 0 {
		return
	}

	currentIdx := -1
	for i, t := range m.dbTypes {
		if t == m.dbType {
			currentIdx = i
			break
		}
	}

	if currentIdx == -1 {
		// Type not in list, select first
		m.dbType = m.dbTypes[0]
		return
	}

	newIdx := (currentIdx + dir) % len(m.dbTypes)
	if newIdx < 0 {
		newIdx = len(m.dbTypes) - 1
	}
	m.dbType = m.dbTypes[newIdx]
}

func (m *InitInputModel) moveToNextEmptyField() {
	// Check each field in order and move to the first empty one
	if m.name == "" {
		m.cursorIndex = fieldName
		m.nameCursor = len(m.name)
	} else if m.dbType == "" {
		m.cursorIndex = fieldType
	} else if m.connString == "" {
		m.cursorIndex = fieldConnString
		m.connCursor = len(m.connString)
	}
	// If all fields are filled, don't move cursor (will submit on next enter)
}

func (m InitInputModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(styles.Title.Render("Initialize new connection"))
	b.WriteString("\n")

	// Name field
	m.renderField(&b, "Connection name", m.name, m.nameCursor, fieldName == m.cursorIndex)

	// DB Type field (dropdown)
	m.renderTypeDropdown(&b)

	// Connection string field
	m.renderField(&b, "Connection string", m.connString, m.connCursor, m.cursorIndex == fieldConnString)

	b.WriteString("\n")
	b.WriteString(styles.Faint.Render("↑: up  ↓: down  ←/→: move cursor/cycle type  Type: input/paste  Enter: submit  Esc: cancel"))

	return b.String()
}

func (m InitInputModel) renderField(b *strings.Builder, label, value string, cursorPos int, focused bool) {
	if focused {
		prompt := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ActiveScheme.Primary)).
			Bold(true).
			Render(label + " > ")

		// Split value at cursor position
		before := value[:cursorPos]
		after := value[cursorPos:]

		inputBox := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ActiveScheme.Muted)).
			Render(before + "▏" + after)

		b.WriteString(prompt + inputBox + "\n")
	} else {
		prompt := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ActiveScheme.Muted)).
			Render(label + "   ")

		inputBox := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ActiveScheme.Muted)).
			Render(value)

		b.WriteString(prompt + inputBox + "\n")
	}
}

func (m InitInputModel) renderTypeDropdown(b *strings.Builder) {
	label := "Database type"

	if m.cursorIndex == fieldType {
		prompt := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ActiveScheme.Primary)).
			Bold(true).
			Render(label + " > ")

		// Show current selection + dropdown indicator
		displayValue := m.dbType
		if displayValue == "" {
			displayValue = "<select>"
		}
		inputBox := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ActiveScheme.Muted)).
			Render(displayValue + " ▼")

		b.WriteString(prompt + inputBox + "\n")

		// Show available types below
		b.WriteString(styles.Faint.Render("  Available: "))
		for i, t := range m.dbTypes {
			if i > 0 {
				b.WriteString(styles.Faint.Render(", "))
			}
			if t == m.dbType {
				b.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(styles.ActiveScheme.Primary)).
					Bold(true).
					Render(t))
			} else {
				b.WriteString(styles.Faint.Render(t))
			}
		}
		b.WriteString("\n")
	} else {
		prompt := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ActiveScheme.Muted)).
			Render(label + "   ")

		displayValue := m.dbType
		if displayValue == "" {
			displayValue = "<select>"
		}
		inputBox := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ActiveScheme.Muted)).
			Render(displayValue)

		b.WriteString(prompt + inputBox + "\n")
	}
}

func (m InitInputModel) GetName() string {
	return m.name
}

func (m InitInputModel) GetDBType() string {
	return m.dbType
}

func (m InitInputModel) GetConnString() string {
	return m.connString
}

func (m InitInputModel) WasAborted() bool {
	return m.aborted
}

func CollectInitParameters(name, dbType, connString string) (string, string, string, error) {
	model := NewInitInputModel(name, dbType, connString)
	program := tea.NewProgram(model)

	finalModel, err := program.Run()
	if err != nil {
		return "", "", "", err
	}

	inputModel := finalModel.(InitInputModel)
	if inputModel.WasAborted() {
		return "", "", "", ErrAborted
	}

	return inputModel.GetName(), inputModel.GetDBType(), inputModel.GetConnString(), nil
}

var ErrAborted = fmt.Errorf("init input aborted")
