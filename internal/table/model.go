package table

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/parser"
)

type Model struct {
	width             int
	height            int
	selectedRow       int
	selectedCol       int
	offsetX           int
	offsetY           int
	visibleCols       int
	visibleRows       int
	columns           []string
	columnTypes       []string
	data              [][]string
	elapsed           time.Duration
	blinkCopiedCell   bool
	visualMode        bool
	visualStartRow    int
	visualStartCol    int
	dbConnection      db.DatabaseConnection
	tableName         string
	primaryKeyCol     string
	blinkUpdatedCell  bool
	updatedRow        int
	updatedCol        int
	blinkDeletedRow   bool
	deletedRow        int
	currentQuery      db.Query
	shouldRerunQuery  bool
	editedQuery       string
	lastExecutedQuery string
	cellWidth         int
	detailViewMode    bool
	detailViewContent string
	detailViewScroll  int
	isTablesList      bool
	onTableSelect     func(string) tea.Cmd
	selectedTableName string
	sortColumn        string
	sortDirection     string // "", "ASC", "DESC"
}

type blinkMsg struct{}

func New(
	columns []string,
	data [][]string,
	elapsed time.Duration,
	conn db.DatabaseConnection,
	tableName, primaryKeyCol string,
	query db.Query,
	columnWidth int,
) Model {
	columnTypes := make([]string, len(columns))
	if tableName != "" && conn != nil {
		metadata, err := conn.GetTableMetadata(tableName)

		if err == nil && metadata != nil {
			colTypeMap := map[string]string{}
			for i, colName := range metadata.Columns {
				if i < len(metadata.ColumnTypes) {
					colTypeMap[colName] = metadata.ColumnTypes[i]
				}
			}
			for i, col := range columns {
				if t, ok := colTypeMap[col]; ok {
					columnTypes[i] = t
				}
			}
		}
	}

	// Extract sort information from query if present
	sortCol, sortDir := extractSortFromQuery(query.SQL)

	return Model{
		selectedRow:      0,
		selectedCol:      0,
		offsetX:          0,
		offsetY:          0,
		columns:          columns,
		columnTypes:      columnTypes,
		data:             data,
		elapsed:          elapsed,
		visualMode:       false,
		dbConnection:     conn,
		tableName:        tableName,
		primaryKeyCol:    primaryKeyCol,
		currentQuery:     query,
		shouldRerunQuery: false,
		editedQuery:      "",
		cellWidth:        columnWidth,
		isTablesList:     false,
		sortColumn:       sortCol,
		sortDirection:    sortDir,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) numRows() int {
	return len(m.data)
}

func (m Model) numCols() int {
	return len(m.columns)
}

func (m Model) ShouldRerunQuery() bool {
	return m.shouldRerunQuery
}

func (m Model) GetEditedQuery() db.Query {
	updatedQuery := m.currentQuery
	if m.editedQuery != "" {
		updatedQuery.SQL = m.editedQuery
	}
	return updatedQuery
}

func (m Model) calculateHeaderLines() int {
	titleLines := 1

	var queryToDisplay string
	if m.lastExecutedQuery != "" {
		queryToDisplay = m.lastExecutedQuery
	} else {
		queryToDisplay = m.currentQuery.SQL
	}

	formattedSQL := parser.FormatSQLWithLineBreaks(queryToDisplay)
	sqlLines := strings.Count(formattedSQL, "\n") + 1

	return titleLines + sqlLines + 1
}

func (m Model) SetTablesList(onSelect func(string) tea.Cmd) Model {
	m.isTablesList = true
	m.onTableSelect = onSelect
	return m
}

func (m Model) GetSelectedTableName() string {
	return m.selectedTableName
}

// toggleSort cycles through sort states for the current column:
// For tables list: no direction → ASC → DESC → no direction
// For regular tables: no sort → ASC → DESC → no sort
func (m Model) toggleSort() (Model, tea.Cmd) {
	if m.selectedCol < 0 || m.selectedCol >= len(m.columns) {
		return m, nil
	}

	currentCol := m.columns[m.selectedCol]

	// Cycle through sort states
	if m.sortColumn != currentCol {
		// New column selected - start with ASC
		m.sortColumn = currentCol
		m.sortDirection = "ASC"
	} else if m.sortDirection == "" {
		// Same column, no direction - change to ASC
		m.sortDirection = "ASC"
	} else if m.sortDirection == "ASC" {
		// Same column, was ASC - change to DESC
		m.sortDirection = "DESC"
	} else if m.sortDirection == "DESC" {
		// Same column, was DESC
		if m.isTablesList {
			// For tables list, keep column but remove direction
			m.sortDirection = ""
		} else {
			// For regular tables, remove both column and direction
			m.sortColumn = ""
			m.sortDirection = ""
		}
	}

	// Modify the query with ORDER BY clause
	m.editedQuery = m.applySortToQuery()
	m.shouldRerunQuery = true

	return m, tea.Quit
}

// applySortToQuery adds or modifies the ORDER BY clause in the SQL query
func (m Model) applySortToQuery() string {
	sql := m.currentQuery.SQL
	if m.lastExecutedQuery != "" {
		// Use the last executed query as base if available
		sql = m.lastExecutedQuery
	}

	sql = strings.TrimSpace(sql)
	sql = strings.TrimRight(sql, ";")

	// Extract LIMIT and OFFSET clauses if present
	var limitClause string
	limitRegex := regexp.MustCompile(
		`(?i)\s+(LIMIT\s+\d+(?:\s+OFFSET\s+\d+)?)\s*$`,
	)
	if match := limitRegex.FindStringSubmatch(sql); match != nil {
		limitClause = " " + match[1]
		sql = limitRegex.ReplaceAllString(sql, "")
	}

	// Remove only the last (outermost) ORDER BY clause
	// This prevents breaking subqueries that have their own ORDER BY
	sql = removeLastOrderBy(sql)

	sql = strings.TrimSpace(sql)

	// Add new ORDER BY if we have a sort column
	if m.sortColumn != "" {
		if m.sortDirection != "" {
			sql = fmt.Sprintf(
				"%s ORDER BY %s %s",
				sql,
				m.sortColumn,
				m.sortDirection,
			)
		} else {
			// No direction means default (ASC)
			sql = fmt.Sprintf(
				"%s ORDER BY %s",
				sql,
				m.sortColumn,
			)
		}
	}

	// Re-add LIMIT clause if it existed
	if limitClause != "" {
		sql = sql + limitClause
	}

	return sql
}

// removeLastOrderBy removes only the last ORDER BY clause in a SQL query
// This is important for queries with subqueries that have their own ORDER BY
func removeLastOrderBy(sql string) string {
	// Find all occurrences of ORDER BY (case-insensitive)
	re := regexp.MustCompile(`(?i)\s+ORDER\s+BY\s+`)
	matches := re.FindAllStringIndex(sql, -1)

	if len(matches) == 0 {
		return sql
	}

	// Find the last ORDER BY that is NOT inside parentheses (i.e., not in a subquery)
	// We check by counting parentheses from the start of the string
	var lastOuterOrderByStart int = -1

	for _, match := range matches {
		orderByPos := match[0]

		// Count open and close parentheses before this ORDER BY
		openParens := strings.Count(sql[:orderByPos], "(")
		closeParens := strings.Count(sql[:orderByPos], ")")

		// If we're at the same level (not inside parentheses), this is an outer ORDER BY
		if openParens == closeParens {
			lastOuterOrderByStart = orderByPos
		}
	}

	// If no outer ORDER BY found, don't remove anything
	if lastOuterOrderByStart == -1 {
		return sql
	}

	// Find the end of the ORDER BY clause
	// It ends at: LIMIT, OFFSET, semicolon, or end of string
	remainder := sql[lastOuterOrderByStart:]
	endMarkers := regexp.MustCompile(`(?i)\s+(LIMIT|OFFSET|;)`)
	endMatch := endMarkers.FindStringIndex(remainder)

	var lastOrderByEnd int
	if endMatch != nil {
		lastOrderByEnd = lastOuterOrderByStart + endMatch[0]
	} else {
		lastOrderByEnd = len(sql)
	}

	// Remove the last outer ORDER BY clause
	return sql[:lastOuterOrderByStart] + sql[lastOrderByEnd:]
}

// extractSortFromQuery analyzes a SQL query and extracts ORDER BY information
func extractSortFromQuery(sql string) (column string, direction string) {
	// Look for ORDER BY clause with optional direction (case-insensitive)
	// We need to find the LAST (outermost) ORDER BY, not the first one
	orderByRegex := regexp.MustCompile(
		`(?i)\s+ORDER\s+BY\s+(\w+)(?:\s+(ASC|DESC))?`,
	)

	// Find all matches
	allMatches := orderByRegex.FindAllStringSubmatch(sql, -1)

	if len(allMatches) == 0 {
		return "", ""
	}

	// We want to find the last ORDER BY that is NOT inside parentheses (i.e., not in a subquery)
	// Simple heuristic: find the last match that appears after all closing parentheses

	// Find all match positions
	allMatchIndexes := orderByRegex.FindAllStringIndex(sql, -1)

	// Find the position of the last closing parenthesis
	lastParenPos := strings.LastIndex(sql, ")")

	// Find the last ORDER BY that appears after the last closing paren
	var lastOuterMatch []string
	for i := len(allMatchIndexes) - 1; i >= 0; i-- {
		matchPos := allMatchIndexes[i][0]
		if lastParenPos == -1 || matchPos > lastParenPos {
			lastOuterMatch = allMatches[i]
			break
		}
	}

	// If we didn't find one after parens, use the last match overall
	if lastOuterMatch == nil {
		lastOuterMatch = allMatches[len(allMatches)-1]
	}

	if len(lastOuterMatch) >= 2 {
		column = lastOuterMatch[1]
		if len(lastOuterMatch) >= 3 && lastOuterMatch[2] != "" {
			direction = strings.ToUpper(lastOuterMatch[2])
		}
		return column, direction
	}

	return "", ""
}
