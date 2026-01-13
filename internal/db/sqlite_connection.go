package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewSQLiteConnection(name, connStr string) (*SQLiteConnection, error) {
	bc := &BaseConnection{
		Name:       name,
		DbType:     "sqlite",
		ConnString: connStr,
	}
	return &SQLiteConnection{BaseConnection: bc}, nil
}

func (s *SQLiteConnection) Open() error {
	db, err := sql.Open("sqlite3", s. ConnString)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *SQLiteConnection) Ping() error {
	if s.db == nil {
		return fmt.Errorf("database is not open")
	}
	return s.db.Ping()
}

func (s *SQLiteConnection) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLiteConnection) Query(queryName string, args ...any) (any, error) {
	query, exists := s.Queries[queryName]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", queryName)
	}
	return s.db.Query(query.SQL, args...)
}

func (s *SQLiteConnection) ExecQuery(sql string, args ...any) (*sql.Rows, error) {
	return s.db.Query(sql, args...)
}

func (s *SQLiteConnection) Exec(sql string, args ...any) error {
	_, err := s.db.Exec(sql, args...)
	return err
}

func (s *SQLiteConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not open")
	}
	
	pkQuery := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	
	rows, err := s.db.Query(pkQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite table info: %w", err)
	}
	defer rows.Close()
	
	metadata := &TableMetadata{
		TableName: tableName,
	}
	
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString
		
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			continue
		}
		
		metadata.Columns = append(metadata. Columns, name)
		metadata.ColumnTypes = append(metadata.ColumnTypes, colType)  // ADD THIS
		
		if pk == 1 {
			metadata.PrimaryKey = name
		}
	}
	
	return metadata, nil
}

func (s *SQLiteConnection) GetInfoSQL(infoType string) string {
	switch infoType {
	case "tables":
		return `SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
		ORDER BY name`
	case "views":
		return `SELECT name
		FROM sqlite_master
		WHERE type = 'view'
		ORDER BY name`
	default:
		return ""
	}
}

func (s *SQLiteConnection) BuildUpdateStatement(tableName, columnName, currentValue, pkColumn, pkValue string) string {
	escapedValue := strings.ReplaceAll(currentValue, "'", "''")
	
	if pkColumn != "" && pkValue != "" {
		escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")
		return fmt.Sprintf(
			"-- SQLite UPDATE statement\nUPDATE %s\nSET %s = '%s'\nWHERE %s = '%s';",
			tableName,
			columnName,
			escapedValue,
			pkColumn,
			escapedPkValue,
		)
	}
	
	return fmt.Sprintf(
		"-- SQLite UPDATE statement\n-- No primary key specified. Edit WHERE clause manually.\nUPDATE %s\nSET %s = '%s'\nWHERE <condition>;",
		tableName,
		columnName,
		escapedValue,
	)
}

func (s *SQLiteConnection) BuildDeleteStatement(tableName, primaryKeyCol, pkValue string) string {
	escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")
	
	return fmt.Sprintf(
		"-- SQLite DELETE statement\n-- WARNING: This will permanently delete data!\n-- Ensure the WHERE clause is correct.\n\nDELETE FROM %s\nWHERE %s = '%s';",
		tableName,
		primaryKeyCol,
		escapedPkValue,
	)
}
