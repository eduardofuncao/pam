package table

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eduardofuncao/pam/internal/db"
)

const (
	defaultEditor      = "vi"
	tempFilePattern    = "pam-cell-*.txt"
	msgUpdateSuccess   = "Updated successfully"
	msgUpdateFailedFmt = "Update failed: %v"
	dbTypePostgres     = "postgres"
	dbTypeOracle       = "oracle"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case blinkMsg:
		m.blinkCopiedCell = false
	case clearStatusMsg:
		m.clearStatus()
		return m, nil
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg), nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle confirm mode keys
	if m.confirmMode {
		switch msg.Type {
		case tea.KeyEscape, tea.KeyCtrlC:
			m.confirmMode = false
			m.confirmAction = ""
			return m, nil
		case tea.KeyEnter:
			return m.executeConfirmAction()
		default:
			return m, nil
		}
	}

	// Handle command mode keys
	if m.commandMode {
		switch msg.Type {
		case tea.KeyEscape, tea.KeyCtrlC:
			m.commandMode = false
			m.commandInput.Reset()
			return m, nil
		case tea.KeyEnter:
			return m.runCommand(m.commandInput.Value())
		default:
			var cmd tea.Cmd
			m.commandInput, cmd = m.commandInput.Update(msg)
			return m, cmd
		}
	}

	// Normal mode keys
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case ";":
		m.commandMode = true
		m.commandInput.Focus()
		return m, nil

	case "up", "k":
		m.clearStatus()
		return m.moveUp(), nil
	case "down", "j":
		m.clearStatus()
		return m.moveDown(), nil
	case "left", "h":
		m.clearStatus()
		return m.moveLeft(), nil
	case "right", "l":
		m.clearStatus()
		return m.moveRight(), nil

	case "home", "0", "_":
		return m.jumpToFirstCol(), nil
	case "end", "$":
		return m.jumpToLastCol(), nil
	case "g":
		return m.jumpToFirstRow(), nil
	case "G":
		return m.jumpToLastRow(), nil

	case "pgup", "ctrl+u":
		return m.pageUp(), nil
	case "pgdown", "ctrl+d":
		return m.pageDown(), nil

	case "v":
		return m.toggleVisualMode()

	case "y":
		return m.copySelection()

	case "e":
		return m.editCell()

	case "d":
		return m.enterDeleteConfirm()
	}

	return m, nil
}

func (m Model) handleWindowResize(msg tea.WindowSizeMsg) Model {
	m.width = msg.Width
	m.height = msg.Height

	availableWidth := m.width - horizontalPadding
	m.visibleCols = 0
	widthUsed := 0

	for i := m.offsetX; i < m.numCols(); i++ {
		colWidth := cellWidth
		if i < len(m.columnWidths) {
			colWidth = m.columnWidths[i]
		}

		needWidth := colWidth
		if m.visibleCols > 0 {
			needWidth += columnSeparator
		}

		if widthUsed+needWidth > availableWidth {
			break
		}

		widthUsed += needWidth
		m.visibleCols++
	}

	if m.visibleCols == 0 && m.numCols() > 0 {
		m.visibleCols = 1
	}

	m.visibleRows = m.height - verticalReserved
	if m.visibleRows > m.numRows() {
		m.visibleRows = m.numRows()
	}
	if m.visibleRows < 1 {
		m.visibleRows = 1
	}

	return m
}

func (m Model) editCell() (tea.Model, tea.Cmd) {
	if m.tableData == nil || m.tableData.Connection == nil {
		return m, nil
	}

	if m.tableData.TableName == "" {
		log.Println("Cannot edit: table name unknown (complex query?)")
		return m, nil
	}

	cell := m.getCurrentCell()
	if cell == nil {
		return m, nil
	}

	tmpfile, err := os.CreateTemp("", tempFilePattern)
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		return m, nil
	}
	tmpfilePath := tmpfile.Name()
	defer os.Remove(tmpfilePath)

	// If cell is NULL, open empty editor
	valueToEdit := cell.Value
	if cell.RawValue == nil {
		valueToEdit = ""
	}

	if _, err := tmpfile.WriteString(valueToEdit); err != nil {
		log.Printf("Error writing to temp file: %v", err)
		tmpfile.Close()
		return m, nil
	}
	tmpfile.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = defaultEditor
	}

	cmd := exec.Command(editor, tmpfilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Error running editor: %v", err)
		return m, nil
	}

	newValue, err := os.ReadFile(tmpfilePath)
	if err != nil {
		log.Printf("Error reading edited file: %v", err)
		return m, nil
	}

	newValueStr := strings.TrimSpace(string(newValue))

	// Check if value actually changed
	oldValue := cell.Value
	if cell.RawValue == nil {
		oldValue = ""
	}
	if newValueStr == oldValue {
		m.clearStatus()
		return m, nil
	}

	updateSQL, args := m.buildUpdateQuery(cell, newValueStr)

	_, err = m.tableData.Connection.GetDB().Exec(updateSQL, args...)
	if err != nil {
		return m, m.setError(fmt.Sprintf(msgUpdateFailedFmt, err))
	}

	if newValueStr == "" {
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].Value = "NULL"
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].RawValue = nil
	} else {
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].Value = newValueStr
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].RawValue = newValueStr
	}

	return m, m.setSuccess(msgUpdateSuccess)
}

// buildRowFilter builds a WHERE clause using all columns in the row (except excludeCol if >= 0).
// Returns the WHERE clause string and args, starting parameters at paramStart.
func (m Model) buildRowFilter(rowIndex int, excludeCol int, paramStart int) (string, []any) {
	var conditions []string
	var args []any
	paramIndex := paramStart
	dbType := m.tableData.Connection.GetDbType()

	for _, c := range m.tableData.Rows[rowIndex] {
		if c.ColumnIndex == excludeCol {
			continue
		}
		if c.RawValue == nil {
			conditions = append(conditions, fmt.Sprintf("%s IS NULL", c.ColumnName))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s = %s", c.ColumnName, m.placeholder(dbType, paramIndex)))
			args = append(args, c.RawValue)
			paramIndex++
		}
	}

	return strings.Join(conditions, " AND "), args
}

func (m Model) buildUpdateQuery(cell *db.Cell, newValue string) (string, []any) {
	dbType := m.tableData.Connection.GetDbType()

	var setClause string
	var setArgs []any
	paramIndex := 1

	if newValue == "" {
		setClause = fmt.Sprintf("%s = NULL", cell.ColumnName)
	} else {
		setArgs = append(setArgs, newValue)
		setClause = fmt.Sprintf("%s = %s", cell.ColumnName, m.placeholder(dbType, paramIndex))
		paramIndex++
	}

	whereClause, whereArgs := m.buildRowFilter(cell.RowIndex, cell.ColumnIndex, paramIndex)
	args := append(setArgs, whereArgs...)

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		m.tableData.TableName, setClause, whereClause)

	return sql, args
}

func (m Model) placeholder(dbType string, index int) string {
	switch dbType {
	case dbTypePostgres:
		return fmt.Sprintf("$%d", index)
	case dbTypeOracle:
		return fmt.Sprintf(":%d", index)
	default:
		return "?"
	}
}

func (m Model) enterDeleteConfirm() (tea.Model, tea.Cmd) {
	if m.tableData == nil || m.tableData.TableName == "" {
		return m, m.setError("Cannot delete: table name unknown")
	}
	cell := m.getCurrentCell()
	if cell == nil {
		return m, nil
	}
	m.confirmMode = true
	m.confirmAction = "clear_cell"
	return m, nil
}

func (m Model) executeConfirmAction() (tea.Model, tea.Cmd) {
	m.confirmMode = false
	action := m.confirmAction
	m.confirmAction = ""

	if action == "clear_cell" {
		return m.clearCell()
	}
	return m, nil
}

func (m Model) clearCell() (tea.Model, tea.Cmd) {
	cell := m.getCurrentCell()
	if cell == nil || m.tableData.Connection == nil {
		return m, nil
	}

	// Reuse buildUpdateQuery with empty string (which sets to NULL)
	updateSQL, args := m.buildUpdateQuery(cell, "")

	_, err := m.tableData.Connection.GetDB().Exec(updateSQL, args...)
	if err != nil {
		return m, m.setError(fmt.Sprintf("Clear failed: %v", err))
	}

	// Update local data
	m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].Value = "NULL"
	m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].RawValue = nil

	return m, m.setSuccess("Cell cleared")
}
