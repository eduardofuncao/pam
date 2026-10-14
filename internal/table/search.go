package table

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// shouldUseCaseSensitive returns true if query contains uppercase (smart case)
func shouldUseCaseSensitive(query string) bool {
	for _, r := range query {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

// cellMatchProgress and colMatchProgress returns the 1-based
// position of the match the cursor last landed on
func (m Model) cellMatchProgress() (current, total int) {
	total = len(m.searchMatches)
	if total == 0 {
		return 0, 0
	}
	return m.cellMatchIdx + 1, total
}

func (m Model) colMatchProgress() (current, total int) {
	total = len(m.searchColMatches)
	if total == 0 {
		return 0, 0
	}
	return m.colMatchIdx + 1, total
}

func (m Model) startCellSearch() Model {
	m.searchMode = true
	m.columnSearchMode = false
	m.searchQuery = ""
	m.searchMatches = []CellPosition{}
	m.cellMatchIdx = 0
	m.searchCursor = 0
	return m
}

func (m Model) startColumnSearch() Model {
	m.searchMode = true
	m.columnSearchMode = true
	m.searchQuery = ""
	m.searchColMatches = []int{}
	m.colMatchIdx = 0
	m.searchCursor = 0
	return m
}

// handleSearchInput processes keystrokes during search input
func (m Model) handleSearchInput(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.executeSearch()
	case "esc":
		return m.clearSearch(), nil
	case "backspace":
		if m.searchCursor > 0 && len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:m.searchCursor-1] + m.searchQuery[m.searchCursor:]
			m.searchCursor--
		}
	case "left":
		if m.searchCursor > 0 {
			m.searchCursor--
		}
	case "right":
		if m.searchCursor < len(m.searchQuery) {
			m.searchCursor++
		}
	default:
		if t := msg.Key().Text; t != "" {
			before := m.searchQuery[:m.searchCursor]
			after := ""
			if m.searchCursor < len(m.searchQuery) {
				after = m.searchQuery[m.searchCursor:]
			}
			m.searchQuery = before + t + after
			m.searchCursor += len(t)
		}
	}
	return m, nil
}

// executeSearch performs the actual search and shows status
func (m Model) executeSearch() (Model, tea.Cmd) {
	m.searchMode = false

	if m.searchQuery == "" {
		return m, nil
	}

	var result Model
	var matchCount int
	var searchType string

	if m.columnSearchMode {
		result = m.searchColumnHeaders()
		matchCount = len(result.searchColMatches)
		searchType = "columns"
	} else {
		result = m.searchCells()
		matchCount = len(result.searchMatches)
		searchType = "cells"
	}

	if matchCount > 0 {
		result.statusMessage = fmt.Sprintf("Found %d %s", matchCount, searchType)
	} else {
		result.statusMessage = fmt.Sprintf("No %s found", searchType)
	}

	return result, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return blinkMsg{}
	})
}

func (m Model) searchCells() Model {
	m.searchMatches = []CellPosition{}

	caseSensitive := shouldUseCaseSensitive(m.searchQuery)

	for row := 0; row < m.numRows(); row++ {
		for col := 0; col < m.numCols(); col++ {
			cellValue := m.data[row][col]

			match := false
			if caseSensitive {
				match = strings.Contains(cellValue, m.searchQuery)
			} else {
				match = strings.Contains(strings.ToLower(cellValue), strings.ToLower(m.searchQuery))
			}

			if match {
				m.searchMatches = append(m.searchMatches, CellPosition{Row: row, Col: col})
			}
		}
	}

	if len(m.searchMatches) > 0 {
		found := false
		for i, match := range m.searchMatches {
			if match.Row > m.selectedRow || (match.Row == m.selectedRow && match.Col > m.selectedCol) {
				m.selectedRow = match.Row
				m.selectedCol = match.Col
				m.cellMatchIdx = i
				found = true
				break
			}
		}

		if !found {
			firstMatch := m.searchMatches[0]
			m.selectedRow = firstMatch.Row
			m.selectedCol = firstMatch.Col
			m.cellMatchIdx = 0
		}

		if m.selectedRow < m.offsetY {
			m.offsetY = m.selectedRow
		}
		if m.selectedRow >= m.offsetY+m.visibleRows {
			m.offsetY = m.selectedRow - m.visibleRows + 1
		}
		if m.selectedCol < m.offsetX {
			m.offsetX = m.selectedCol
		}
		if m.selectedCol >= m.offsetX+m.visibleCols {
			m.offsetX = m.selectedCol - m.visibleCols + 1
		}
	}

	return m
}

func (m Model) searchColumnHeaders() Model {
	m.searchColMatches = []int{}

	caseSensitive := shouldUseCaseSensitive(m.searchQuery)

	for col := 0; col < m.numCols(); col++ {
		match := false
		if caseSensitive {
			match = strings.Contains(m.columns[col], m.searchQuery)
		} else {
			match = strings.Contains(strings.ToLower(m.columns[col]), strings.ToLower(m.searchQuery))
		}

		if match {
			m.searchColMatches = append(m.searchColMatches, col)
		}
	}

	if len(m.searchColMatches) > 0 {
		found := false
		for i, matchCol := range m.searchColMatches {
			if matchCol > m.selectedCol {
				m.selectedCol = matchCol
				m.colMatchIdx = i
				found = true
				break
			}
		}

		if !found {
			m.selectedCol = m.searchColMatches[0]
			m.colMatchIdx = 0
		}

		if m.selectedCol < m.offsetX {
			m.offsetX = m.selectedCol
		}
		if m.selectedCol >= m.offsetX+m.visibleCols {
			m.offsetX = m.selectedCol - m.visibleCols + 1
		}
	}

	return m
}

func (m Model) clearSearch() Model {
	m.searchMode = false
	m.columnSearchMode = false
	m.searchQuery = ""
	m.searchMatches = []CellPosition{}
	m.searchColMatches = []int{}
	m.cellMatchIdx = 0
	m.colMatchIdx = 0
	return m
}

func (m Model) nextSearchMatch() Model {
	if len(m.searchMatches) == 0 {
		return m
	}

	for i, match := range m.searchMatches {
		if match.Row > m.selectedRow || (match.Row == m.selectedRow && match.Col > m.selectedCol) {
			m.selectedRow = match.Row
			m.selectedCol = match.Col
			m.cellMatchIdx = i

			if m.selectedRow < m.offsetY {
				m.offsetY = m.selectedRow
			}
			if m.selectedRow >= m.offsetY+m.visibleRows {
				m.offsetY = m.selectedRow - m.visibleRows + 1
			}
			if m.selectedCol < m.offsetX {
				m.offsetX = m.selectedCol
			}
			if m.selectedCol >= m.offsetX+m.visibleCols {
				m.offsetX = m.selectedCol - m.visibleCols + 1
			}
			return m
		}
	}

	firstMatch := m.searchMatches[0]
	m.selectedRow = firstMatch.Row
	m.selectedCol = firstMatch.Col
	m.cellMatchIdx = 0

	if m.selectedRow < m.offsetY {
		m.offsetY = m.selectedRow
	}
	if m.selectedRow >= m.offsetY+m.visibleRows {
		m.offsetY = m.selectedRow - m.visibleRows + 1
	}
	if m.selectedCol < m.offsetX {
		m.offsetX = m.selectedCol
	}
	if m.selectedCol >= m.offsetX+m.visibleCols {
		m.offsetX = m.selectedCol - m.visibleCols + 1
	}

	return m
}

func (m Model) prevSearchMatch() Model {
	if len(m.searchMatches) == 0 {
		return m
	}

	for i := len(m.searchMatches) - 1; i >= 0; i-- {
		match := m.searchMatches[i]
		if match.Row < m.selectedRow || (match.Row == m.selectedRow && match.Col < m.selectedCol) {
			m.selectedRow = match.Row
			m.selectedCol = match.Col
			m.cellMatchIdx = i

			if m.selectedRow < m.offsetY {
				m.offsetY = m.selectedRow
			}
			if m.selectedRow >= m.offsetY+m.visibleRows {
				m.offsetY = m.selectedRow - m.visibleRows + 1
			}
			if m.selectedCol < m.offsetX {
				m.offsetX = m.selectedCol
			}
			if m.selectedCol >= m.offsetX+m.visibleCols {
				m.offsetX = m.selectedCol - m.visibleCols + 1
			}
			return m
		}
	}

	lastMatch := m.searchMatches[len(m.searchMatches)-1]
	m.selectedRow = lastMatch.Row
	m.selectedCol = lastMatch.Col
	m.cellMatchIdx = len(m.searchMatches) - 1

	if m.selectedRow < m.offsetY {
		m.offsetY = m.selectedRow
	}
	if m.selectedRow >= m.offsetY+m.visibleRows {
		m.offsetY = m.selectedRow - m.visibleRows + 1
	}
	if m.selectedCol < m.offsetX {
		m.offsetX = m.selectedCol
	}
	if m.selectedCol >= m.offsetX+m.visibleCols {
		m.offsetX = m.selectedCol - m.visibleCols + 1
	}

	return m
}

func (m Model) nextColumnMatch() Model {
	if len(m.searchColMatches) == 0 {
		return m
	}

	for i, matchCol := range m.searchColMatches {
		if matchCol > m.selectedCol {
			m.selectedCol = matchCol
			m.colMatchIdx = i

			if m.selectedCol < m.offsetX {
				m.offsetX = m.selectedCol
			}
			if m.selectedCol >= m.offsetX+m.visibleCols {
				m.offsetX = m.selectedCol - m.visibleCols + 1
			}
			return m
		}
	}

	m.selectedCol = m.searchColMatches[0]
	m.colMatchIdx = 0

	if m.selectedCol < m.offsetX {
		m.offsetX = m.selectedCol
	}
	if m.selectedCol >= m.offsetX+m.visibleCols {
		m.offsetX = m.selectedCol - m.visibleCols + 1
	}

	return m
}

func (m Model) prevColumnMatch() Model {
	if len(m.searchColMatches) == 0 {
		return m
	}

	for i := len(m.searchColMatches) - 1; i >= 0; i-- {
		matchCol := m.searchColMatches[i]
		if matchCol < m.selectedCol {
			m.selectedCol = matchCol
			m.colMatchIdx = i

			if m.selectedCol < m.offsetX {
				m.offsetX = m.selectedCol
			}
			if m.selectedCol >= m.offsetX+m.visibleCols {
				m.offsetX = m.selectedCol - m.visibleCols + 1
			}
			return m
		}
	}

	m.selectedCol = m.searchColMatches[len(m.searchColMatches)-1]
	m.colMatchIdx = len(m.searchColMatches) - 1

	if m.selectedCol < m.offsetX {
		m.offsetX = m.selectedCol
	}
	if m.selectedCol >= m.offsetX+m.visibleCols {
		m.offsetX = m.selectedCol - m.visibleCols + 1
	}

	return m
}
