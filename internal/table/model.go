package table

import (
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
	columnFKs         []string  // Maps column index to FK reference (e.g., "people.id")
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
	saveQueryCallback func(query db.Query) (db.Query, error)
	statusMessage     string
	exportWaiting     exportWaitingFormatState
	exportStatus      string
}

type blinkMsg struct{}

func New(
	columns []string,
	passedColumnTypes []string,
	data [][]string,
	elapsed time.Duration,
	conn db.DatabaseConnection,
	tableName, primaryKeyCol string,
	query db.Query,
	columnWidth int,
) Model {
	columnTypes := make([]string, len(columns))

	// Use passed column types from result set if available (works for JOINs, CTEs, etc.)
	if passedColumnTypes != nil && len(passedColumnTypes) > 0 {
		// Use types from SQL driver's ColumnTypes()
		for i := range columns {
			if i < len(passedColumnTypes) {
				columnTypes[i] = passedColumnTypes[i]
			}
		}
	} else if tableName != "" && conn != nil {
		// Fallback to querying table metadata for single-table queries
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

	// Build FK map for display
	columnFKs := make([]string, len(columns))
	if tableName != "" && conn != nil {
		metadata, err := conn.GetTableMetadata(tableName)
		if err == nil && metadata != nil && len(metadata.ForeignKeys) > 0 {
			// Build FK map with lowercase keys for case-insensitive lookup
			fkMap := map[string]db.ForeignKey{}
			for _, fk := range metadata.ForeignKeys {
				fkMap[strings.ToLower(fk.Column)] = fk
			}

			for i, col := range columns {
				if fk, ok := fkMap[strings.ToLower(col)]; ok {
					columnFKs[i] = fk.ReferencedTable + "." + fk.ReferencedColumn
				}
			}
		}
	}

	return Model{
		selectedRow:      0,
		selectedCol:      0,
		offsetX:          0,
		offsetY:          0,
		columns:          columns,
		columnTypes:      columnTypes,
		columnFKs:        columnFKs,
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
