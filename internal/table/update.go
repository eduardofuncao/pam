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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case blinkMsg:
		m.blinkCopiedCell = false
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg), nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		m.statusMessage = ""
		return m.moveUp(), nil
	case "down", "j":
		m.statusMessage = ""
		return m.moveDown(), nil
	case "left", "h":
		m.statusMessage = ""
		return m.moveLeft(), nil
	case "right", "l":
		m.statusMessage = ""
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

	case "y", "enter":
		return m.copySelection()

	case "e":
		return m.editCell()
	}

	return m, nil
}

func (m Model) handleWindowResize(msg tea.WindowSizeMsg) Model {
	m.width = msg.Width
	m.height = msg.Height

	m.visibleCols = (m.width - horizontalPadding) / (cellWidth + columnSeparator)
	if m.visibleCols > m.numCols() {
		m.visibleCols = m.numCols()
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

	tmpfile, err := os.CreateTemp("", "pam-cell-*.txt")
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
		editor = "vi"
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
		m.statusMessage = ""
		m.isError = false
		return m, nil
	}

	updateSQL, args := m.buildUpdateQuery(cell, newValueStr)

	_, err = m.tableData.Connection.GetDB().Exec(updateSQL, args...)
	if err != nil {
		m.statusMessage = fmt.Sprintf("Update failed: %v", err)
		m.isError = true
		return m, nil
	}

	// Update the cell with new value
	if newValueStr == "" {
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].Value = "NULL"
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].RawValue = nil
	} else {
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].Value = newValueStr
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].RawValue = newValueStr
	}

	m.statusMessage = "Updated successfully"
	m.isError = false

	return m, nil
}

func (m Model) buildUpdateQuery(cell *db.Cell, newValue string) (string, []any) {
	var conditions []string
	var args []any
	paramIndex := 1

	dbType := m.tableData.Connection.GetDbType()

	// Handle empty string as NULL
	var setClause string
	if newValue == "" {
		setClause = fmt.Sprintf("%s = NULL", cell.ColumnName)
	} else {
		args = append(args, newValue)
		setClause = fmt.Sprintf("%s = %s", cell.ColumnName, m.placeholder(dbType, paramIndex))
		paramIndex++
	}

	// Build WHERE conditions using all other columns in the row
	for _, c := range m.tableData.Rows[cell.RowIndex] {
		if c.ColumnIndex == cell.ColumnIndex {
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

	whereClause := strings.Join(conditions, " AND ")

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		m.tableData.TableName, setClause, whereClause)

	return sql, args
}

func (m Model) placeholder(dbType string, index int) string {
	switch dbType {
	case "postgres":
		return fmt.Sprintf("$%d", index)
	case "oracle":
		return fmt.Sprintf(":%d", index)
	default: // mysql, sqlite3
		return "?"
	}
}
