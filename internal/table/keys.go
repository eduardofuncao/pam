// Package table implements keyboard input handling for the interactive TUI.
//
// This file contains all key binding logic and routes keypresses to appropriate
// actions based on the current mode (normal, delete, confirm, command).
package table

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyPress routes keyboard input to appropriate handlers based on current mode.
// It handles four distinct modes:
// - Confirm mode: y/n prompts for destructive actions
// - Delete mode: cell/row deletion workflow (d → r/c → enter)
// - Command mode: SQL command input (triggered by ';')
// - Normal mode: navigation, editing, and viewing
func handleKeyPress(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle confirm mode keys
	if m.confirmMode {
		switch msg.Type {
		case tea.KeyEscape, tea.KeyCtrlC:
			m.confirmMode = false
			m.confirmAction = ""
			return m, nil
		case tea.KeyEnter:
			return m.executeConfirmAction()
		default:
			return m, nil
		}
	}

	// Handle delete mode keys
	if m.deleteMode {
		switch msg.String() {
		case "esc", "ctrl+c":
			m.deleteMode = false
			m.deleteTarget = ""
			return m, nil
		case "r":
			m.deleteTarget = "row"
			return m, nil
		case "c":
			m.deleteTarget = "cell"
			return m, nil
		case "enter":
			return m.executeDelete()
		default:
			return m, nil
		}
	}

	// Handle command mode keys
	if m.commandMode {
		switch msg.Type {
		case tea.KeyEscape, tea.KeyCtrlC:
			m.commandMode = false
			m.commandInput.Reset()
			return m, nil
		case tea.KeyEnter:
			return m.runCommand(m.commandInput.Value())
		default:
			var cmd tea.Cmd
			m.commandInput, cmd = m.commandInput.Update(msg)
			return m, cmd
		}
	}

	// Normal mode keys
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case ";":
		m.commandMode = true
		m.commandInput.Focus()
		return m, nil

	case "up", "k":
		m.clearStatus()
		return m.moveUp(), nil
	case "down", "j":
		m.clearStatus()
		return m.moveDown(), nil
	case "left", "h":
		m.clearStatus()
		return m.moveLeft(), nil
	case "right", "l":
		m.clearStatus()
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

	case "y":
		return m.copySelection()

	case "e":
		return m.editCell()

	case "d":
		return m.enterDeleteMode()
	}

	return m, nil
}
