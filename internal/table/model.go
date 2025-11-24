package table

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eduardofuncao/pam/internal/db"
)

const (
	cellWidth         = 15 // Width of each table cell in characters
	columnSeparator   = 1  // Width of the border separator between columns
	horizontalPadding = 2  // Padding for left/right table borders
	verticalReserved  = 9  // Reserved vertical space for header/footer
)

type Model struct {
	width           int
	height          int
	selectedRow     int
	selectedCol     int
	offsetX         int
	offsetY         int
	visibleCols     int
	visibleRows     int
	tableData       *db.TableData
	elapsed         time.Duration
	blinkCopiedCell bool
	visualMode      bool
	visualStartRow  int
	visualStartCol  int
	statusMessage   string
	isError         bool
}

type blinkMsg struct{}

func New(tableData *db.TableData, elapsed time.Duration) Model {
	return Model{
		selectedRow: 0,
		selectedCol: 0,
		offsetX:     0,
		offsetY:     0,
		tableData:   tableData,
		elapsed:     elapsed,
		visualMode:  false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) numRows() int {
	if m.tableData == nil {
		return 0
	}
	return len(m.tableData.Rows)
}

func (m Model) numCols() int {
	if m.tableData == nil {
		return 0
	}
	return len(m.tableData.Columns)
}

func (m Model) getCell(row, col int) *db.Cell {
	if m.tableData == nil || row < 0 || row >= len(m.tableData.Rows) {
		return nil
	}
	if col < 0 || col >= len(m.tableData.Rows[row]) {
		return nil
	}
	return &m.tableData.Rows[row][col]
}

func (m Model) getCurrentCell() *db.Cell {
	return m.getCell(m.selectedRow, m.selectedCol)
}
