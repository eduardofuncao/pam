package table

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const (
	msgLoading         = "Loading..."
	msgNoData          = "Nothing to show here..."
	borderSeparator    = "│"
	truncationEllipsis = "…"
)

func (m Model) View() string {
	if m.width == 0 {
		return msgLoading
	}

	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	endRow := min(m.offsetY+m.visibleRows, m.numRows())
	for i := m.offsetY; i < endRow; i++ {
		b.WriteString(m.renderDataRow(i))
		b.WriteString("\n")
	}
	if m.tableData == nil || len(m.tableData.Rows) < 1 {
		b.WriteString(msgNoData)
	}

	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) renderHeader() string {
	var cells []string
	endCol := min(m.offsetX+m.visibleCols, m.numCols())

	for j := m.offsetX; j < endCol; j++ {
		width := cellWidth
		if j < len(m.columnWidths) {
			width = m.columnWidths[j]
		}
		content := formatCell(m.tableData.Columns[j], width)
		cells = append(cells, headerStyle.Render(content))
	}

	return strings.Join(cells, borderStyle.Render(borderSeparator))
}

func (m Model) renderDataRow(rowIndex int) string {
	var cells []string
	endCol := min(m.offsetX+m.visibleCols, m.numCols())

	for j := m.offsetX; j < endCol; j++ {
		width := cellWidth
		if j < len(m.columnWidths) {
			width = m.columnWidths[j]
		}
		content := formatCell(m.tableData.Rows[rowIndex][j].Value, width)
		style := m.getCellStyle(rowIndex, j)
		cells = append(cells, style.Render(content))
	}

	return strings.Join(cells, borderStyle.Render(borderSeparator))
}

func (m Model) renderFooter() string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorKeyHighlight)).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorNormal))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess)).Bold(true)
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorError)).Bold(true)

	edit := keyStyle.Render("e") + normalStyle.Render("dit")
	del := keyStyle.Render("d") + normalStyle.Render("el")
	yank := keyStyle.Render("y") + normalStyle.Render("ank")
	cmd := keyStyle.Render(";") + normalStyle.Render("cmd")
	quit := keyStyle.Render("q") + normalStyle.Render("uit")
	nav := normalStyle.Render("hjkl: navigate")

	cell := m.getCurrentCell()
	colType := "?"
	cellValue := ""
	if cell != nil {
		colType = cell.ColumnType
		cellValue = cell.Value
	}

	// First line: confirm prompt OR command prompt OR column type + (status message OR cell content)
	var firstLine string
	if m.confirmMode {
		promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		firstLine = fmt.Sprintf("\n%s %s",
			promptStyle.Render("Clear cell to NULL?"),
			hintStyle.Render("[Enter] confirm  [Esc] cancel"))
	} else if m.commandMode {
		firstLine = fmt.Sprintf("\n%s", m.commandInput.View())
	} else if m.statusMessage != "" {
		statusStyle := successStyle
		if m.isError {
			statusStyle = errorStyle
		}
		firstLine = fmt.Sprintf("\n%s  %s", colType, statusStyle.Render(m.statusMessage))
	} else {
		mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		firstLine = fmt.Sprintf("\n%s  %s", colType, mutedStyle.Render(cellValue))
	}

	// Second line: coordinates + commands (always shown)
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorKeyHighlight))
	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	positionStr := fmt.Sprintf("[%d,%d]", m.selectedRow+1, m.selectedCol+1)
	sizeStr := fmt.Sprintf("of %d×%d", m.numRows(), m.numCols())

	secondLine := fmt.Sprintf("%s %s | %s  %s  %s  %s  %s  %s",
		highlightStyle.Render(positionStr),
		whiteStyle.Render(sizeStr),
		edit, del, yank, cmd, quit, nav,
	)

	return firstLine + "\n" + secondLine
}

func (m Model) getCellStyle(row, col int) lipgloss.Style {
	if m.isCellInSelection(row, col) {
		if m.blinkCopiedCell {
			return copiedBlinkStyle
		}
		return selectedStyle
	}

	cell := m.getCell(row, col)
	if cell != nil && cell.Value == "NULL" {
		return nullStyle
	}

	return cellStyle
}

func formatCell(content string, cellWidth int) string {
	if cellWidth < 2 {
		return strings.Repeat(" ", cellWidth)
	}

	effectiveWidth := cellWidth - 1
	width := runewidth.StringWidth(content)

	if width > effectiveWidth {
		return runewidth.Truncate(content, effectiveWidth, truncationEllipsis) + " "
	}

	padding := effectiveWidth - width
	return content + strings.Repeat(" ", padding) + " "
}
