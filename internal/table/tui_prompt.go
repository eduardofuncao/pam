package table

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) runCommand(input string) (tea.Model, tea.Cmd) {
	input = strings.TrimSpace(input)

	if input == "" {
		m.commandMode = false
		m.commandInput.Reset()
		return m, nil
	}

	if m.executeCommand == nil {
		m.commandMode = false
		m.commandInput.Reset()
		return m, m.setError("Command executor not available")
	}

	// Auto-prepend "run" if input looks like SQL
	if looksLikeSQL(input) {
		input = "run " + input
	}

	// Expand SQL if current table is available
	if m.tableData != nil && m.tableData.TableName != "" {
		input = m.expandSQL(input, m.tableData.TableName)
	}

	// Parse command into args
	args := []string{"pam"}
	args = append(args, strings.Fields(input)...)

	// Execute command via injected executor
	tableData, err := m.executeCommand(args)
	if err != nil {
		m.commandMode = false
		m.commandInput.Reset()
		errorMsg := strings.ReplaceAll(err.Error(), "\n", " ")
		return m, m.setError(errorMsg)
	}

	// If command returned TableData, switch to that view
	if tableData != nil {
		m.tableData = tableData
		m.columnWidths = calculateColumnWidths(tableData)
		m.selectedRow = 0
		m.selectedCol = 0
		m.offsetX = 0
		m.offsetY = 0
		m.commandMode = false
		m.commandInput.Reset()
		return m, m.setSuccess("View updated")
	}

	// No TableData returned, refresh original query
	if m.originalSQL != "" {
		refreshArgs := []string{"pam", "run", m.originalSQL}
		refreshData, refreshErr := m.executeCommand(refreshArgs)
		if refreshErr != nil {
			m.commandMode = false
			m.commandInput.Reset()
			return m, m.setError("Failed to refresh: " + refreshErr.Error())
		}
		if refreshData != nil {
			m.tableData = refreshData
			m.columnWidths = calculateColumnWidths(refreshData)
			m.selectedRow = 0
			m.selectedCol = 0
			m.offsetX = 0
			m.offsetY = 0
		}
	}

	m.commandMode = false
	m.commandInput.Reset()
	return m, m.setSuccess("Command executed")
}

func (m Model) expandSQL(input, tableName string) string {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return input
	}

	// Check if first word is "run" or "query"
	if parts[0] != "run" && parts[0] != "query" {
		return input
	}

	// Join the rest as SQL
	sql := strings.Join(parts[1:], " ")
	upperSQL := strings.ToUpper(sql)

	// Expand SELECT without FROM
	if strings.HasPrefix(upperSQL, "SELECT") && !strings.Contains(upperSQL, " FROM ") {
		// Find where to inject FROM
		keywords := []string{" WHERE ", " ORDER ", " GROUP ", " LIMIT ", " HAVING ", " UNION "}
		insertPos := len(sql)
		for _, kw := range keywords {
			if pos := strings.Index(upperSQL, kw); pos != -1 && pos < insertPos {
				insertPos = pos
			}
		}
		sql = sql[:insertPos] + " FROM " + tableName + sql[insertPos:]
	}

	// Expand UPDATE without table name
	if strings.HasPrefix(upperSQL, "UPDATE") && strings.Contains(upperSQL, " SET ") {
		// Check if table name is missing (UPDATE SET instead of UPDATE table SET)
		if strings.HasPrefix(upperSQL, "UPDATE SET") {
			sql = "UPDATE " + tableName + sql[len("UPDATE"):]
		}
	}

	// Expand DELETE without FROM
	if strings.HasPrefix(upperSQL, "DELETE") && !strings.Contains(upperSQL, " FROM ") {
		// Inject FROM after DELETE
		sql = "DELETE FROM " + tableName + sql[len("DELETE"):]
	}

	// Expand INSERT without INTO table
	if strings.HasPrefix(upperSQL, "INSERT") && !strings.Contains(upperSQL, " INTO ") {
		// Inject INTO table after INSERT
		sql = "INSERT INTO " + tableName + sql[len("INSERT"):]
	}

	return parts[0] + " " + sql
}

func looksLikeSQL(input string) bool {
	sqlKeywords := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "WITH", "EXPLAIN", "DESCRIBE", "SHOW", "PRAGMA"}
	upper := strings.ToUpper(strings.TrimSpace(input))

	for _, keyword := range sqlKeywords {
		if strings.HasPrefix(upper, keyword) {
			return true
		}
	}

	return false
}
