package table

import (
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	blinkDuration = 200 * time.Millisecond
)

func (m Model) moveUp() Model {
	if m.selectedRow > 0 {
		m.selectedRow--
		if m.selectedRow < m.offsetY {
			m.offsetY = m.selectedRow
		}
	}
	return m
}

func (m Model) moveDown() Model {
	if m.selectedRow < m.numRows()-1 {
		m.selectedRow++
		if m.selectedRow >= m.offsetY+m.visibleRows {
			m.offsetY = m.selectedRow - m.visibleRows + 1
		}
	}
	return m
}

func (m Model) moveLeft() Model {
	if m.selectedCol > 0 {
		m.selectedCol--
		if m.selectedCol < m.offsetX {
			m.offsetX = m.selectedCol
		}
	}
	return m
}

func (m Model) moveRight() Model {
	if m.selectedCol < m.numCols()-1 {
		m.selectedCol++
		if m.selectedCol >= m.offsetX+m.visibleCols {
			m.offsetX = m.selectedCol - m.visibleCols + 1
		}
	}
	return m
}

func (m Model) jumpToFirstCol() Model {
	m.selectedCol = 0
	m.offsetX = 0
	return m
}

func (m Model) jumpToLastCol() Model {
	m.selectedCol = m.numCols() - 1
	if m.visibleCols < m.numCols() {
		m.offsetX = m.numCols() - m.visibleCols
	}
	return m
}

func (m Model) jumpToFirstRow() Model {
	m.selectedRow = 0
	m.offsetY = 0
	return m
}

func (m Model) jumpToLastRow() Model {
	m.selectedRow = m.numRows() - 1
	m.offsetY = m.numRows() - m.visibleRows
	return m
}

func (m Model) pageUp() Model {
	m.selectedRow -= m.visibleRows
	if m.selectedRow < 0 {
		m.selectedRow = 0
	}
	m.offsetY = m.selectedRow
	return m
}

func (m Model) pageDown() Model {
	m.selectedRow += m.visibleRows
	if m.selectedRow >= m.numRows() {
		m.selectedRow = m.numRows() - 1
	}
	if m.selectedRow >= m.offsetY+m.visibleRows {
		m.offsetY = m.selectedRow - m.visibleRows + 1
	}
	return m
}

// func (m Model) copySelectedCell() (Model, tea.Cmd) {
// 	if m.selectedRow >= 0 && m.selectedRow < m.numRows() &&
// 		m.selectedCol >= 0 && m.selectedCol < m.numCols() {
// 		go clipboard.WriteAll(m.data[m.selectedRow][m.selectedCol])
// 		m.blinkCopiedCell = true
// 		return m, tea.Tick(time.Millisecond*400, func(time.Time) tea.Msg {
// 			return blinkMsg{}
// 		})
// 	}
// 	return m, nil
// }

func (m Model) toggleVisualMode() (Model, tea.Cmd) {
	m.visualMode = !m.visualMode
	
	if m.visualMode {
		m.visualStartRow = m.selectedRow
		m.visualStartCol = m.selectedCol
	}
	
	return m, nil
}

func (m Model) getSelectionBounds() (minRow, maxRow, minCol, maxCol int) {
	if !m.visualMode {
		return m.selectedRow, m.selectedRow, m.selectedCol, m.selectedCol
	}
	
	// Multi-cell selection
	minRow = min(m.visualStartRow, m.selectedRow)
	maxRow = max(m.visualStartRow, m.selectedRow)
	minCol = min(m.visualStartCol, m.selectedCol)
	maxCol = max(m.visualStartCol, m.selectedCol)
	
	return
}

func (m Model) isCellInSelection(row, col int) bool {
	minRow, maxRow, minCol, maxCol := m.getSelectionBounds()
	return row >= minRow && row <= maxRow && col >= minCol && col <= maxCol
}

func (m Model) copySelection() (Model, tea.Cmd) {
	minRow, maxRow, minCol, maxCol := m.getSelectionBounds()
	
	var result strings.Builder
	
	if m.visualMode {
		for col := minCol; col <= maxCol; col++ {
			if col > minCol {
				result.WriteString("\t")
			}
			result.WriteString(m.tableData.Columns[col])
		}
		result.WriteString("\n")
	}

	for row := minRow; row <= maxRow; row++ {
		for col := minCol; col <= maxCol; col++ {
			if col > minCol {
				result.WriteString("\t")
			}
			result.WriteString(m.tableData.Rows[row][col].Value)
		}
		if row < maxRow {
			result.WriteString("\n")
		}
	}
	
	content := result.String()
	clipboard.WriteAll(content)

	m.visualMode = false
	m.blinkCopiedCell = true

	return m, tea.Batch(
		m.setSuccess("Copied to clipboard!"),
		func() tea.Msg {
			time.Sleep(blinkDuration)
			return blinkMsg{}
		},
	)
}
