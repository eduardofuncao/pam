package table

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/eduardofuncao/pam/internal/styles"
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
	endCol := min(m. offsetX+m.visibleCols, m.numCols())

	for j := m.offsetX; j < endCol; j++ {
		typeIcon := ""
		if j < len(m.columnTypes) && m.columnTypes[j] != "" {
			typeIcon = getTypeIcon(m.columnTypes[j]) + " "
		}
		
		pkIcon := ""
		if m.primaryKeyCol != "" && j < len(m.columns) && m.columns[j] == m.primaryKeyCol {
			pkIcon = "⚿ "
		}
		
		columnDisplay := pkIcon + typeIcon + m.columns[j]
		content := formatCell(columnDisplay)
		cells = append(cells, styles.TableHeader.Render(content))
	}

	return strings.Join(cells, styles. TableBorder.Render("│"))
}

func (m Model) renderDataRow(rowIndex int) string {
	var cells []string
	endCol := min(m.offsetX+m.visibleCols, m.numCols())

	for j := m.offsetX; j < endCol; j++ {
		content := formatCell(m.data[rowIndex][j])
		style := m.getCellStyle(rowIndex, j)
		cells = append(cells, style.Render(content))
	}

	return strings.Join(cells, styles.TableBorder.Render("│"))
}

func (m Model) renderFooter() string {
	updateInfo := ""
	if m.tableName != "" && m.primaryKeyCol != "" {
		updateInfo = styles. TableHeader.Render("u") + styles.Faint. Render("pdate")
	} else if m.tableName != "" {
		updateInfo = styles. TableHeader.Render("u") + styles.Faint. Render("pdate (no PK)")
	}

	sel := styles.TableHeader.Render("v") + styles.Faint.Render("sel")
	del := styles.TableHeader.Render("d") + styles.Faint.Render("el")
	yank := styles.TableHeader.Render("y") + styles.Faint.Render("ank")
	cmd := styles.TableHeader. Render(";") + styles.Faint.Render("cmd")
	quit := styles.TableHeader.Render("q") + styles.Faint.Render("uit")
	hjkl := styles.TableHeader.Render("hjkl") + styles.Faint.Render("←↓↑→")

	footer := fmt.Sprintf("\n%s %s | %s | %s  %s  %s  %s  %s  %s  %s",
		styles.  Faint.Render(fmt. Sprintf("%dx%d", m.numRows(), m.numCols())),
		styles.Faint.Render(fmt. Sprintf("In %.2fs", m.elapsed.Seconds())),
		styles.Faint.  Render(fmt. Sprintf("[%d/%d]", m. selectedRow+1, m.selectedCol+1)),
		updateInfo, del, yank, sel, cmd, quit, hjkl)
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
	
	return styles. TableCell
}

func formatCell(content string) string {
	if len(content) > cellWidth {
		return content[:cellWidth-1] + "…"
	}
	return fmt.Sprintf("%-*s", cellWidth, content)
}

func getTypeIcon(typeName string) string {
	upper := strings.ToUpper(typeName)
	
	// String/Text types
	if strings.Contains(upper, "CHAR") || strings.Contains(upper, "TEXT") || 
	   strings.Contains(upper, "STRING") || strings.Contains(upper, "CLOB") ||
	   strings.Contains(upper, "VARCHAR") || strings.Contains(upper, "NVARCHAR") {
		return "α"
	}
	
	// Integer types
	if strings.Contains(upper, "INT") || strings.Contains(upper, "SERIAL") ||
	   strings.Contains(upper, "BIGINT") || strings.Contains(upper, "SMALLINT") ||
	   strings.Contains(upper, "TINYINT") {
		return "№"
	}
	
	// Decimal/Float types
	if strings.Contains(upper, "DECIMAL") || strings.Contains(upper, "NUMERIC") ||
	   strings.Contains(upper, "FLOAT") || strings.Contains(upper, "DOUBLE") ||
	   strings.Contains(upper, "REAL") || strings.Contains(upper, "NUMBER") ||
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
	   strings.Contains(upper, "VARBINARY") || strings.Contains(upper, "IMAGE") {
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
	if strings. Contains(upper, "GEOMETRY") || strings.Contains(upper, "POINT") ||
	   strings.Contains(upper, "POLYGON") || strings.Contains(upper, "LINE") {
		return "◉"
	}
	
	// Default fallback
	return "•"
}
