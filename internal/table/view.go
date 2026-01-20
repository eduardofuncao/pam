package table

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/eduardofuncao/pam/internal/parser"
	"github.com/eduardofuncao/pam/internal/styles"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Se estiver no modo de visualização detalhada, mostrar a view detalhada
	if m.detailViewMode {
		return m.renderDetailView()
	}

	var b strings.Builder

	// Display query name header
	b.WriteString(styles.Title.Render("◆ " + m.currentQuery.Name))
	b.WriteString("\n")

	// Show the last executed query (for updates) or the current query (for selects)
	var queryToDisplay string
	if m.lastExecutedQuery != "" {
		queryToDisplay = m.lastExecutedQuery
	} else {
		queryToDisplay = m.currentQuery.SQL
	}

	// Format and highlight the SQL
	formattedSQL := parser.FormatSQLWithLineBreaks(queryToDisplay)
	highlightedSQL := parser.HighlightSQL(formattedSQL)
	b.WriteString(highlightedSQL)
	b.WriteString("\n")

	// Add separator line
	separatorWidth := 0
	endCol := min(m.offsetX+m.visibleCols, m.numCols())
	for j := m.offsetX; j < endCol; j++ {
		separatorWidth += m.cellWidth
		if j < endCol-1 {
			separatorWidth += 1
		}
	}

	b.WriteString(styles.Separator.Render(strings.Repeat("─", separatorWidth)))
	b.WriteString("\n")

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
		typeIcon := ""
		if j < len(m.columnTypes) && m.columnTypes[j] != "" {
			typeIcon = getTypeIcon(m.columnTypes[j]) + " "
		}

		pkIcon := ""
		if m.primaryKeyCol != "" && j < len(m.columns) &&
			m.columns[j] == m.primaryKeyCol {
			pkIcon = "⚿ "
		}

		columnDisplay := pkIcon + typeIcon + m.columns[j]
		content := formatCell(columnDisplay, m.cellWidth)
		cells = append(cells, styles.TableHeader.Render(content))
	}

	return strings.Join(cells, styles.TableBorder.Render("│"))
}

func (m Model) renderDataRow(rowIndex int) string {
	var cells []string
	endCol := min(m.offsetX+m.visibleCols, m.numCols())

	for j := m.offsetX; j < endCol; j++ {
		content := formatCell(m.data[rowIndex][j], m.cellWidth)
		style := m.getCellStyle(rowIndex, j)
		cells = append(cells, style.Render(content))
	}

	return strings.Join(cells, styles.TableBorder.Render("│"))
}

func (m Model) renderFooter() string {
	currentCellValue := ""
	columnType := ""

	if m.selectedRow >= 0 && m.selectedRow < len(m.data) &&
		m.selectedCol >= 0 && m.selectedCol < len(m.data[m.selectedRow]) {
		currentCellValue = m.data[m.selectedRow][m.selectedCol]
	}

	if m.selectedCol >= 0 && m.selectedCol < len(m.columnTypes) {
		columnType = m.columnTypes[m.selectedCol]
	}

	maxPreviewWidth := m.width - len(columnType) - 10
	displayValue := currentCellValue
	if len(displayValue) > maxPreviewWidth && maxPreviewWidth > 0 {
		displayValue = displayValue[:maxPreviewWidth-3] + "..."
	}

	cellPreview := fmt.Sprintf("%s %s\n",
		styles.Faint.Render(columnType),
		styles.TableCell.Render(displayValue))

	updateInfo := ""
	delInfo := ""
	enterInfo := ""

	if m.isTablesList {
		// Special footer for tables list
		enterInfo = styles.TableHeader.Render(
			"↵",
		) + styles.Faint.Render(
			"enter",
		)
		updateInfo = ""
		delInfo = ""
	} else if m.tableName != "" && m.primaryKeyCol != "" {
		updateInfo = styles.TableHeader.Render(
			"u",
		) + styles.Faint.Render(
			"pdate",
		)
		delInfo = styles.TableHeader.Render("D") + styles.Faint.Render("el")
	} else if m.tableName != "" {
		updateInfo = styles.TableHeader.Render(
			"u",
		) + styles.Faint.Render(
			"pdate (no PK)",
		)
		delInfo = ""
	} else {
		// No table name means JOIN or complex query
		updateInfo = styles.Faint.Render("(update/delete disabled)")
		delInfo = ""
	}

	sel := styles.TableHeader.Render("v") + styles.Faint.Render("sel")
	edit := styles.TableHeader.Render("e") + styles.Faint.Render("ditSQL")
	yank := styles.TableHeader.Render("y") + styles.Faint.Render("ank")
	quit := styles.TableHeader.Render("q") + styles.Faint.Render("uit")
	hjkl := styles.TableHeader.Render("hjkl") + styles.Faint.Render("←↓↑→")

	var footer string
	if m.isTablesList {
		footer = fmt.Sprintf(
			"\n%s%s %s | %s | %s  %s  %s  %s  %s",
			cellPreview,
			styles.Faint.Render(fmt.Sprintf("%dx%d", m.numRows(), m.numCols())),
			styles.Faint.Render(fmt.Sprintf("In %.2fs", m.elapsed.Seconds())),
			styles.Faint.Render(
				fmt.Sprintf("[%d/%d]", m.selectedRow+1, m.selectedCol+1),
			),
			enterInfo,
			yank,
			edit,
			quit,
			hjkl,
		)
	} else {
		footer = fmt.Sprintf(
			"\n%s%s %s | %s | %s  %s  %s  %s  %s  %s  %s",
			cellPreview,
			styles.Faint.Render(fmt.Sprintf("%dx%d", m.numRows(), m.numCols())),
			styles.Faint.Render(fmt.Sprintf("In %.2fs", m.elapsed.Seconds())),
			styles.Faint.Render(
				fmt.Sprintf("[%d/%d]", m.selectedRow+1, m.selectedCol+1),
			),
			updateInfo,
			delInfo,
			yank,
			sel,
			edit,
			quit,
			hjkl,
		)
	}

	return footer
}

func (m Model) getCellStyle(row, col int) lipgloss.Style {
	if m.blinkDeletedRow && m.deletedRow == row {
		return styles.TableDeleted
	}

	if m.blinkUpdatedCell && m.updatedRow == row && m.updatedCol == col {
		return styles.TableUpdated
	}

	if m.isCellInSelection(row, col) {
		if m.blinkCopiedCell {
			return styles.TableCopiedBlink
		}
		return styles.TableSelected
	}

	return styles.TableCell
}

func formatCell(content string, cellWidth int) string {
	runes := []rune(content)
	runeCount := len(runes)

	if runeCount > cellWidth {
		return string(runes[:cellWidth-1]) + "…"
	}

	padding := cellWidth - runeCount
	return content + strings.Repeat(" ", padding)
}

func getTypeIcon(typeName string) string {
	upper := strings.ToUpper(typeName)

	// String/Text types
	if strings.Contains(upper, "CHAR") || strings.Contains(upper, "TEXT") ||
		strings.Contains(upper, "STRING") || strings.Contains(upper, "CLOB") ||
		strings.Contains(
			upper,
			"VARCHAR",
		) || strings.Contains(upper, "NVARCHAR") {
		return "α"
	}

	// Integer types
	if strings.Contains(upper, "INT") || strings.Contains(upper, "SERIAL") ||
		strings.Contains(
			upper,
			"BIGINT",
		) || strings.Contains(upper, "SMALLINT") ||
		strings.Contains(upper, "TINYINT") {
		return "№"
	}

	// Decimal/Float types
	if strings.Contains(upper, "DECIMAL") || strings.Contains(upper, "NUMERIC") ||
		strings.Contains(upper, "FLOAT") || strings.Contains(upper, "DOUBLE") ||
		strings.Contains(upper, "REAL") ||
		strings.Contains(upper, "NUMBER") ||
		strings.Contains(upper, "MONEY") {
		return "≈"
	}

	// Date types
	if strings.Contains(upper, "DATE") && !strings.Contains(upper, "TIME") {
		return "⊞"
	}

	// Time/Timestamp types
	if strings.Contains(upper, "TIME") || strings.Contains(upper, "TIMESTAMP") {
		return "◷"
	}

	// Boolean types
	if strings.Contains(upper, "BOOL") || strings.Contains(upper, "BIT") {
		return "✓"
	}

	// Binary/Blob types
	if strings.Contains(upper, "BLOB") || strings.Contains(upper, "BINARY") ||
		strings.Contains(upper, "BYTEA") || strings.Contains(upper, "RAW") ||
		strings.Contains(
			upper,
			"VARBINARY",
		) || strings.Contains(upper, "IMAGE") {
		return "◆"
	}

	// JSON types
	if strings.Contains(upper, "JSON") || strings.Contains(upper, "JSONB") {
		return "{ }"
	}

	// UUID types
	if strings.Contains(upper, "UUID") || strings.Contains(upper, "GUID") {
		return "I"
	}

	// Array types
	if strings.Contains(upper, "ARRAY") || strings.HasSuffix(upper, "[]") {
		return "≡"
	}

	// Enum types
	if strings.Contains(upper, "ENUM") || strings.Contains(upper, "SET") {
		return "⋮"
	}

	// XML types
	if strings.Contains(upper, "XML") {
		return "⟨⟩"
	}

	// Geometric/Spatial types
	if strings.Contains(upper, "GEOMETRY") || strings.Contains(upper, "POINT") ||
		strings.Contains(upper, "POLYGON") ||
		strings.Contains(upper, "LINE") {
		return "◉"
	}

	// Default fallback
	return "•"
}

func (m Model) renderDetailView() string {
	var b strings.Builder

	// Obter informações da célula selecionada
	columnName := ""
	columnType := ""
	if m.selectedCol >= 0 && m.selectedCol < len(m.columns) {
		columnName = m.columns[m.selectedCol]
	}
	if m.selectedCol >= 0 && m.selectedCol < len(m.columnTypes) {
		columnType = m.columnTypes[m.selectedCol]
	}

	// Header
	titleLine := fmt.Sprintf("◆ Cell Value - %s", columnName)
	if columnType != "" {
		titleLine += fmt.Sprintf(" (%s)", columnType)
	}
	b.WriteString(styles.Title.Render(titleLine))
	b.WriteString("\n")

	// Informação da posição
	posInfo := fmt.Sprintf(
		"Row %d, Column %d",
		m.selectedRow+1,
		m.selectedCol+1,
	)
	b.WriteString(styles.Faint.Render(posInfo))

	// Mostrar se pode editar
	if m.tableName != "" && m.primaryKeyCol != "" {
		b.WriteString(" ")
		b.WriteString(styles.Faint.Render("• Press 'e' to edit"))
	}

	b.WriteString("\n\n")

	// Separator
	separatorWidth := m.width - 4
	if separatorWidth < 0 {
		separatorWidth = 0
	}
	b.WriteString(styles.Separator.Render(strings.Repeat("─", separatorWidth)))
	b.WriteString("\n\n")

	// Conteúdo com scroll
	lines := strings.Split(m.detailViewContent, "\n")
	availableHeight := m.height - 10 // Reservar espaço para header e footer

	if availableHeight < 5 {
		availableHeight = 5
	}

	startLine := m.detailViewScroll
	endLine := startLine + availableHeight
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine >= len(lines) {
		startLine = 0
		endLine = availableHeight
		if endLine > len(lines) {
			endLine = len(lines)
		}
	}

	// Renderizar as linhas visíveis
	for i := startLine; i < endLine; i++ {
		line := lines[i]
		// Truncar linha se for muito longa
		if len(line) > m.width-4 {
			line = line[:m.width-7] + "..."
		}
		b.WriteString(styles.TableCell.Render(line))
		b.WriteString("\n")
	}

	// Padding se necessário
	renderedLines := endLine - startLine
	for i := renderedLines; i < availableHeight; i++ {
		b.WriteString("\n")
	}

	// Separator
	b.WriteString("\n")
	b.WriteString(styles.Separator.Render(strings.Repeat("─", separatorWidth)))
	b.WriteString("\n")

	// Footer com instruções
	scrollInfo := ""
	if len(lines) > availableHeight {
		scrollInfo = styles.Faint.Render(
			fmt.Sprintf(
				"[%d-%d of %d lines] ",
				startLine+1,
				endLine,
				len(lines),
			),
		)
	}

	hjkl := styles.TableHeader.Render("↑↓") + styles.Faint.Render(" scroll")

	edit := ""
	if m.tableName != "" && m.primaryKeyCol != "" {
		edit = styles.TableHeader.Render("e") + styles.Faint.Render(" edit  ")
	}

	quit := styles.TableHeader.Render(
		"q/esc/enter",
	) + styles.Faint.Render(
		" close",
	)

	footer := fmt.Sprintf("\n%s%s  %s%s", scrollInfo, hjkl, edit, quit)
	b.WriteString(footer)

	return b.String()
}
