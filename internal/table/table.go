package table

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const cellWidth = 15

type TableModel struct {
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
	blinkCopiedCell bool
}

type blinkMsg struct{}

var (
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	cellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	copiedBlinkStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("205")).
				Bold(true)
)

func NewTableModel(columns []string, data [][]string) TableModel {
	return TableModel{
		selectedRow: 0,
		selectedCol: 0,
		offsetX:     0,
		offsetY:     0,
		columns:     columns,
		data:        data,
	}
}

func (m TableModel) Init() tea.Cmd {
	return nil
}

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	numColumns := len(m.columns)
	numRows := len(m.data)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.selectedRow > 0 {
				m.selectedRow--
				if m.selectedRow < m.offsetY {
					m.offsetY = m.selectedRow
				}
			}

		case "down", "j":
			if m.selectedRow < numRows-1 {
				m.selectedRow++
				if m.selectedRow >= m.offsetY+m.visibleRows {
					m.offsetY = m.selectedRow - m.visibleRows + 1
				}
			}

		case "left", "h":
			if m.selectedCol > 0 {
				m.selectedCol--
				if m.selectedCol < m.offsetX {
					m.offsetX = m.selectedCol
				}
			}

		case "right", "l":
			if m.selectedCol < numColumns-1 {
				m.selectedCol++
				if m.selectedCol >= m.offsetX+m.visibleCols {
					m.offsetX = m.selectedCol - m.visibleCols + 1
				}
			}

		case "home", "0", "_":
			m.selectedCol = 0
			m.offsetX = 0

		case "end", "$":
			m.selectedCol = numColumns - 1
			if m.visibleCols < numColumns {
				m.offsetX = numColumns - m.visibleCols
			}
		case "g":
			m.selectedRow = 0

		case "G":
			m.selectedRow = numRows - 1

		case "pgup", "ctrl+u":
			m.selectedRow -= m.visibleRows
			if m.selectedRow < 0 {
				m.selectedRow = 0
			}
			m.offsetY = m.selectedRow

		case "pgdown", "ctrl+d":
			m.selectedRow += m.visibleRows
			if m.selectedRow >= numRows {
				m.selectedRow = numRows - 1
			}
			if m.selectedRow >= m.offsetY+m.visibleRows {
				m.offsetY = m.selectedRow - m.visibleRows + 1
			}

		case "y", "enter":
			if m.selectedRow >= 0 && m.selectedRow < numRows && m.selectedCol >= 0 && m.selectedCol < numColumns {
				go clipboard.WriteAll(m.data[m.selectedRow][m.selectedCol])
				m.blinkCopiedCell = true
				return m, tea.Tick(time.Millisecond*400, func(time.Time) tea.Msg {
					return blinkMsg{}
				})
			}
		}
	case blinkMsg:
		m.blinkCopiedCell = false

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.visibleCols = (m.width - 2) / (cellWidth + 1)
		if m.visibleCols > numColumns {
			m.visibleCols = numColumns
		}
		m.visibleRows = m.height - 7
		if m.visibleRows > numRows {
			m.visibleRows = numRows
		}
		if m.visibleRows < 1 {
			m.visibleRows = 1
		}
	}

	return m, nil
}
func (m TableModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Render header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Render data rows
	endRow := min(m.offsetY+m.visibleRows, len(m.data))

	for i := m.offsetY; i < endRow; i++ {
		b.WriteString(m.renderDataRow(i))
		b.WriteString("\n")
	}

	// Render footer
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m TableModel) renderHeader() string {
	var cells []string

	endCol := min(m.offsetX+m.visibleCols, len(m.columns))

	for j := m.offsetX; j < endCol; j++ {
		content := m.formatCell(m.columns[j])
		cells = append(cells, headerStyle.Render(content))
	}

	return strings.Join(cells, borderStyle.Render("│"))
}

func (m TableModel) renderDataRow(rowIndex int) string {
	var cells []string

	endCol := min(m.offsetX+m.visibleCols, len(m.columns))

	for j := m.offsetX; j < endCol; j++ {
		content := m.formatCell(m.data[rowIndex][j])
		style := m.getCellStyle(rowIndex, j)
		cells = append(cells, style.Render(content))
	}

	return strings.Join(cells, borderStyle.Render("│"))
}

func (m TableModel) getCellStyle(row, col int) lipgloss.Style {
	if row == m.selectedRow && col == m.selectedCol {
		if m.blinkCopiedCell {
			return copiedBlinkStyle
		}
		return selectedStyle
	}
	return cellStyle
}

func (m TableModel) formatCell(content string) string {
	if len(content) > cellWidth {
		return content[:cellWidth-1] + "…"
	}
	return fmt.Sprintf("%-*s", cellWidth, content)
}

func (m TableModel) renderFooter() string {
	footer := fmt.Sprintf("\nPosition: Row %d/%d, Col %d/%d | Scroll: H/L (left/right), K/J (up/down) | Copy: y/enter",
		m.selectedRow+1, len(m.data), m.selectedCol+1, len(m.columns))
	return lipgloss.NewStyle().Faint(true).Render(footer)
}

func RenderTable(columns []string, data [][]string) error {
	model := NewTableModel(columns, data)
	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}
