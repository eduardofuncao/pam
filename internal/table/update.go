package table

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case blinkMsg:
		m.blinkCopiedCell = false
		m.blinkUpdatedCell = false
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg), nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		return m.moveUp(), nil
	case "down", "j":
		return m.moveDown(), nil
	case "left", "h":
		return m.moveLeft(), nil
	case "right", "l":
		return m.moveRight(), nil

	case "home", "0", "_":
		return m.jumpToFirstCol(), nil
	case "end", "$":
		return m.jumpToLastCol(), nil
	case "g":
		return m.jumpToFirstRow(), nil
	case "G":
		return m.jumpToLastRow(), nil

	case "pgup", "ctrl+u":
		return m.pageUp(), nil
	case "pgdown", "ctrl+d":
		return m.pageDown(), nil

	case "v":
		return m.toggleVisualMode()

	case "y", "enter":
		return m.copySelection()

	case "u":
		return m.updateCell()
	}

	return m, nil
}

func (m Model) handleWindowResize(msg tea.WindowSizeMsg) Model {
	m.width = msg.Width
	m.height = msg.Height

	m.visibleCols = (m.width - 2) / (cellWidth + 1)
	if m.visibleCols > m.numCols() {
		m.visibleCols = m.numCols()
	}

	// Offset so table doesn't take full terminal height
	m.visibleRows = m.height - 9
	if m.visibleRows > m.numRows() {
		m.visibleRows = m.numRows()
	}
	if m.visibleRows < 1 {
		m.visibleRows = 1
	}

	return m
}
