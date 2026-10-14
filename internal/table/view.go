package table

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/eduardofuncao/squix/internal/config"
	"github.com/eduardofuncao/squix/internal/parser"
	"github.com/eduardofuncao/squix/internal/styles"
)

func (m Model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("Loading...")
	}

	// If in detailed view mode, show the detailed view
	if m.detailViewMode {
		return tea.NewView(m.renderDetailView())
	}

	// If in help overlay mode, show the keybinds overlay
	if m.helpOverlayMode {
		return tea.NewView(m.renderHelpOverlay())
	}

	// Don't render if we're about to rerun the query (prevents duplicate output)
	if m.shouldRerunQuery {
		return tea.NewView("")
	}

	var b strings.Builder

	// Display query name header
	if m.uiVisibility.QueryName {
		b.WriteString(styles.Title.Render("◆ " + m.currentQuery.Name))
		b.WriteString("\n")
	}

	// Show the last executed query (for updates) or the current query (for selects)
	if m.uiVisibility.QuerySQL {
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
	}

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

	// Always add a newline for status message area
	b.WriteString("\n")

	// Display status message if present
	if m.statusMessage != "" {
		b.WriteString(styles.SearchMatch.Render(m.statusMessage))
	}

	return tea.NewView(b.String())
}

func (m Model) renderHeader() string {
	var cells []string
	endCol := min(m.offsetX+m.visibleCols, m.numCols())

	for j := m.offsetX; j < endCol; j++ {
		typeIcon := ""
		if m.uiVisibility.TypeDisplay && j < len(m.columnTypes) && m.columnTypes[j] != "" {
			typeIcon = getTypeIcon(m.columnTypes[j]) + " "
		}

		pkIcon := ""
		if m.uiVisibility.KeyIcons && m.primaryKeyCol != "" && j < len(m.columns) &&
			m.columns[j] == m.primaryKeyCol {
			pkIcon = "⚿ "
		}

		fkIcon := ""
		if m.uiVisibility.KeyIcons && j < len(m.columnFKs) && m.columnFKs[j] != "" {
			fkIcon = "⚭ "
		}

		columnDisplay := pkIcon + fkIcon + typeIcon + m.columns[j]
		content := formatCell(columnDisplay, m.cellWidth)

		var headerStyle lipgloss.Style
		if m.isColumnMatch(j) {
			headerStyle = styles.SearchMatch
		} else {
			headerStyle = styles.TableHeader
		}

		cells = append(cells, headerStyle.Render(content))
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
	// Show search input when active
	if m.searchMode {
		prompt := "/"
		if m.columnSearchMode {
			prompt = "f"
		}

		cursorBefore := m.searchQuery[:m.searchCursor]
		cursorAfter := ""
		if m.searchCursor < len(m.searchQuery) {
			cursorAfter = m.searchQuery[m.searchCursor:]
		}

		input := styles.SearchMatch.Render(prompt) + " " + cursorBefore + "█" + cursorAfter + "\n" + styles.Faint.Render("Enter: search  Esc: cancel")
		return "\n" + input
	}

	// Show export format prompt if active
	if m.exportWaiting.active {
		promptText := fmt.Sprintf(
			"Export as %sSV %sSON %sSV %sTML %sQL %sarkdown",
			styles.TableHeader.Render("[C]"),
			styles.TableHeader.Render("[J]"),
			styles.TableHeader.Render("[T]"),
			styles.TableHeader.Render("[H]"),
			styles.TableHeader.Render("[S]"),
			styles.TableHeader.Render("[M]"),
		)
		return "\n" + promptText
	}

	// Show export status if available
	if m.exportStatus != "" {
		return "\n" + styles.Success.Render(m.exportStatus)
	}

	// Build cell preview (conditional)
	cellPreview := ""
	if m.uiVisibility.FooterCellContent {
		currentCellValue := ""
		columnType := ""
		fkRef := ""

		if m.selectedRow >= 0 && m.selectedRow < len(m.data) &&
			m.selectedCol >= 0 && m.selectedCol < len(m.data[m.selectedRow]) {
			currentCellValue = m.data[m.selectedRow][m.selectedCol]
		}

		if m.selectedCol >= 0 && m.selectedCol < len(m.columnTypes) {
			columnType = m.columnTypes[m.selectedCol]
		}

		if m.selectedCol >= 0 && m.selectedCol < len(m.columnFKs) && m.columnFKs[m.selectedCol] != "" {
			fkRef = fmt.Sprintf(" FK → %s", m.columnFKs[m.selectedCol])
		}

		maxPreviewWidth := m.width - len(columnType) - len(fkRef) - 10
		displayValue := formatFooterPreview(currentCellValue, maxPreviewWidth)

		modeIndicator := ""
		if m.visualLineMode || m.visualMode {
			minR, maxR, minC, maxC := m.getSelectionBounds()
			label := "VISUAL"
			if m.visualLineMode {
				label = "V-LINE"
			}
			modeIndicator = styles.SearchMatch.Render(fmt.Sprintf("-- %s %dx%d -- ", label, maxR-minR+1, maxC-minC+1))
		}
		cellPreview = fmt.Sprintf("%s%s%s %s\n",
			modeIndicator,
			styles.Faint.Render(columnType),
			styles.Faint.Render(fkRef),
			styles.TableCell.Render(displayValue))
	}

	// Build stats info (conditional)
	statsInfo := ""
	if m.uiVisibility.FooterStats {
		statsInfo = fmt.Sprintf("%s | %s | %s",
			styles.Faint.Render(fmt.Sprintf("%dx%d", m.numRows(), m.numCols())),
			styles.Faint.Render(fmt.Sprintf("In %.2fs", m.elapsed.Seconds())),
			styles.Faint.Render(fmt.Sprintf("%d,%d", m.selectedRow+1, m.selectedCol+1)),
		)
		switch {
		case !m.columnSearchMode && len(m.searchMatches) > 0:
			cur, tot := m.cellMatchProgress()
			statsInfo += " | " + styles.Faint.Render(matchLabel(cur, tot))
		case m.columnSearchMode && len(m.searchColMatches) > 0:
			cur, tot := m.colMatchProgress()
			statsInfo += " | " + styles.Faint.Render(matchLabel(cur, tot))
		}
	}

	// Build keymaps info (conditional)
	keymapsInfo := ""
	if m.uiVisibility.FooterKeymaps {
		km := m.keyMap
		var parts []string
		addEntry := func(key, desc string) {
			parts = append(parts, styles.TableHeader.Render(key)+":"+styles.Faint.Render(desc))
		}
		addEntry(km.FirstKey(config.ActionMoveLeft)+km.FirstKey(config.ActionMoveDown)+km.FirstKey(config.ActionMoveUp)+km.FirstKey(config.ActionMoveRight), "←↓↑→")
		addEntry(km.FirstKey(config.ActionUpdate), "update")
		addEntry(km.FirstKey(config.ActionDeleteRow), "delete")
		addEntry(km.FirstKey(config.ActionVisualMode), "visual")
		addEntry(km.FirstKey(config.ActionYank), "copy")
		addEntry(km.FirstKey(config.ActionExport), "export")
		addEntry(km.FirstKey(config.ActionSearch), "search")
		addEntry(km.FirstKey(config.ActionHelp), "help")
		addEntry(km.FirstKey(config.ActionQuit), "quit")
		keymapsInfo = "  " + strings.Join(parts, " ")
	}

	// Assemble footer from conditional parts
	return fmt.Sprintf("\n%s%s%s", cellPreview, statsInfo, keymapsInfo)
}

func matchLabel(cur, tot int) string {
	return fmt.Sprintf("[%d/%d]", cur, tot)
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

	// Check if this cell is a search match
	if m.isCellSearchMatch(row, col) {
		return styles.SearchMatch
	}

	return styles.TableCell
}

func formatCell(content string, cellWidth int) string {
	content = strings.ReplaceAll(content, "\t", " ")
	if idx := strings.IndexAny(content, "\r\n\v\f"); idx >= 0 {
		content = content[:idx] + "…"
	}
	runes := []rune(content)
	runeCount := len(runes)

	if runeCount > cellWidth {
		return string(runes[:cellWidth-1]) + "…"
	}

	padding := cellWidth - runeCount
	return content + strings.Repeat(" ", padding)
}

func formatFooterPreview(value string, max int) string {
	value = strings.ReplaceAll(value, "\t", " ")
	if idx := strings.IndexAny(value, "\r\n\v\f"); idx >= 0 {
		value = value[:idx] + "…"
	}
	if max > 0 && len(value) > max {
		return value[:max-1] + "…"
	}
	return value
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
		strings.Contains(upper, "FLOAT") ||
		strings.Contains(upper, "DOUBLE") ||
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
	if strings.Contains(upper, "GEOMETRY") ||
		strings.Contains(upper, "POINT") ||
		strings.Contains(upper, "POLYGON") ||
		strings.Contains(upper, "LINE") {
		return "◉"
	}

	// Default fallback
	return "•"
}

func (m Model) renderHelpOverlay() string {
	type keyBind struct {
		keys string
		desc string
	}
	type category struct {
		name  string
		binds []keyBind
	}

	km := m.keyMap

	categories := []category{
		{"Navigation", []keyBind{
			{km.FirstKey(config.ActionMoveLeft), "Move left"},
			{km.FirstKey(config.ActionMoveDown), "Move down"},
			{km.FirstKey(config.ActionMoveUp), "Move up"},
			{km.FirstKey(config.ActionMoveRight), "Move right"},
			{km.DisplayKeys(config.ActionJumpFirstRow) + " / " + km.DisplayKeys(config.ActionJumpLastRow), "Jump to first/last row"},
			{km.DisplayKeys(config.ActionJumpFirstCol), "Jump to first column"},
			{km.DisplayKeys(config.ActionJumpLastCol), "Jump to last column"},
			{km.DisplayKeys(config.ActionPageUp), "Page up"},
			{km.DisplayKeys(config.ActionPageDown), "Page down"},
		}},
		{"Selection", []keyBind{
			{km.DisplayKeys(config.ActionVisualMode), "Visual (characterwise) mode"},
			{km.DisplayKeys(config.ActionVisualLineMode), "Visual line mode"},
		}},
		{"Actions", []keyBind{
			{km.DisplayKeys(config.ActionYank), "Yank (copy) selection"},
			{km.DisplayKeys(config.ActionExport), "Export selected cells"},
			{km.DisplayKeys(config.ActionExportAll), "Export all rows"},
			{km.DisplayKeys(config.ActionUpdate), "Update cell value"},
			{km.DisplayKeys(config.ActionDeleteRow), "Delete row"},
			{km.DisplayKeys(config.ActionEnter), "Open cell detail view"},
		}},
		{"Query", []keyBind{
			{km.DisplayKeys(config.ActionEditSQL), "Edit and re-run SQL"},
			{km.DisplayKeys(config.ActionSaveQuery), "Save query"},
			{km.DisplayKeys(config.ActionSearch), "Search cells"},
			{km.DisplayKeys(config.ActionSearchCol), "Search columns"},
			{km.DisplayKeys(config.ActionNextMatch) + " / " + km.DisplayKeys(config.ActionPrevMatch), "Next/previous search match"},
			{km.DisplayKeys(config.ActionPrevColMatch) + " / " + km.DisplayKeys(config.ActionNextColMatch), "Previous/next column match"},
		}},
		{"Other", []keyBind{
			{km.DisplayKeys(config.ActionToggleFooter), "Toggle footer keybinds"},
			{km.DisplayKeys(config.ActionHelp), "Show/hide this help"},
			{km.DisplayKeys(config.ActionQuit), "Quit"},
		}},
	}

	// Context adjustments
	if m.tableName == "" {
		categories[2].binds = []keyBind{
			{km.DisplayKeys(config.ActionYank), "Yank (copy) selection"},
			{km.DisplayKeys(config.ActionExport), "Export selected cells"},
			{km.DisplayKeys(config.ActionExportAll), "Export all rows"},
			{km.DisplayKeys(config.ActionEnter), "Open cell detail view"},
		}
	} else if m.primaryKeyCol == "" {
		categories[2].binds = []keyBind{
			{km.DisplayKeys(config.ActionYank), "Yank (copy) selection"},
			{km.DisplayKeys(config.ActionExport), "Export selected cells"},
			{km.DisplayKeys(config.ActionExportAll), "Export all rows"},
			{km.DisplayKeys(config.ActionUpdate), "Update cell value (no PK)"},
			{km.DisplayKeys(config.ActionEnter), "Open cell detail view"},
		}
	}

	separatorWidth := m.width - 4
	if separatorWidth < 0 {
		separatorWidth = 0
	}

	renderColumn := func(cats []category) string {
		var col strings.Builder
		for _, cat := range cats {
			col.WriteString(styles.TableHeader.Render(cat.name))
			col.WriteString("\n")
			for _, bind := range cat.binds {
				line := fmt.Sprintf("  %-16s %s", bind.keys, bind.desc)
				col.WriteString(styles.Faint.Render(line))
				col.WriteString("\n")
			}
		}
		return col.String()
	}

	colWidth := (separatorWidth - 2) / 2
	leftStr := renderColumn([]category{categories[0], categories[1], categories[4]})
	rightStr := renderColumn([]category{categories[2], categories[3]})
	leftCol := lipgloss.NewStyle().Width(colWidth).Render(leftStr)
	rightCol := lipgloss.NewStyle().Width(colWidth).Render(rightStr)

	var b strings.Builder
	b.WriteString(styles.Title.Render("Keyboard Shortcuts"))
	b.WriteString("\n")
	b.WriteString(styles.Separator.Render(strings.Repeat("─", separatorWidth)))
	b.WriteString("\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, leftCol, rightCol))
	b.WriteString("\n")
	b.WriteString(styles.Separator.Render(strings.Repeat("─", separatorWidth)))
	b.WriteString("\n")
	b.WriteString(styles.Faint.Render(km.DisplayKeys(config.ActionHelpClose) + " to close"))

	return b.String()
}

func (m Model) renderDetailView() string {
	var b strings.Builder

	// Get selected cell information
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

	// Position information
	posInfo := fmt.Sprintf(
		"Row %d, Column %d",
		m.selectedRow+1,
		m.selectedCol+1,
	)
	b.WriteString(styles.Faint.Render(posInfo))

	b.WriteString("\n\n")

	// Separator
	separatorWidth := m.width - 4
	if separatorWidth < 0 {
		separatorWidth = 0
	}
	b.WriteString(styles.Separator.Render(strings.Repeat("─", separatorWidth)))
	b.WriteString("\n\n")

	// Content with scroll (word-wrapped to viewport width)
	lines := m.wrappedDetailLines()
	availableHeight := m.detailViewportHeight()

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

	// Render visible lines
	for i := startLine; i < endLine; i++ {
		b.WriteString(styles.TableCell.Render(lines[i]))
		b.WriteString("\n")
	}

	// Padding if necessary
	renderedLines := endLine - startLine
	for i := renderedLines; i < availableHeight; i++ {
		b.WriteString("\n")
	}

	// Separator
	b.WriteString("\n")
	b.WriteString(styles.Separator.Render(strings.Repeat("─", separatorWidth)))
	b.WriteString("\n")

	// Footer with instructions
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

	hjkl := styles.TableHeader.Render(m.keyMap.FirstKey(config.ActionDetailScrollDown)+"/"+m.keyMap.FirstKey(config.ActionDetailScrollUp)) + styles.Faint.Render(" scroll")

	edit := ""
	if m.tableName != "" && m.primaryKeyCol != "" {
		edit = styles.TableHeader.Render(m.keyMap.FirstKey(config.ActionDetailEdit)) + styles.Faint.Render(" update")
	}

	yank := styles.TableHeader.Render(m.keyMap.FirstKey(config.ActionDetailYank)) + styles.Faint.Render(" yank")

	openHint := styles.TableHeader.Render(m.keyMap.FirstKey(config.ActionDetailOpen)) + styles.Faint.Render(" open")

	quit := styles.TableHeader.Render(m.keyMap.DisplayKeys(config.ActionDetailClose)) + styles.Faint.Render(" close")

	footer := fmt.Sprintf("\n%s  %s  %s  %s  %s  %s", scrollInfo, hjkl, edit, yank, openHint, quit)
	b.WriteString(footer)

	return b.String()
}

func (m Model) isCellSearchMatch(row, col int) bool {
	for _, match := range m.searchMatches {
		if match.Row == row && match.Col == col {
			return true
		}
	}
	return false
}

func (m Model) isColumnMatch(col int) bool {
	for _, matchCol := range m.searchColMatches {
		if matchCol == col {
			return true
		}
	}
	return false
}
