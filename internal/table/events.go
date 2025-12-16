// Package table implements event handling for the interactive TUI table viewer.
//
// This file contains the Bubble Tea Update() method which handles all incoming messages
// (keyboard input, window resize, custom messages) and delegates to appropriate handlers.
// Database operations and key routing are handled in separate files (operations.go, keys.go).
package table

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eduardofuncao/pam/internal/db"
)

// Update is the Bubble Tea update function that handles all incoming messages.
// It delegates keyboard input to keys.go and handles window resize events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return handleKeyPress(m, msg)
	case blinkMsg:
		m.blinkCopiedCell = false
	case clearStatusMsg:
		m.clearStatus()
		return m, nil
	case refreshTickMsg:
		return m.handleRefresh()
	case tea.WindowSizeMsg:
		return handleWindowResize(m, msg), nil
	}

	return m, nil
}

func handleWindowResize(m Model, msg tea.WindowSizeMsg) Model {
	m.width = msg.Width
	m.height = msg.Height

	availableWidth := m.width - horizontalPadding
	m.visibleCols = 0
	widthUsed := 0

	for i := m.offsetX; i < m.numCols(); i++ {
		colWidth := cellWidth
		if i < len(m.columnWidths) {
			colWidth = m.columnWidths[i]
		}

		needWidth := colWidth
		if m.visibleCols > 0 {
			needWidth += columnSeparator
		}

		if widthUsed+needWidth > availableWidth {
			break
		}

		widthUsed += needWidth
		m.visibleCols++
	}

	if m.visibleCols == 0 && m.numCols() > 0 {
		m.visibleCols = 1
	}

	m.visibleRows = m.height - verticalReserved
	if m.visibleRows > m.numRows() {
		m.visibleRows = m.numRows()
	}
	if m.visibleRows < 1 {
		m.visibleRows = 1
	}

	return m
}

func (m Model) handleRefresh() (tea.Model, tea.Cmd) {
	if !m.autoRefresh || m.queryDescriptor == nil {
		return m, nil
	}

	newTableData, err := db.ExecuteQuery(m.queryDescriptor)
	if err != nil {
		return m, tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
			return refreshTickMsg{}
		})
	}

	if newTableData != nil {
		m.tableData = newTableData
		m.columnWidths = calculateColumnWidths(newTableData)
		m.lastRefresh = time.Now()

		if m.selectedRow >= len(newTableData.Rows) {
			m.selectedRow = len(newTableData.Rows) - 1
			if m.selectedRow < 0 {
				m.selectedRow = 0
			}
		}
		if m.selectedCol >= len(newTableData.Columns) {
			m.selectedCol = len(newTableData.Columns) - 1
			if m.selectedCol < 0 {
				m.selectedCol = 0
			}
		}

		if m.offsetY > m.selectedRow {
			m.offsetY = m.selectedRow
		}
		if m.offsetX > m.selectedCol {
			m.offsetX = m.selectedCol
		}
	}

	return m, tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}
