package table

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

type exportFormat string

const (
	exportCSV      exportFormat = "csv"
	exportJSON     exportFormat = "json"
	exportTSV      exportFormat = "tsv"
	exportHTML     exportFormat = "html"
	exportSQL      exportFormat = "sql"
	exportMarkdown exportFormat = "markdown"
)

type exportWaitingFormatState struct {
	active bool
}

type exportCompleteMsg struct {
	format     exportFormat
	cells      int
	formatName string
	err        error
}

func (m Model) startExportFormatSelection() (Model, tea.Cmd) {
	m.exportWaiting = exportWaitingFormatState{active: true}
	return m, nil
}

func (m Model) cancelExportFormatSelection() Model {
	m.exportWaiting = exportWaitingFormatState{active: false}
	return m
}

func (m Model) executeExportForFormat(key string) (Model, tea.Cmd) {
	if !m.exportWaiting.active {
		return m, nil
	}

	var format exportFormat
	var formatName string

	switch strings.ToLower(key) {
	case "c", "csv":
		format = exportCSV
		formatName = "CSV"
	case "j", "json":
		format = exportJSON
		formatName = "JSON"
	case "t", "tsv":
		format = exportTSV
		formatName = "TSV"
	case "h", "html":
		format = exportHTML
		formatName = "HTML"
	case "s", "sql":
		format = exportSQL
		formatName = "SQL"
	case "m", "markdown", "md":
		format = exportMarkdown
		formatName = "Markdown"
	case "esc", "q":
		return m.cancelExportFormatSelection(), nil
	default:
		return m, nil
	}

	m = m.cancelExportFormatSelection()

	minRow, maxRow, minCol, maxCol := m.getSelectionBounds()

	headers := make([]string, 0)
	for col := minCol; col <= maxCol; col++ {
		headers = append(headers, m.columns[col])
	}

	rows := make([][]string, 0)
	for row := minRow; row <= maxRow; row++ {
		dataRow := make([]string, 0)
		for col := minCol; col <= maxCol; col++ {
			dataRow = append(dataRow, m.data[row][col])
		}
		rows = append(rows, dataRow)
	}

	cellCount := (maxRow - minRow + 1) * (maxCol - minCol + 1)

	return m, func() tea.Msg {
		content, err := m.formatExportContent(headers, rows, format)
		if err != nil {
			return exportCompleteMsg{err: err}
		}

		if err := clipboard.WriteAll(content); err != nil {
			return exportCompleteMsg{err: err}
		}

		return exportCompleteMsg{
			format:     format,
			cells:      cellCount,
			formatName: formatName,
		}
	}
}

func (m Model) formatExportContent(headers []string, rows [][]string, format exportFormat) (string, error) {
	switch format {
	case exportCSV:
		return m.formatCSV(headers, rows)
	case exportJSON:
		return m.formatJSON(headers, rows)
	case exportTSV:
		return m.formatTSV(headers, rows)
	case exportHTML:
		return m.formatHTML(headers, rows)
	case exportSQL:
		return m.formatSQL(headers, rows)
	case exportMarkdown:
		return m.formatMarkdown(headers, rows)
	default:
		return m.formatCSV(headers, rows)
	}
}

func (m Model) formatCSV(headers []string, rows [][]string) (string, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	if err := writer.Write(headers); err != nil {
		return "", err
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (m Model) formatJSON(headers []string, rows [][]string) (string, error) {
	objects := make([]map[string]string, 0, len(rows))

	for _, row := range rows {
		obj := make(map[string]string)
		for i, header := range headers {
			if i < len(row) {
				obj[header] = row[i]
			}
		}
		objects = append(objects, obj)
	}

	data, err := json.MarshalIndent(objects, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (m Model) formatTSV(headers []string, rows [][]string) (string, error) {
	var buf strings.Builder

	buf.WriteString(strings.Join(headers, "\t") + "\n")
	for _, row := range rows {
		buf.WriteString(strings.Join(row, "\t") + "\n")
	}

	return buf.String(), nil
}

func (m Model) formatHTML(headers []string, rows [][]string) (string, error) {
	var buf strings.Builder

	// HTML document structure
	buf.WriteString("<!DOCTYPE html>\n")
	buf.WriteString("<html>\n")
	buf.WriteString("<head>\n")
	buf.WriteString("<meta charset=\"UTF-8\"/>\n")
	buf.WriteString("<style>\n")
	buf.WriteString("table {border-collapse: collapse; width: auto;}\n")
	buf.WriteString("th {font-family: sans-serif; border: 1px solid #ccc; padding: 8px; background-color: #f2f2f2; text-align: left; font-weight: bold;}\n")
	buf.WriteString("td {font-family: sans-serif; border: 1px solid #ccc; padding: 8px; text-align: left;}\n")
	buf.WriteString("tr.odd {background-color: #f9f9f9;}\n")
	buf.WriteString("h3 {font-family: sans-serif; font-size: 16px; font-weight: bold; margin: 0 0 10px 0;}\n")
	buf.WriteString("</style>\n")
	buf.WriteString("</head>\n")
	buf.WriteString("<body>\n")

	// Table title with query name and database info
	if m.currentQuery.Name != "" {
		title := escapeHTML(m.currentQuery.Name)
		if m.dbConnection != nil {
			dbName := m.dbConnection.GetName()
			dbType := m.dbConnection.GetDbType()
			if dbName != "" || dbType != "" {
				title = fmt.Sprintf("%s (%s/%s)", escapeHTML(m.currentQuery.Name), escapeHTML(dbName), escapeHTML(dbType))
			}
		}
		buf.WriteString(fmt.Sprintf("<h3>%s</h3>\n", title))
	}

	// Table structure
	buf.WriteString("<table>\n")
	buf.WriteString("<thead>\n")
	buf.WriteString("<tr>\n")
	for _, header := range headers {
		buf.WriteString(fmt.Sprintf("<th>%s</th>\n", escapeHTML(header)))
	}
	buf.WriteString("</tr>\n")
	buf.WriteString("</thead>\n")
	buf.WriteString("<tbody>\n")

	// Data rows with alternating colors
	for i, row := range rows {
		rowClass := ""
		if i%2 == 1 {
			rowClass = " class=\"odd\""
		}
		buf.WriteString(fmt.Sprintf("<tr%s>\n", rowClass))
		for _, cell := range row {
			buf.WriteString(fmt.Sprintf("<td>%s</td>\n", escapeHTML(cell)))
		}
		buf.WriteString("</tr>\n")
	}

	buf.WriteString("</tbody>\n")
	buf.WriteString("</table>\n")
	buf.WriteString("</body>\n")
	buf.WriteString("</html>")

	return buf.String(), nil
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

func (m Model) formatSQL(headers []string, rows [][]string) (string, error) {
	if m.tableName == "" {
		return "", fmt.Errorf("no table name available for SQL export")
	}

	var buf strings.Builder

	for _, row := range rows {
		buf.WriteString(fmt.Sprintf("INSERT INTO %s (", m.tableName))

		columns := make([]string, 0, len(headers))
		for _, header := range headers {
			columns = append(columns, fmt.Sprintf(`"%s"`, header))
		}

		buf.WriteString(strings.Join(columns, ", "))
		buf.WriteString(") VALUES (")

		values := make([]string, 0, len(row))
		for _, val := range row {
			if val == "" || val == "NULL" {
				values = append(values, "NULL")
			} else {
				values = append(values, fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "''")))
			}
		}

		buf.WriteString(strings.Join(values, ", "))
		buf.WriteString(");\n")
	}

	return buf.String(), nil
}

func (m Model) formatMarkdown(headers []string, rows [][]string) (string, error) {
	var buf strings.Builder

	buf.WriteString("|")
	for _, header := range headers {
		buf.WriteString(" " + header + " |")
	}
	buf.WriteString("\n")

	buf.WriteString("|")
	for range headers {
		buf.WriteString(" --- |")
	}
	buf.WriteString("\n")

	for _, row := range rows {
		buf.WriteString("|")
		for _, cell := range row {
			buf.WriteString(" " + cell + " |")
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func (m Model) handleExportComplete(msg exportCompleteMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.exportStatus = fmt.Sprintf("Export failed: %v", msg.err)
	} else {
		cellText := "cells"
		if msg.cells == 1 {
			cellText = "cell"
		}
		m.exportStatus = fmt.Sprintf("Copied %d %s as %s to clipboard", msg.cells, cellText, msg.formatName)
		m.blinkCopiedCell = true
	}

	return m, tea.Batch(
		func() tea.Msg {
			time.Sleep(2 * time.Second)
			return clearExportStatusMsg{}
		},
		m.blinkCmd(),
	)
}

type clearExportStatusMsg struct{}

func (m Model) handleClearExportStatus() Model {
	m.exportStatus = ""
	return m
}
