// Package table implements database modification operations for the TUI.
//
// This file contains cell editing, row deletion, and related operations that
// interact with both the TUI (external editor, status messages) and the database
// (using SQL generation helpers from internal/db).
package table

import (
	"database/sql"
	"encoding/json"
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
)

// editCell opens the current cell value in an external editor ($EDITOR or vi),
// allows the user to modify it, then updates the database if the value changed.
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

	// Pretty-print JSON if valid
	valueToEdit = prettyPrintJSON(valueToEdit)

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

	// Use db package helper to generate SQL
	updateSQL, args := db.BuildUpdateQuery(
		m.tableData.Connection.GetDbType(),
		m.tableData.TableName,
		cell,
		newValueStr,
		m.tableData.Rows[cell.RowIndex],
	)

	if err := executeAndVerify(m.tableData.Connection.GetDB(), updateSQL, args, "update"); err != nil {
		return m, m.setError(err.Error())
	}

	// Update local data
	if newValueStr == "" {
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].Value = "NULL"
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].RawValue = nil
	} else {
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].Value = newValueStr
		m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].RawValue = newValueStr
	}

	return m, m.setSuccess(msgUpdateSuccess)
}

// deleteRow deletes the currently selected row from the database.
// It uses all columns as WHERE conditions to ensure the correct row is deleted.
func (m Model) deleteRow() (tea.Model, tea.Cmd) {
	if m.tableData == nil || m.tableData.Connection == nil {
		return m, m.setError("Cannot delete row: no connection")
	}
	if m.selectedRow < 0 || m.selectedRow >= len(m.tableData.Rows) {
		return m, m.setError("Cannot delete row: invalid row")
	}

	// Use db package helper to generate SQL
	deleteSQL, args := db.BuildDeleteQuery(
		m.tableData.Connection.GetDbType(),
		m.tableData.TableName,
		m.tableData.Rows[m.selectedRow],
	)

	if deleteSQL == "" || strings.Contains(deleteSQL, "WHERE ") == false {
		return m, m.setError("Cannot delete row: no filter conditions")
	}

	if err := executeAndVerify(m.tableData.Connection.GetDB(), deleteSQL, args, "delete"); err != nil {
		return m, m.setError(err.Error())
	}

	// Remove row from local data
	m.tableData.Rows = append(m.tableData.Rows[:m.selectedRow], m.tableData.Rows[m.selectedRow+1:]...)

	// Adjust cursor if needed
	if m.selectedRow >= len(m.tableData.Rows) && m.selectedRow > 0 {
		m.selectedRow--
	}

	return m, m.setSuccess("Row deleted")
}

// clearCell sets the current cell value to NULL in the database.
func (m Model) clearCell() (tea.Model, tea.Cmd) {
	cell := m.getCurrentCell()
	if cell == nil || m.tableData.Connection == nil {
		return m, nil
	}

	// Use db package helper to generate SQL (empty string means NULL)
	updateSQL, args := db.BuildUpdateQuery(
		m.tableData.Connection.GetDbType(),
		m.tableData.TableName,
		cell,
		"", // Empty string triggers NULL in BuildUpdateQuery
		m.tableData.Rows[cell.RowIndex],
	)

	if err := executeAndVerify(m.tableData.Connection.GetDB(), updateSQL, args, "clear"); err != nil {
		return m, m.setError(err.Error())
	}

	// Update local data
	m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].Value = "NULL"
	m.tableData.Rows[cell.RowIndex][cell.ColumnIndex].RawValue = nil

	return m, m.setSuccess("Cell cleared")
}

// enterDeleteMode enters delete mode where the user can choose to delete a cell or row.
func (m Model) enterDeleteMode() (tea.Model, tea.Cmd) {
	if m.tableData == nil || m.tableData.TableName == "" {
		return m, m.setError("Cannot delete: table name unknown")
	}
	if m.tableData.Connection == nil {
		return m, m.setError("Cannot delete: no database connection")
	}
	cell := m.getCurrentCell()
	if cell == nil {
		return m, nil
	}
	m.deleteMode = true
	m.deleteTarget = "cell" // default to cell
	return m, nil
}

// executeDelete executes the delete operation based on the selected target (cell or row).
func (m Model) executeDelete() (tea.Model, tea.Cmd) {
	m.deleteMode = false
	target := m.deleteTarget
	m.deleteTarget = ""

	switch target {
	case "row":
		return m.deleteRow()
	default:
		// Default to cell deletion (clear to NULL)
		return m.clearCell()
	}
}

// enterDeleteConfirm enters confirmation mode for destructive operations.
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

// executeConfirmAction executes the confirmed action.
func (m Model) executeConfirmAction() (tea.Model, tea.Cmd) {
	m.confirmMode = false
	action := m.confirmAction
	m.confirmAction = ""

	if action == "clear_cell" {
		return m.clearCell()
	}
	return m, nil
}

// executeAndVerify executes a SQL statement and validates that exactly one row was affected.
// Returns an error if the execution fails, verification fails, or the affected row count is not 1.
func executeAndVerify(sqlDB *sql.DB, query string, args []any, operation string) error {
	result, err := sqlDB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("%s failed: %w", operation, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify %s: %w", operation, err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("%s failed: expected 1 row affected, got %d (row may have been modified or deleted)", operation, rowsAffected)
	}

	return nil
}

// prettyPrintJSON attempts to pretty-print JSON strings.
// Returns the original string if it's not valid JSON.
func prettyPrintJSON(s string) string {
	var data any
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return s
	}
	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return s
	}
	return string(pretty)
}
