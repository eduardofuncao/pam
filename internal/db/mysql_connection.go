package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewMySQLConnection(name, connStr string) (*MySQLConnection, error) {
	bc := &BaseConnection{
		Name:       name,
		DbType:     "mysql",
		ConnString: connStr,
	}
	return &MySQLConnection{BaseConnection: bc}, nil
}

func (m *MySQLConnection) Open() error {
	db, err := sql.Open("mysql", m.ConnString)
	if err != nil {
		return err
	}
	m.db = db

	if m.Schema != "" {
		setDatabaseSQL := fmt.Sprintf("USE `%s`", m.Schema)
		_, err = m.db.Exec(setDatabaseSQL)
		if err != nil {
			m.db.Close()
			return fmt.Errorf(
				"failed to set database to '%s': %w",
				m.Schema,
				err,
			)
		}
	}

	return nil
}

func (m *MySQLConnection) Ping() error {
	if m.db == nil {
		return fmt.Errorf("database is not open")
	}
	return m.db.Ping()
}

func (m *MySQLConnection) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

func (m *MySQLConnection) Query(queryName string, args ...any) (any, error) {
	query, exists := m.Queries[queryName]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", queryName)
	}
	return m.db.Query(query.SQL, args...)
}

func (m *MySQLConnection) ExecQuery(
	sql string,
	args ...any,
) (*sql.Rows, error) {
	return m.db.Query(sql, args...)
}

func (m *MySQLConnection) Exec(sql string, args ...any) error {
	_, err := m.db.Exec(sql, args...)
	return err
}

func (m *MySQLConnection) GetTableMetadata(
	tableName string,
) (*TableMetadata, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	pkQuery := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE TABLE_NAME = ?
		AND CONSTRAINT_NAME = 'PRIMARY'
		AND TABLE_SCHEMA = DATABASE()
		ORDER BY ORDINAL_POSITION
		LIMIT 1
	`

	rows, err := m.db.Query(pkQuery, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query mysql primary key: %w", err)
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
		       COLUMN_TYPE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_NAME = ?
		AND TABLE_SCHEMA = DATABASE()
		ORDER BY ORDINAL_POSITION
	`

	colRows, err := m.db.Query(colQuery, tableName)
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

func (m *MySQLConnection) GetTables() ([]string, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	query := `
		SELECT TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %w", err)
	}

	return tables, nil
}

func (m *MySQLConnection) GetInfoSQL(infoType string) string {
	switch infoType {
	case "tables":
		return "SELECT TABLE_SCHEMA as `schema`,\n		       TABLE_NAME as name\n\t	FROM information_schema.TABLES\n\t\tWHERE TABLE_SCHEMA = DATABASE()\n\t\t  AND TABLE_TYPE = 'BASE TABLE'\n\t\tORDER BY TABLE_SCHEMA, TABLE_NAME"
	case "views":
		return "SELECT TABLE_SCHEMA as `schema`,\n		       TABLE_NAME as name\n\t	FROM information_schema.VIEWS\n\t\tWHERE TABLE_SCHEMA = DATABASE()\n\t\tORDER BY TABLE_SCHEMA, TABLE_NAME"
	default:
		return ""
	}
}

func (m *MySQLConnection) BuildUpdateStatement(
	tableName, columnName, currentValue, pkColumn, pkValue string,
) string {
	escapedValue := strings.ReplaceAll(currentValue, "'", "''")

	if pkColumn != "" && pkValue != "" {
		escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")
		return fmt.Sprintf(
			"-- MySQL UPDATE statement\nUPDATE %s\nSET %s = '%s'\nWHERE %s = '%s';",
			tableName,
			columnName,
			escapedValue,
			pkColumn,
			escapedPkValue,
		)
	}

	return fmt.Sprintf(
		"-- MySQL UPDATE statement\n-- No primary key specified. Edit WHERE clause manually.\nUPDATE `%s`\nSET `%s` = '%s'\nWHERE <condition>;",
		tableName,
		columnName,
		escapedValue,
	)
}

func (m *MySQLConnection) BuildDeleteStatement(
	tableName, primaryKeyCol, pkValue string,
) string {
	escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")

	return fmt.Sprintf(
		"-- MySQL DELETE statement\n-- WARNING: This will permanently delete data!\n-- Ensure the WHERE clause is correct.\n\nDELETE FROM %s\nWHERE %s = '%s';",
		tableName,
		primaryKeyCol,
		escapedPkValue,
	)
}
