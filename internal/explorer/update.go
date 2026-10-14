package explorer

import (
	tea "charm.land/bubbletea/v2"
	"github.com/eduardofuncao/squix/internal/config"
	"github.com/eduardofuncao/squix/internal/styles"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	case relationsTreeMsg:
		return m.handleRelationsTree(msg), nil
	case spinnerTickMsg:
		return m.handleSpinnerTick(), m.spinCmd()
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg), nil
	}
	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		return m, tea.Quit
	}

	if m.helpOverlay {
		action, matched := m.keyMap.ResolveKey(config.ModeHelp, key)
		if matched && action == config.ActionHelpClose {
			m.helpOverlay = false
		}
		return m, nil
	}

	if m.relationsMode {
		return m.handleRelationsKeyPress(key)
	}

	if m.searchMode {
		return m.handleSearchInput(msg)
	}

	if key == "esc" && len(m.searchMatches) > 0 {
		return m.clearSearch(), nil
	}

	action, matched := m.keyMap.ResolveKey(config.ModeExplorer, key)
	if !matched {
		return m, nil
	}

	switch action {
	case config.ActionQuit:
		return m, tea.Quit
	case config.ActionHelp:
		m.helpOverlay = true
		return m, nil
	case config.ActionToggleFooter:
		m.uiVisibility.FooterKeymaps = !m.uiVisibility.FooterKeymaps
		return m, nil

	case config.ActionMoveUp:
		return m.moveUp(), nil
	case config.ActionMoveDown:
		return m.moveDown(), nil
	case config.ActionJumpFirstRow:
		return m.jumpToFirstRow(), nil
	case config.ActionJumpLastRow:
		return m.jumpToLastRow(), nil
	case config.ActionPageUp:
		return m.pageUp(), nil
	case config.ActionPageDown:
		return m.pageDown(), nil

	case config.ActionExplorerSelect:
		return m.selectForQuery(ActionSelectAll)
	case config.ActionExplorerColumns:
		return m.selectForQuery(ActionColumns)
	case config.ActionExplorerRelations:
		return m.openRelations()
	case config.ActionExplorerToggleFold:
		return m.toggleFold(), nil

	case config.ActionSearch:
		return m.startSearch(), nil
	case config.ActionNextMatch:
		return m.nextMatch(), nil
	case config.ActionPrevMatch:
		return m.prevMatch(), nil
	}

	return m, nil
}

func (m Model) selectForQuery(action PendingAction) (tea.Model, tea.Cmd) {
	name, ok := m.selectedItemName()
	if !ok {
		m.statusMessage = styles.Faint.Render("Select a table or view first")
		return m, nil
	}
	m.pendingAction = action
	m.selectedTable = name
	return m, tea.Quit
}

func (m Model) moveUp() Model {
	if m.cursor > 0 {
		m.cursor--
		m.clampOffset()
	}
	return m
}

func (m Model) moveDown() Model {
	if m.cursor < m.numRows()-1 {
		m.cursor++
		m.clampOffset()
	}
	return m
}

func (m Model) jumpToFirstRow() Model {
	m.cursor = 0
	m.offsetY = 0
	return m
}

func (m Model) jumpToLastRow() Model {
	m.cursor = m.numRows() - 1
	m.clampOffset()
	return m
}

func (m Model) pageUp() Model {
	m.cursor -= m.visibleRows
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.clampOffset()
	return m
}

func (m Model) pageDown() Model {
	if m.numRows() == 0 {
		return m
	}
	m.cursor += m.visibleRows
	if m.cursor >= m.numRows() {
		m.cursor = m.numRows() - 1
	}
	m.clampOffset()
	return m
}

func (m Model) toggleFold() Model {
	si := m.currentSectionIdx()
	if si < 0 || si >= len(m.sections) {
		return m
	}

	prevKind := m.rows[m.cursor].kind
	prevItem := m.rows[m.cursor].item
	wasFolded := m.sections[si].folded
	m.sections[si].folded = !m.sections[si].folded
	m.rebuildRows()
	if len(m.searchMatches) > 0 {
		m.searchMatches = nil
		m.matchIdx = 0
	}

	if wasFolded {
		m.setCursorToHeader(si)
	} else if prevKind == rowItem {
		if !m.setCursorToItem(si, prevItem) {
			m.setCursorToHeader(si)
		}
	} else {
		m.setCursorToHeader(si)
	}
	return m
}

func (m *Model) setCursorToHeader(sectionIdx int) {
	for i, r := range m.rows {
		if r.kind == rowHeader && r.section == sectionIdx {
			m.cursor = i
			m.clampOffset()
			return
		}
	}
	m.clampCursor()
}

func (m *Model) setCursorToItem(sectionIdx, itemIdx int) bool {
	for i, r := range m.rows {
		if r.kind == rowItem && r.section == sectionIdx && r.item == itemIdx {
			m.cursor = i
			m.clampOffset()
			return true
		}
	}
	return false
}

func (m *Model) clampCursor() {
	if m.cursor >= m.numRows() {
		m.cursor = m.numRows() - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) clampOffset() {
	m.clampCursor()
	if m.visibleRows <= 0 {
		return
	}
	if m.cursor < m.offsetY {
		m.offsetY = m.cursor
	}
	if m.cursor >= m.offsetY+m.visibleRows {
		m.offsetY = m.cursor - m.visibleRows + 1
	}
	if m.offsetY > max(m.numRows()-m.visibleRows, 0) {
		m.offsetY = max(m.numRows()-m.visibleRows, 0)
	}
	if m.offsetY < 0 {
		m.offsetY = 0
	}
}

func (m Model) handleWindowResize(msg tea.WindowSizeMsg) Model {
	m.width = msg.Width
	m.height = msg.Height
	m.visibleRows = m.height - chromeLines()
	if m.visibleRows > m.numRows() {
		m.visibleRows = m.numRows()
	}
	if m.visibleRows < 3 {
		m.visibleRows = 3
	}
	m.clampOffset()
	return m
}
