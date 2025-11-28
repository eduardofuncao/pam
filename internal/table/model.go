package table

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eduardofuncao/pam/internal/db"
)

const cellWidth = 15

type Model struct {
	width           int
	height          int
	selectedRow     int
	selectedCol     int
	offsetX         int
	offsetY         int
	visibleCols     int
	visibleRows     int
	columns         []string
	data            [][]string
	elapsed         time.Duration
	blinkCopiedCell bool
	visualMode      bool
	visualStartRow  int
	visualStartCol  int
	dbConnection    db.DatabaseConnection
	tableName       string
	primaryKeyCol   string
}

type blinkMsg struct{}

func New(columns []string, data [][]string, elapsed time.Duration, conn db.DatabaseConnection, tableName, primaryKeyCol string) Model {
	return Model{
		selectedRow: 0,
		selectedCol: 0,
		offsetX:     0,
		offsetY:     0,
		columns:     columns,
		data:        data,
		elapsed:     elapsed,
		visualMode:  false,
		dbConnection:  conn,
		tableName:     tableName,
		primaryKeyCol: primaryKeyCol,
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
