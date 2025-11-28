package table

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	if len(m.data) < 1 {
		b.WriteString("Nothing to show here...")
	}

	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) renderHeader() string {
	var cells []string
	endCol := min(m.offsetX+m.visibleCols, m.numCols())

	for j := m.offsetX; j < endCol; j++ {
		content := formatCell(m.columns[j])
		cells = append(cells, headerStyle.Render(content))
	}

	return strings.Join(cells, borderStyle.Render("│"))
}

func (m Model) renderDataRow(rowIndex int) string {
	var cells []string
	endCol := min(m.offsetX+m.visibleCols, m.numCols())

	for j := m.offsetX; j < endCol; j++ {
		content := formatCell(m.data[rowIndex][j])
		style := m.getCellStyle(rowIndex, j)
		cells = append(cells, style.Render(content))
	}

	return strings.Join(cells, borderStyle.Render("│"))
}

func (m Model) renderFooter() string {
	updateInfo := ""
	if m.tableName != "" && m.primaryKeyCol != "" {
		updateInfo = " | Update: u"
	} else if m.tableName != "" {
		updateInfo = " | Update: u (no PK)"
	}

footer := fmt.Sprintf("\nIn %.2fs | Position: Row %d/%d, Col %d/%d | Scroll: H/L (left/right), K/J (up/down) | Copy: y/enter%s",
		m.elapsed.Seconds(), m.selectedRow+1, m.numRows(), m.selectedCol+1, m.numCols(), updateInfo)
	return lipgloss.NewStyle().Faint(true).Render(footer)
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
	if len(content) > cellWidth {
		return content[:cellWidth-1] + "…"
	}
	return fmt.Sprintf("%-*s", cellWidth, content)
}
