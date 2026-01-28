
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
			metadata.PrimaryKeys = append(metadata.PrimaryKeys, pkColumn)
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

	// Fetch foreign keys
	fks, err := s.GetForeignKeys(tableName)
	if err == nil {
		metadata.ForeignKeys = fks
	}

	return metadata, nil
}

func (s *SQLServerConnection) GetForeignKeys(tableName string) ([]ForeignKey, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var currentSchema string
	schemaQuery := `SELECT SCHEMA_NAME()`
	row := s.db.QueryRow(schemaQuery)
	if err := row.Scan(&currentSchema); err != nil {
		if s.Schema != "" {
			currentSchema = s.Schema
		} else {
			currentSchema = "dbo"
		}
	}

	query := `
		SELECT
			kcu.COLUMN_NAME,
			rcu.TABLE_NAME AS FOREIGN_TABLE_NAME,
			rcu.COLUMN_NAME AS FOREIGN_COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
		JOIN INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS rc
			ON kcu.CONSTRAINT_NAME = rc.CONSTRAINT_NAME
		JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE rcu
			ON rc.UNIQUE_CONSTRAINT_NAME = rcu.CONSTRAINT_NAME
		WHERE kcu.TABLE_NAME = @p1
		  AND kcu.TABLE_SCHEMA = @p2
		  AND rc.CONSTRAINT_SCHEMA = @p2
		ORDER BY kcu.COLUMN_NAME
	`

	rows, err := s.db.Query(query, tableName, currentSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	var foreignKeys []ForeignKey
	for rows.Next() {
		var fk ForeignKey
		if err := rows.Scan(&fk.Column, &fk.ReferencedTable, &fk.ReferencedColumn); err == nil {
			foreignKeys = append(foreignKeys, fk)
		}
	}

	return foreignKeys, nil
}

func (s *SQLServerConnection) GetForeignKeysReferencingTable(tableName string) ([]ForeignKey, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var currentSchema string
	schemaQuery := `SELECT SCHEMA_NAME()`
	row := s.db.QueryRow(schemaQuery)
	if err := row.Scan(&currentSchema); err != nil {
		if s.Schema != "" {
			currentSchema = s.Schema
		} else {
			currentSchema = "dbo"
		}
	}

	query := `
		SELECT
			kcu.COLUMN_NAME,
			kcu.TABLE_NAME AS FOREIGN_TABLE_NAME,
			rcu.COLUMN_NAME AS FOREIGN_COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
		JOIN INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS rc
			ON kcu.CONSTRAINT_NAME = rc.CONSTRAINT_NAME
		JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE rcu
			ON rc.UNIQUE_CONSTRAINT_NAME = rcu.CONSTRAINT_NAME
		WHERE rcu.TABLE_NAME = @p1
		  AND rcu.TABLE_SCHEMA = @p2
		  AND rc.CONSTRAINT_SCHEMA = @p2
		ORDER BY kcu.TABLE_NAME, kcu.COLUMN_NAME
	`

	rows, err := s.db.Query(query, tableName, currentSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to query referencing foreign keys: %w", err)
	}
	defer rows.Close()

	var foreignKeys []ForeignKey
	for rows.Next() {
		var fk ForeignKey
		if err := rows.Scan(&fk.Column, &fk.ReferencedTable, &fk.ReferencedColumn); err == nil {
			foreignKeys = append(foreignKeys, fk)
		}
	}

	return foreignKeys, nil
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

func (s *SQLServerConnection) GetTables() ([]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT t.NAME
		FROM sys.tables t
		INNER JOIN sys.schemas s ON t.schema_id = s.schema_id
		WHERE s.NAME = SCHEMA_NAME()
		ORDER BY t.NAME
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err == nil {
			tables = append(tables, tableName)
		}
	}

	return tables, nil
}

func (s *SQLServerConnection) GetViews() ([]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT v.NAME
		FROM sys.views v
		INNER JOIN sys.schemas s ON v.schema_id = s.schema_id
		WHERE s.NAME = SCHEMA_NAME()
		ORDER BY v.NAME
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query views: %w", err)
	}
	defer rows.Close()

	var views []string
	for rows.Next() {
		var viewName string
		if err := rows.Scan(&viewName); err == nil {
			views = append(views, viewName)
		}
	}

	return views, nil
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

func (s *SQLServerConnection) GetPlaceholder(paramIndex int) string {
	return "@p" + fmt.Sprintf("%d", paramIndex)
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
