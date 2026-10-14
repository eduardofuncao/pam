package table

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
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

// FormatOptions provides metadata for export formats that need it (HTML, SQL).
type FormatOptions struct {
	QueryName string
	DbType    string
	DbName    string
	TableName string
}

// ParseFormat validates and normalizes a format string.
func ParseFormat(s string) (string, error) {
	switch strings.ToLower(s) {
	case "csv":
		return "csv", nil
	case "json":
		return "json", nil
	case "tsv":
		return "tsv", nil
	case "html":
		return "html", nil
	case "sql":
		return "sql", nil
	case "markdown", "md":
		return "markdown", nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: csv, json, tsv, html, sql, markdown)", s)
	}
}

// FormatExport dispatches to the correct standalone format function.
func FormatExport(headers []string, rows [][]string, format string, opts FormatOptions) (string, error) {
	switch strings.ToLower(format) {
	case "csv":
		return FormatCSV(headers, rows)
	case "json":
		return FormatJSON(headers, rows)
	case "tsv":
		return FormatTSV(headers, rows)
	case "html":
		return FormatHTML(headers, rows, opts)
	case "sql":
		return FormatSQL(headers, rows, opts)
	case "markdown", "md":
		return FormatMarkdown(headers, rows)
	default:
		return FormatCSV(headers, rows)
	}
}

type exportWaitingFormatState struct {
	active  bool
	allRows bool
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

func (m Model) startExportAllFormatSelection() (Model, tea.Cmd) {
	m.exportWaiting = exportWaitingFormatState{active: true, allRows: true}
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

	allRows := m.exportWaiting.allRows
	m = m.cancelExportFormatSelection()

	var headers []string
	var rows [][]string
	var cellCount int

	if allRows {
		headers = m.columns
		rows = m.data
		cellCount = m.numRows() * m.numCols()
	} else {
		minRow, maxRow, minCol, maxCol := m.getSelectionBounds()

		headers = make([]string, 0)
		for col := minCol; col <= maxCol; col++ {
			headers = append(headers, m.columns[col])
		}

		rows = make([][]string, 0)
		for row := minRow; row <= maxRow; row++ {
			dataRow := make([]string, 0)
			for col := minCol; col <= maxCol; col++ {
				dataRow = append(dataRow, m.data[row][col])
			}
			rows = append(rows, dataRow)
		}

		cellCount = (maxRow - minRow + 1) * (maxCol - minCol + 1)
	}

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
	opts := FormatOptions{
		QueryName: m.currentQuery.Name,
		TableName: m.tableName,
	}

	if m.dbConnection != nil {
		opts.DbType = m.dbConnection.GetDbType()
		opts.DbName = m.dbConnection.GetName()
	}

	return FormatExport(headers, rows, string(format), opts)
}

// --- Standalone format functions ---

func FormatCSV(headers []string, rows [][]string) (string, error) {
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

func FormatJSON(headers []string, rows [][]string) (string, error) {
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

func FormatTSV(headers []string, rows [][]string) (string, error) {
	var buf strings.Builder

	buf.WriteString(strings.Join(headers, "\t") + "\n")
	for _, row := range rows {
		buf.WriteString(strings.Join(row, "\t") + "\n")
	}

	return buf.String(), nil
}

func FormatHTML(headers []string, rows [][]string, opts FormatOptions) (string, error) {
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
	if opts.QueryName != "" {
		title := escapeHTML(opts.QueryName)
		if opts.DbName != "" || opts.DbType != "" {
			title = fmt.Sprintf("%s (%s/%s)", escapeHTML(opts.QueryName), escapeHTML(opts.DbName), escapeHTML(opts.DbType))
		}
		fmt.Fprintf(&buf, "<h3>%s</h3>\n", title)
	}

	// Table structure
	buf.WriteString("<table>\n")
	buf.WriteString("<thead>\n")
	buf.WriteString("<tr>\n")
	for _, header := range headers {
		fmt.Fprintf(&buf, "<th>%s</th>\n", escapeHTML(header))
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
		fmt.Fprintf(&buf, "<tr%s>\n", rowClass)
		for _, cell := range row {
			fmt.Fprintf(&buf, "<td>%s</td>\n", escapeHTML(cell))
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

func FormatSQL(headers []string, rows [][]string, opts FormatOptions) (string, error) {
	if opts.TableName == "" {
		return "", fmt.Errorf("no table name available for SQL export")
	}

	var buf strings.Builder

	for _, row := range rows {
		fmt.Fprintf(&buf, "INSERT INTO %s (", opts.TableName)

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

func FormatMarkdown(headers []string, rows [][]string) (string, error) {
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
