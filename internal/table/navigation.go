package table

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
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

	var allRows [][]string

	if m.visualMode {
		headerRow := make([]string, 0)
		for col := minCol; col <= maxCol; col++ {
			headerRow = append(headerRow, m.columns[col])
		}
		allRows = append(allRows, headerRow)
	}

	for row := minRow; row <= maxRow; row++ {
		dataRow := make([]string, 0)
		for col := minCol; col <= maxCol; col++ {
			dataRow = append(dataRow, m.data[row][col])
		}
		allRows = append(allRows, dataRow)
	}

	numCols := maxCol - minCol + 1
	colWidths := make([]int, numCols)

	for _, row := range allRows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var result strings.Builder

	for rowIdx, row := range allRows {
		for colIdx, cell := range row {
			paddedCell := fmt.Sprintf("%-*s", colWidths[colIdx], cell)
			result.WriteString(paddedCell)

			if colIdx < len(row)-1 {
				result.WriteString("  ")
			}
		}

		if rowIdx < len(allRows)-1 {
			result.WriteString("\n")
		}
	}

	content := result.String()
	clipboard.WriteAll(content)

	m.visualMode = false
	m.blinkCopiedCell = true

	return m, func() tea.Msg {
		time.Sleep(200 * time.Millisecond)
		return blinkMsg{}
	}
}

func (m Model) showDetailView() Model {
	if m.selectedRow < 0 || m.selectedRow >= len(m.data) ||
		m.selectedCol < 0 || m.selectedCol >= len(m.data[m.selectedRow]) {
		return m
	}

	cellValue := m.data[m.selectedRow][m.selectedCol]

	// Tentar formatar como JSON
	formattedValue := formatValueIfJSON(cellValue)

	m.detailViewMode = true
	m.detailViewContent = formattedValue
	m.detailViewScroll = 0

	return m
}

func (m Model) editFromDetailView() (Model, tea.Cmd) {
	if m.selectedRow < 0 || m.selectedRow >= len(m.data) ||
		m.selectedCol < 0 || m.selectedCol >= len(m.data[m.selectedRow]) {
		return m, nil
	}

	// Verificar se pode editar (precisa de tableName e primaryKey)
	if m.tableName == "" {
		return m, nil
	}

	// Construir UPDATE statement com o valor atual (formatado se for JSON)
	columnName := m.columns[m.selectedCol]
	currentValue := m.data[m.selectedRow][m.selectedCol]

	// If the content is formatted (JSON), use the formatted value
	if m.detailViewContent != currentValue {
		// It's formatted, use the formatted content
		currentValue = m.detailViewContent
	}

	pkValue := ""
	if m.primaryKeyCol != "" {
		for i, col := range m.columns {
			if col == m.primaryKeyCol {
				pkValue = m.data[m.selectedRow][i]
				break
			}
		}
	}

	updateStmt := m.dbConnection.BuildUpdateStatement(
		m.tableName,
		columnName,
		currentValue,
		m.primaryKeyCol,
		pkValue,
	)

	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
	}

	tmpFile, err := os.CreateTemp("", "pam-update-*.sql")
	if err != nil {
		return m, nil
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write([]byte(updateStmt)); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return m, nil
	}
	tmpFile.Close()

	cmd := buildEditorCommandForDetailView(editorCmd, tmpPath, updateStmt)

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		editedSQL, readErr := os.ReadFile(tmpPath)
		os.Remove(tmpPath)

		if err != nil || readErr != nil {
			return nil
		}

		return detailViewEditCompleteMsg{
			sql:      string(editedSQL),
			colIndex: m.selectedCol,
		}
	})
}

type detailViewEditCompleteMsg struct {
	sql      string
	colIndex int
}

func (m Model) handleDetailViewEditComplete(
	msg detailViewEditCompleteMsg,
) (tea.Model, tea.Cmd) {
	// Validar o UPDATE statement
	if err := validateUpdateStatement(msg.sql); err != nil {
		printError("Update validation failed: %v", err)
		m.detailViewMode = false
		return m, nil
	}

	// Extrair o novo valor do SQL
	newValue := m.extractNewValue(msg.sql, m.columns[msg.colIndex])

	// Executar update
	if err := m.executeUpdate(msg.sql); err != nil {
		printError("Could not execute update: %v", err)
		m.detailViewMode = false
		return m, nil
	}

	// Atualizar dados locais
	m.data[m.selectedRow][m.selectedCol] = newValue

	// Close detail view and return to table with highlighted cell
	m.detailViewMode = false
	m.blinkUpdatedCell = true
	m.updatedRow = m.selectedRow
	m.updatedCol = m.selectedCol

	return m, tea.Batch(
		tea.ClearScreen,
		m.blinkCmd(),
	)
}

func (m Model) closeDetailView() Model {
	m.detailViewMode = false
	m.detailViewContent = ""
	m.detailViewScroll = 0
	return m
}

func (m Model) scrollDetailViewUp() Model {
	if m.detailViewScroll > 0 {
		m.detailViewScroll--
	}
	return m
}

func (m Model) scrollDetailViewDown() Model {
	lines := strings.Count(m.detailViewContent, "\n") + 1
	maxScroll := lines - (m.height - 10) // Reserve space for header and footer
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.detailViewScroll < maxScroll {
		m.detailViewScroll++
	}
	return m
}

func formatValueIfJSON(value string) string {
	trimmed := strings.TrimSpace(value)

	// Verificar se parece ser JSON
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return value
	}

	// Tentar fazer parse do JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(trimmed), &jsonData); err != nil {
		// Not valid JSON, return original value
		return value
	}

	// Format JSON with indentation
	formatted, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return value
	}

	return string(formatted)
}

func buildEditorCommandForDetailView(
	editorCmd, tmpPath, updateStmt string,
) *exec.Cmd {
	// Use the same logic as buildEditorCommand for positioning cursor
	return buildEditorCommand(
		editorCmd,
		tmpPath,
		updateStmt,
		CursorAtUpdateValue,
	)
}
