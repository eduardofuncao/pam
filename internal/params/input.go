package params

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eduardofuncao/pam/internal/parser"
	"github.com/eduardofuncao/pam/internal/styles"
)

type InputModel struct {
	sql           string
	missingParams []string
	defaults      map[string]string
	currentValues map[string]string
	cursorIndex   int
	aborted       bool
}

func NewInputModel(sql string, missingParams []string, defaults map[string]string) InputModel {
	currentValues := make(map[string]string)
	for _, param := range missingParams {
		if def, ok := defaults[param]; ok {
			currentValues[param] = def
		} else {
			currentValues[param] = ""
		}
	}

	return InputModel{
		sql:           sql,
		missingParams: missingParams,
		defaults:      defaults,
		currentValues: currentValues,
		cursorIndex:   0,
		aborted:       false,
	}
}

func (m InputModel) Init() tea.Cmd {
	return nil
}

func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.aborted = true
			return m, tea.Quit

		case "enter":
			// Submit and quit
			return m, tea.Quit

		case "down":
			if m.cursorIndex < len(m.missingParams)-1 {
				m.cursorIndex++
			}

		case "up":
			if m.cursorIndex > 0 {
				m.cursorIndex--
			}

		case "backspace":
			currentParam := m.missingParams[m.cursorIndex]
			currentVal := m.currentValues[currentParam]
			if len(currentVal) > 0 {
				m.currentValues[currentParam] = currentVal[:len(currentVal)-1]
			}

		default:
			// Handle regular character input
			currentParam := m.missingParams[m.cursorIndex]
			// Allow typing any character
			m.currentValues[currentParam] += msg.String()
		}
	}

	return m, nil
}

func (m InputModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(styles.Title.Render("Enter runtime params"))
	b.WriteString("\n")

	// SQL display (formatted with line breaks and syntax highlighting)
	formattedSQL := parser.FormatSQLWithLineBreaks(m.sql)
	highlightedSQL := parser.HighlightSQL(formattedSQL)
	b.WriteString(highlightedSQL)
	b.WriteString("\n\n")

	// Parameter input fields
	for i, param := range m.missingParams {
		currentValue := m.currentValues[param]

		// Style differently for focused vs unfocused
		if i == m.cursorIndex {
			// Focused field
			prompt := lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.ColorAccent)).
				Bold(true).
				Render(param + " > ")

			inputBox := lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.ColorCellNormal)).
				Render(currentValue + "▏")

			b.WriteString(prompt + inputBox + "\n")
		} else {
			// Unfocused field
			prompt := lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.ColorCellNormal)).
				Render(param + "   ")

			inputBox := lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.ColorCellNormal)).
				Render(currentValue)

			b.WriteString(prompt + inputBox + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.Faint.Render("↑: up  ↓: down  Enter: submit  Esc/q: cancel"))

	return b.String()
}

func (m InputModel) GetValues() map[string]string {
	return m.currentValues
}

func (m InputModel) WasAborted() bool {
	return m.aborted
}

func CollectParameters(sql string, missingParams []string, defaults map[string]string) (map[string]string, error) {
	model := NewInputModel(sql, missingParams, defaults)
	program := tea.NewProgram(model)

	finalModel, err := program.Run()
	if err != nil {
		return nil, err
	}

	inputModel := finalModel.(InputModel)
	if inputModel.WasAborted() {
		return nil, ErrAborted
	}

	return inputModel.GetValues(), nil
}

var ErrAborted = fmt.Errorf("parameter input aborted")
