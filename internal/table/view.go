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

type shortcut struct {
	key   string
	label string
}

func renderShortcuts(shortcuts []shortcut) string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorKeyHighlight)).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorNormal))

	var parts []string
	for _, s := range shortcuts {
		parts = append(parts, keyStyle.Render(s.key)+normalStyle.Render(s.label))
	}
	return strings.Join(parts, "  ")
}

var (
	shortcutsNormal = []shortcut{
		{"e", "dit"}, {"d", "el"}, {"y", "ank"}, {";", "cmd"}, {"q", "uit"},
	}
	shortcutsDelete = []shortcut{
		{"c", "ell"}, {"r", "ow"}, {"esc", " cancel"},
	}
)

func (m Model) renderFooter() string {
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess)).Bold(true)
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorError)).Bold(true)

	cell := m.getCurrentCell()
	colType := "?"
	cellValue := ""
	if cell != nil {
		colType = cell.ColumnType
		cellValue = cell.Value
	}

	// First line: delete mode OR confirm prompt OR command prompt OR column type + (status message OR cell content)
	var firstLine string
	if m.deleteMode {
		dangerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorError)).Bold(true)
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

		target := "cell"
		if m.deleteTarget == "row" {
			target = "row"
		}

		firstLine = fmt.Sprintf("\n%s %s",
			dangerStyle.Render(fmt.Sprintf("DELETE %s?", target)),
			hintStyle.Render("[Enter] confirm  [Esc] cancel"))
	} else if m.confirmMode {
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

	// Second line: coordinates + shortcuts (changes based on mode)
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorKeyHighlight))
	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorNormal))

	positionStr := fmt.Sprintf("[%d,%d]", m.selectedRow+1, m.selectedCol+1)
	sizeStr := fmt.Sprintf("of %d×%d", m.numRows(), m.numCols())

	shortcuts := shortcutsNormal
	if m.deleteMode {
		shortcuts = shortcutsDelete
	}

	secondLine := fmt.Sprintf("%s %s | %s  %s",
		highlightStyle.Render(positionStr),
		whiteStyle.Render(sizeStr),
		renderShortcuts(shortcuts),
		normalStyle.Render("hjkl: navigate"),
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
	if cellWidth < 1 {
		return ""
	}

	contentWidth := runewidth.StringWidth(content)

	if contentWidth > cellWidth {
		return runewidth.Truncate(content, cellWidth-1, truncationEllipsis) + " "
	}

	padding := cellWidth - contentWidth
	return content + strings.Repeat(" ", padding)
}
