// Package db provides database operations including SQL generation and query execution.
package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/eduardofuncao/pam/internal/db/connections"
)

const (
	dbTypePostgres  = "postgres"
	dbTypeOracle    = "oracle"
	dbTypeSqlServer = "sqlserver"
)

// BuildUpdateQuery generates a database-specific UPDATE statement.
// It sets the specified column to newValue and uses all other columns in the row
// as WHERE clause conditions for safe updates.
//
// Parameters:
//   - dbType: Database type (postgres, oracle, etc.) for placeholder syntax
//   - tableName: Name of the table to update
//   - cell: The cell being updated (contains column info and row position)
//   - newValue: New value for the cell (empty string means SET to NULL)
//   - row: The entire row data for building WHERE clause
//
// Returns SQL query string and arguments slice.
func BuildUpdateQuery(dbType, tableName string, cell *Cell, newValue string, row Row) (string, []any) {
	var setClause string
	var setArgs []any
	paramIndex := 1

	if newValue == "" {
		setClause = fmt.Sprintf("%s = NULL", cell.ColumnName)
	} else {
		setArgs = append(setArgs, newValue)
		setClause = fmt.Sprintf("%s = %s", cell.ColumnName, GetPlaceholder(dbType, paramIndex))
		paramIndex++
	}

	whereClause, whereArgs := BuildRowFilter(dbType, row, cell.ColumnIndex, paramIndex)
	args := append(setArgs, whereArgs...)

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, setClause, whereClause)

	return sql, args
}

// BuildDeleteQuery generates a database-specific DELETE statement.
// It uses all columns in the row as WHERE clause conditions for safe deletion.
//
// Parameters:
//   - dbType: Database type (postgres, oracle, etc.) for placeholder syntax
//   - tableName: Name of the table to delete from
//   - row: The row data for building WHERE clause
//
// Returns SQL query string and arguments slice.
func BuildDeleteQuery(dbType, tableName string, row Row) (string, []any) {
	whereClause, args := BuildRowFilter(dbType, row, -1, 1)
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, whereClause)
	return sql, args
}

// BuildRowFilter generates a WHERE clause using all columns in the row (except excludeCol if >= 0).
// This creates a safe filter that matches the exact row state, preventing accidental updates
// to modified or deleted rows.
//
// Parameters:
//   - dbType: Database type for placeholder syntax
//   - row: The row data to build conditions from
//   - excludeCol: Column index to exclude from WHERE clause (typically the column being updated)
//   - paramStart: Starting parameter index for placeholders
//
// Returns WHERE clause string (without "WHERE" keyword) and arguments slice.
func BuildRowFilter(dbType string, row Row, excludeCol int, paramStart int) (string, []any) {
	var conditions []string
	var args []any
	paramIndex := paramStart

	for _, c := range row {
		if c.ColumnIndex == excludeCol {
			continue
		}
		if c.RawValue == nil {
			conditions = append(conditions, fmt.Sprintf("%s IS NULL", c.ColumnName))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s = %s", c.ColumnName, GetPlaceholder(dbType, paramIndex)))
			args = append(args, c.RawValue)
			paramIndex++
		}
	}

	return strings.Join(conditions, " AND "), args
}

// GetPlaceholder returns the database-specific parameter placeholder syntax.
func GetPlaceholder(dbType string, index int) string {
	switch dbType {
	case dbTypePostgres:
		return fmt.Sprintf("$%d", index)
	case dbTypeOracle:
		return fmt.Sprintf(":%d", index)
	case dbTypeSqlServer:
		return fmt.Sprintf("@p%d", index)
	default:
		return "?"
	}
}

type QueryDescriptor struct {
	Type       string // "explore", "saved_query", "direct_sql"
	TableName  string
	QueryName  string
	SQL        string
	Limit      int
	Connection connections.DatabaseConnection
}

// ExecuteQuery executes a query based on the descriptor and returns TableData.
// Returns (nil, nil) for DML statements that don't return rows.
func ExecuteQuery(desc *QueryDescriptor) (*TableData, error) {
	if err := desc.Connection.Open(); err != nil {
		return nil, fmt.Errorf("opening connection: %w", err)
	}

	var rows *sql.Rows
	var displaySQL string
	var err error

	switch desc.Type {
	case "explore":
		rows, err = desc.Connection.QueryTableWithLimit(desc.TableName, desc.Limit)
		if err != nil {
			return nil, fmt.Errorf("querying table: %w", err)
		}
		displaySQL = fmt.Sprintf("SELECT * FROM %s (LIMIT %d)", desc.TableName, desc.Limit)

	case "saved_query":
		result, err := desc.Connection.Query(desc.QueryName)
		if err != nil {
			return nil, fmt.Errorf("executing query: %w", err)
		}
		rows = result.(*sql.Rows)
		displaySQL = desc.SQL

	case "direct_sql":
		result, err := desc.Connection.QueryDirect(desc.SQL)
		if err != nil {
			return nil, fmt.Errorf("executing SQL: %w", err)
		}
		rows = result.(*sql.Rows)
		displaySQL = desc.SQL

	default:
		return nil, fmt.Errorf("unknown query type: %s", desc.Type)
	}

	columns, err := rows.Columns()
	if err != nil || len(columns) == 0 {
		rows.Close()
		return nil, nil
	}

	return BuildTableData(rows, displaySQL, desc.Connection)
}
