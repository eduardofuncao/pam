package table

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
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
		b.WriteString("Nothing to show here...")
	}

	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) renderHeader() string {
	var cells []string
	endCol := min(m.offsetX+m.visibleCols, m.numCols())

	for j := m.offsetX; j < endCol; j++ {
		content := formatCell(m.tableData.Columns[j])
		cells = append(cells, headerStyle.Render(content))
	}

	return strings.Join(cells, borderStyle.Render("│"))
}

func (m Model) renderDataRow(rowIndex int) string {
	var cells []string
	endCol := min(m.offsetX+m.visibleCols, m.numCols())

	for j := m.offsetX; j < endCol; j++ {
		content := formatCell(m.tableData.Rows[rowIndex][j].Value)
		style := m.getCellStyle(rowIndex, j)
		cells = append(cells, style.Render(content))
	}

	return strings.Join(cells, borderStyle.Render("│"))
}

func (m Model) renderFooter() string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("40")).Bold(true)
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	edit := keyStyle.Render("e") + normalStyle.Render("dit")
	yank := keyStyle.Render("y") + normalStyle.Render("ank")
	quit := keyStyle.Render("q") + normalStyle.Render("uit")
	nav := normalStyle.Render("hjkl: navigate")

	cell := m.getCurrentCell()
	colType := "?"
	if cell != nil {
		colType = cell.ColumnType
	}

	var footer string
	if m.statusMessage != "" {
		statusStyle := successStyle
		if m.isError {
			statusStyle = errorStyle
		}
		footer = fmt.Sprintf("\n%s | %s | %s  %s  %s  %s",
			colType,
			statusStyle.Render(m.statusMessage),
			edit, yank, quit, nav,
		)
	} else {
		footer = fmt.Sprintf("\n%s | %d/%d rows, %d/%d cols | %s  %s  %s  %s",
			colType,
			m.selectedRow+1, m.numRows(),
			m.selectedCol+1, m.numCols(),
			edit, yank, quit, nav,
		)
	}

	return footer
}

func (m Model) getCellStyle(row, col int) lipgloss.Style {
	if m.isCellInSelection(row, col) {
		if m.blinkCopiedCell {
			return copiedBlinkStyle
		}
		return selectedStyle
	}
	return cellStyle
}

func formatCell(content string) string {
	width := runewidth.StringWidth(content)
	if width > cellWidth {
		return runewidth.Truncate(content, cellWidth, "…")
	}
	padding := cellWidth - width
	return content + strings.Repeat(" ", padding)
}
