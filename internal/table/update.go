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
		m.blinkDeletedRow = false
	case editorCompleteMsg:
		return m.handleEditorComplete(msg)
	case deleteCompleteMsg:
		return m.handleDeleteComplete(msg)
	case queryEditCompleteMsg:
		return m.handleQueryEditComplete(msg)
	case detailViewEditCompleteMsg:
		return m.handleDetailViewEditComplete(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg), nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Se estiver no modo de visualização detalhada, lidar com teclas específicas
	if m.detailViewMode {
		switch msg.String() {
		case "q", "esc":
			return m.closeDetailView(), nil
		case "enter":
			// Fechar detail view e voltar para tabela
			return m.closeDetailView(), nil
		case "e":
			// Editar o conteúdo da célula
			if m.tableName != "" && m.primaryKeyCol != "" {
				return m.editFromDetailView()
			}
			return m, nil
		case "up", "k":
			return m.scrollDetailViewUp(), nil
		case "down", "j":
			return m.scrollDetailViewDown(), nil
		case "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	}

	// Navegação normal da tabela
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

	case "y":
		return m.copySelection()

	case "enter":
		// If this is a tables list, select the table
		if m.isTablesList {
			if m.selectedRow >= 0 && m.selectedRow < m.numRows() {
				// Get table name from the first column (should be "name")
				m.selectedTableName = m.data[m.selectedRow][0]
				return m, tea.Quit
			}
		}
		// Otherwise, show detail view (JSON viewer)
		return m.showDetailView(), nil

	case "u":
		return m.updateCell()
	case "D":
		return m.deleteRow()
	case "e":
		return m.editAndRerunQuery()
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

	return m
}
