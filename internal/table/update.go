package table

import (
	tea "charm.land/bubbletea/v2"
	"github.com/eduardofuncao/squix/internal/config"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	case exportCompleteMsg:
		return m.handleExportComplete(msg)
	case clearExportStatusMsg:
		return m.handleClearExportStatus(), nil
	case blinkMsg:
		m.blinkCopiedCell = false
		m.blinkUpdatedCell = false
		m.blinkDeletedRow = false
		m.statusMessage = ""
	case editorCompleteMsg:
		return m.handleEditorComplete(msg)
	case deleteCompleteMsg:
		return m.handleDeleteComplete(msg)
	case queryEditCompleteMsg:
		return m.handleQueryEditComplete(msg)
	case detailViewEditCompleteMsg:
		return m.handleDetailViewEditComplete(msg)
	case detailViewEditorClosedMsg:
		return m, nil
	case saveQueryCompleteMsg:
		return m.handleSaveQueryComplete(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg), nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Handle export format selection
	if m.exportWaiting.active {
		return m.executeExportForFormat(msg.String())
	}

	// Handle search input mode
	if m.searchMode {
		return m.handleSearchInput(msg)
	}

	key := msg.String()

	// ctrl+c always quits (safety net)
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	// If in help overlay mode
	if m.helpOverlayMode {
		action, matched := m.keyMap.ResolveKey(config.ModeHelp, key)
		if matched {
			switch action {
			case config.ActionHelpClose:
				m.helpOverlayMode = false
				return m, nil
			}
		}
		return m, nil
	}

	// If in detailed view mode
	if m.detailViewMode {
		action, matched := m.keyMap.ResolveKey(config.ModeDetail, key)
		if matched {
			switch action {
			case config.ActionDetailClose:
				return m.closeDetailView(), nil
			case config.ActionDetailEdit:
				if m.tableName != "" && m.primaryKeyCol != "" {
					return m.editFromDetailView()
				}
				return m, nil
			case config.ActionDetailOpen:
				return m.openValueInEditor()
			case config.ActionDetailYank:
				return m.copySelection()
			case config.ActionDetailScrollUp:
				return m.scrollDetailViewUp(), nil
			case config.ActionDetailScrollDown:
				return m.scrollDetailViewDown(), nil
			}
		}
		return m, nil
	}

	// Normal table navigation
	action, matched := m.keyMap.ResolveKey(config.ModeNormal, key)
	if matched {
		switch action {
		case config.ActionQuit:
			return m, tea.Quit
		case config.ActionToggleFooter:
			m.uiVisibility.FooterKeymaps = !m.uiVisibility.FooterKeymaps
			return m, nil
		case config.ActionHelp:
			m.helpOverlayMode = true
			return m, nil

		case config.ActionMoveUp:
			return m.moveUp(), nil
		case config.ActionMoveDown:
			return m.moveDown(), nil
		case config.ActionMoveLeft:
			return m.moveLeft(), nil
		case config.ActionMoveRight:
			return m.moveRight(), nil
		case config.ActionJumpFirstCol:
			return m.jumpToFirstCol(), nil
		case config.ActionJumpLastCol:
			return m.jumpToLastCol(), nil
		case config.ActionJumpFirstRow:
			return m.jumpToFirstRow(), nil
		case config.ActionJumpLastRow:
			return m.jumpToLastRow(), nil
		case config.ActionPageUp:
			return m.pageUp(), nil
		case config.ActionPageDown:
			return m.pageDown(), nil

		case config.ActionVisualMode:
			return m.toggleVisualMode()
		case config.ActionVisualLineMode:
			return m.toggleVisualLineMode()

		case config.ActionYank:
			return m.copySelection()
		case config.ActionExport:
			return m.startExportFormatSelection()
		case config.ActionExportAll:
			return m.startExportAllFormatSelection()

		case config.ActionEnter:
			return m.showDetailView(), nil

		case config.ActionUpdate:
			return m.updateCell()
		case config.ActionDeleteRow:
			return m.deleteRow()
		case config.ActionEditSQL:
			return m.editAndRerunQuery()
		case config.ActionSaveQuery:
			return m.saveQuery()

		case config.ActionSearch:
			return m.startCellSearch(), nil
		case config.ActionSearchCol:
			return m.startColumnSearch(), nil
		case config.ActionNextMatch:
			return m.nextSearchMatch(), nil
		case config.ActionPrevMatch:
			return m.prevSearchMatch(), nil
		case config.ActionPrevColMatch:
			return m.prevColumnMatch(), nil
		case config.ActionNextColMatch:
			return m.nextColumnMatch(), nil
		}
	}

	return m, nil
}

func (m Model) handleWindowResize(msg tea.WindowSizeMsg) Model {
	m.width = msg.Width
	m.height = msg.Height

	m.visibleCols = (m.width - 2) / (m.cellWidth + 1)
	if m.visibleCols > m.numCols() {
		m.visibleCols = m.numCols()
	}

	// Calculate dynamic header height
	headerLines := m.calculateHeaderLines()

	// Reserve space for:  header + footer + data header row + separator
	reservedLines := headerLines + 5

	m.visibleRows = m.height - reservedLines
	if m.visibleRows > m.numRows() {
		m.visibleRows = m.numRows()
	}
	if m.visibleRows < 3 {
		m.visibleRows = 3
	}

	// Keep the detail-view scroll within bounds after a resize changes the
	// wrapped-line count or viewport height (otherwise it can sit past the new
	// maxScroll and leave the view scrolled into blank space).
	if m.detailViewMode {
		maxScroll := max(len(m.wrappedDetailLines())-m.detailViewportHeight(), 0)
		if m.detailViewScroll > maxScroll {
			m.detailViewScroll = maxScroll
		}
	}

	return m
}
