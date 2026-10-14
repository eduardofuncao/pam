package explorer

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/eduardofuncao/squix/internal/styles"
)

func shouldUseCaseSensitive(query string) bool {
	for _, r := range query {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func containsMatch(value, query string, caseSensitive bool) bool {
	if caseSensitive {
		return strings.Contains(value, query)
	}
	return strings.Contains(strings.ToLower(value), strings.ToLower(query))
}

func (m Model) startSearch() Model {
	m.searchMode = true
	m.searchCursor = len(m.searchQuery)
	return m
}

func (m Model) handleSearchInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.executeSearch(), nil
	case "esc":
		m.searchMode = false
		if m.searchQuery == "" {
			m.clearSearch()
		}
		return m, nil
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
		if len(msg.String()) == 1 {
			before := m.searchQuery[:m.searchCursor]
			after := ""
			if m.searchCursor < len(m.searchQuery) {
				after = m.searchQuery[m.searchCursor:]
			}
			m.searchQuery = before + msg.String() + after
			m.searchCursor++
		}
	}
	return m, nil
}

func (m Model) executeSearch() Model {
	m.searchMode = false

	if m.searchQuery == "" {
		m.clearSearch()
		return m
	}

	caseSensitive := shouldUseCaseSensitive(m.searchQuery)
	m.searchMatches = m.searchMatches[:0]
	for i, r := range m.rows {
		if r.kind != rowItem {
			continue
		}
		item := m.sections[r.section].items[r.item]
		if containsMatch(item, m.searchQuery, caseSensitive) {
			m.searchMatches = append(m.searchMatches, i)
		}
	}

	switch count := len(m.searchMatches); count {
	case 0:
		m.matchIdx = 0
		m.statusMessage = styles.Error.Render(
			fmt.Sprintf("No matches for %q", m.searchQuery),
		)
	default:
		m.jumpToFirstMatchFrom(m.cursor)
		m.statusMessage = styles.Faint.Render(
			fmt.Sprintf("Found %d matches for %q", count, m.searchQuery),
		)
	}
	return m
}

func (m *Model) jumpToFirstMatchFrom(from int) {
	for idx, rowIdx := range m.searchMatches {
		if rowIdx >= from {
			m.matchIdx = idx
			m.cursor = rowIdx
			m.clampOffset()
			return
		}
	}
	m.matchIdx = 0
	m.cursor = m.searchMatches[0]
	m.clampOffset()
}

func (m Model) clearSearch() Model {
	m.searchMode = false
	m.searchQuery = ""
	m.searchCursor = 0
	m.searchMatches = nil
	m.matchIdx = 0
	m.statusMessage = ""
	return m
}

func (m Model) nextMatch() Model {
	if len(m.searchMatches) == 0 {
		return m
	}
	m.matchIdx = (m.matchIdx + 1) % len(m.searchMatches)
	m.cursor = m.searchMatches[m.matchIdx]
	m.clampOffset()
	return m
}

func (m Model) prevMatch() Model {
	if len(m.searchMatches) == 0 {
		return m
	}
	m.matchIdx = (m.matchIdx - 1 + len(m.searchMatches)) % len(m.searchMatches)
	m.cursor = m.searchMatches[m.matchIdx]
	m.clampOffset()
	return m
}
