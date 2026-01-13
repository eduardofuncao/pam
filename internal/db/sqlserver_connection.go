
package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/microsoft/go-mssqldb"
)

type SQLServerConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewSQLServerConnection(name, connStr string) (*SQLServerConnection, error) {
	bc := &BaseConnection{
		Name:       name,
		DbType:     "sqlserver",
		ConnString: connStr,
	}
	return &SQLServerConnection{BaseConnection: bc}, nil
}

func (s *SQLServerConnection) Open() error {
	db, err := sql.Open("sqlserver", s.ConnString)
	if err != nil {
		return err
	}
	s.db = db

	if s.Schema != "" {
		setSchemaSQL := fmt.Sprintf("ALTER SESSION SET CURRENT_SCHEMA = %s", s.Schema)
		_, err = s.db.Exec(setSchemaSQL)
		if err != nil {
			s.db.Close()
			return fmt.Errorf("failed to set schema to '%s': %w", s.Schema, err)
		}
	}

	return nil
}

func (s *SQLServerConnection) Ping() error {
	if s.db == nil {
		return fmt.Errorf("database is not open")
	}
	return s.db.Ping()
}

func (s *SQLServerConnection) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLServerConnection) Query(queryName string, args ...any) (any, error) {
	query, exists := s.Queries[queryName]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", queryName)
	}
	return s.db.Query(query.SQL, args...)
}

func (s *SQLServerConnection) ExecQuery(sql string, args ...any) (*sql.Rows, error) {
	return s.db.Query(sql, args...)
}

func (s *SQLServerConnection) Exec(sql string, args ...any) error {
	_, err := s.db.Exec(sql, args...)
	return err
}

func (s *SQLServerConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var currentSchema string
	schemaQuery := `SELECT SCHEMA_NAME()`
	row := s.db.QueryRow(schemaQuery)
	if err := row.Scan(&currentSchema); err != nil {
		// Fallback to configured schema or 'dbo'
		if s.Schema != "" {
			currentSchema = s.Schema
		} else {
			currentSchema = "dbo"
		}
	}

	pkQuery := `
		SELECT kcu.COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
		JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
			ON kcu.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
			AND kcu.TABLE_SCHEMA = tc.TABLE_SCHEMA
			AND kcu.TABLE_NAME = tc.TABLE_NAME
		WHERE kcu.TABLE_NAME = @p1
			AND tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
			AND kcu.TABLE_SCHEMA = @p2
		ORDER BY kcu.ORDINAL_POSITION
	`

	rows, err := s.db.Query(pkQuery, tableName, currentSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlserver primary key: %w", err)
	}
	defer rows.Close()

	metadata := &TableMetadata{
		TableName: tableName,
	}

	if rows.Next() {
		var pkColumn string
		if err := rows.Scan(&pkColumn); err == nil {
			metadata.PrimaryKey = pkColumn
		}
	}

	colQuery := `
		SELECT COLUMN_NAME,
		       DATA_TYPE +
		       CASE
			       WHEN CHARACTER_MAXIMUM_LENGTH IS NOT NULL
			       THEN '(' + CAST(CHARACTER_MAXIMUM_LENGTH AS VARCHAR) + ')'
			       WHEN NUMERIC_PRECISION IS NOT NULL AND NUMERIC_SCALE IS NOT NULL
			       THEN '(' + CAST(NUMERIC_PRECISION AS VARCHAR) + ',' + CAST(NUMERIC_SCALE AS VARCHAR) + ')'
			       WHEN NUMERIC_PRECISION IS NOT NULL
			       THEN '(' + CAST(NUMERIC_PRECISION AS VARCHAR) + ')'
			       ELSE ''
		       END as FULL_TYPE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_NAME = @p1
		  AND TABLE_SCHEMA = @p2
		ORDER BY ORDINAL_POSITION
	`

	colRows, err := s.db.Query(colQuery, tableName, currentSchema)
	if err == nil {
		defer colRows.Close()
		for colRows.Next() {
			var colName, colType string
			if err := colRows.Scan(&colName, &colType); err == nil {
				metadata.Columns = append(metadata.Columns, colName)
				metadata.ColumnTypes = append(metadata.ColumnTypes, colType)
			}
		}
	}

	return metadata, nil
}

func (s *SQLServerConnection) GetInfoSQL(infoType string) string {
	schema := s.Schema
	if schema == "" {
		schema = "dbo"
	}
	schema = "'" + schema + "'"

	switch infoType {
	case "tables":
		return fmt.Sprintf(`SELECT
			s.NAME as [schema],
			t.NAME as name,
			s.NAME as owner
		FROM sys.tables t
		INNER JOIN sys.schemas s ON t.schema_id = s.schema_id
		WHERE s.NAME = %s
		ORDER BY s.NAME, t.NAME`, schema)
	case "views":
		return fmt.Sprintf(`SELECT
			s.NAME as [schema],
			v.NAME as name,
			s.NAME as owner
		FROM sys.views v
		INNER JOIN sys.schemas s ON v.schema_id = s.schema_id
		WHERE s.NAME = %s
		ORDER BY s.NAME, v.NAME`, schema)
	default:
		return ""
	}
}

func (s *SQLServerConnection) BuildUpdateStatement(tableName, columnName, currentValue, pkColumn, pkValue string) string {
	quotedTableName := fmt.Sprintf("%s", tableName)
	quotedColumnName := fmt.Sprintf("%s", columnName)

	escapedValue := strings.ReplaceAll(currentValue, "'", "''")

	if pkColumn != "" && pkValue != "" {
		quotedPkColumn := fmt.Sprintf("%s", pkColumn)
		escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")
		return fmt.Sprintf(
			"-- SQL Server UPDATE statement\nUPDATE %s\nSET %s = '%s'\nWHERE %s = '%s';",
			quotedTableName,
			quotedColumnName,
			escapedValue,
			quotedPkColumn,
			escapedPkValue,
		)
	}

	return fmt.Sprintf(
		"-- SQL Server UPDATE statement\n-- No primary key specified. Edit WHERE clause manually.\nUPDATE %s\nSET %s = '%s'\nWHERE <condition>;",
		quotedTableName,
		quotedColumnName,
		escapedValue,
	)
}

func (s *SQLServerConnection) BuildDeleteStatement(tableName, primaryKeyCol, pkValue string) string {
	quotedTableName := fmt.Sprintf("%s", tableName)
	quotedPkColumn := fmt.Sprintf("%s", primaryKeyCol)
	escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")

	return fmt.Sprintf(
		"-- SQL Server DELETE statement\n-- WARNING:  This will permanently delete data!\n-- Ensure the WHERE clause is correct.\n\nDELETE FROM %s\nWHERE %s = '%s';",
		quotedTableName,
		quotedPkColumn,
		escapedPkValue,
	)
}

func (s *SQLServerConnection) ApplyRowLimit(sql string, limit int) string {
	trimmedSQL := strings.ToUpper(strings.TrimSpace(sql))

	if !strings.HasPrefix(trimmedSQL, "SELECT") && !strings.HasPrefix(trimmedSQL, "WITH") {
		return sql
	}

	upperSQL := strings.ToUpper(sql)

	if strings.Contains(upperSQL, " TOP ") {
		return sql
	}

	if strings.Contains(upperSQL, "OFFSET") && strings.Contains(upperSQL, "FETCH") {
		return sql
	}

	// Use TOP clause for SQL Server
	if strings.HasPrefix(trimmedSQL, "SELECT") {
		trimmed := strings.TrimLeft(sql, " \t")
		upperTrimmed := strings.ToUpper(trimmed)

		if strings.HasPrefix(upperTrimmed, "SELECT") {
			restOfSQL := trimmed[6:] // Remove "SELECT" (6 characters)
			restOfSQL = strings.TrimLeft(restOfSQL, " \t") // Remove any remaining whitespace after SELECT

			return fmt.Sprintf("SELECT TOP %d %s", limit, restOfSQL)
		}
	}

	return fmt.Sprintf("%s\nOFFSET 0 ROWS FETCH NEXT %d ROWS ONLY",
		strings.TrimRight(sql, ";"),
		limit)
}
