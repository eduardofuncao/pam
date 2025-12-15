package table

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/mattn/go-runewidth"
)

const (
	cellWidth         = 15 // Width of each table cell in characters
	columnSeparator   = 1  // Width of the border separator between columns
	horizontalPadding = 2  // Padding for left/right table borders
	verticalReserved  = 9  // Reserved vertical space for header/footer
)

type CommandExecutor func(args []string) (*db.TableData, error)

type Model struct {
	width           int
	height          int
	selectedRow     int
	selectedCol     int
	offsetX         int
	offsetY         int
	visibleCols     int
	visibleRows     int
	columnWidths    []int
	tableData       *db.TableData
	elapsed         time.Duration
	blinkCopiedCell bool
	visualMode      bool
	visualStartRow  int
	visualStartCol  int
	statusMessage   string
	isError         bool
	commandMode     bool
	commandInput    textinput.Model
	queries         map[string]string
	originalSQL     string
	executeCommand  CommandExecutor
	confirmMode     bool
	confirmAction   string
	deleteMode      bool
	deleteTarget    string // "cell" or "row"
}

type blinkMsg struct{}

type clearStatusMsg struct{}

func New(tableData *db.TableData, elapsed time.Duration, cmdExec CommandExecutor) Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 500
	ti.Width = 80

	originalSQL := ""
	if tableData != nil {
		originalSQL = tableData.SQL
	}

	return Model{
		selectedRow:     0,
		selectedCol:     0,
		offsetX:         0,
		offsetY:         0,
		tableData:       tableData,
		elapsed:         elapsed,
		visualMode:      false,
		columnWidths:    calculateColumnWidths(tableData),
		commandMode:     false,
		commandInput:    ti,
		queries:         make(map[string]string),
		originalSQL:     originalSQL,
		executeCommand:  cmdExec,
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

// Status message helpers
func (m *Model) setStatus(msg string, isError bool, autoClear bool) tea.Cmd {
	m.statusMessage = msg
	m.isError = isError
	if autoClear {
		return tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
			return clearStatusMsg{}
		})
	}
	return nil
}

func (m *Model) setSuccess(msg string) tea.Cmd {
	return m.setStatus(msg, false, true)
}

func (m *Model) setError(msg string) tea.Cmd {
	return m.setStatus(msg, true, false)
}

func (m *Model) clearStatus() {
	m.statusMessage = ""
	m.isError = false
}

func getTypeMaxWidth(colType string) int {
	colTypeLower := strings.ToLower(colType)

	switch {
	case strings.Contains(colTypeLower, "bool"):
		return 5
	case strings.Contains(colTypeLower, "int"), strings.Contains(colTypeLower, "serial"):
		return 12
	case strings.Contains(colTypeLower, "uuid"):
		return 10
	case strings.Contains(colTypeLower, "json"):
		return 12
	case strings.Contains(colTypeLower, "date"), strings.Contains(colTypeLower, "time"):
		return 10
	default:
		return 20
	}
}

func calculateColumnWidths(tableData *db.TableData) []int {
	if tableData == nil || len(tableData.Columns) == 0 {
		return []int{}
	}

	widths := make([]int, len(tableData.Columns))

	for colIdx, colName := range tableData.Columns {
		headerWidth := runewidth.StringWidth(colName)

		var colType string
		if len(tableData.Rows) > 0 && colIdx < len(tableData.Rows[0]) {
			colType = tableData.Rows[0][colIdx].ColumnType
		}
		typeLimit := getTypeMaxWidth(colType)

		// charLimit is the ceiling - header wins if larger than type limit
		charLimit := typeLimit
		if headerWidth > typeLimit {
			charLimit = headerWidth
		}

		// Find largest content width
		maxContentWidth := 0
		for _, row := range tableData.Rows {
			if colIdx < len(row) {
				contentWidth := runewidth.StringWidth(row[colIdx].Value)
				if contentWidth > maxContentWidth {
					maxContentWidth = contentWidth
				}
			}
		}

		// Width: use content width, cap at charLimit, minimum is headerWidth
		width := maxContentWidth
		if width > charLimit {
			width = charLimit
		}
		if width < headerWidth {
			width = headerWidth
		}

		widths[colIdx] = width
	}

	return widths
}
