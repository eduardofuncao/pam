//go:build cgo

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
	db, err := sql.Open("sqlite3", s.ConnString)
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

func (s *SQLiteConnection) ExecQuery(
	sql string,
	args ...any,
) (*sql.Rows, error) {
	return s.db.Query(sql, args...)
}

func (s *SQLiteConnection) Exec(sql string, args ...any) error {
	_, err := s.db.Exec(sql, args...)
	return err
}

func (s *SQLiteConnection) GetTableMetadata(
	tableName string,
) (*TableMetadata, error) {
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

		if err := rows.Scan(
			&cid,
			&name,
			&colType,
			&notNull,
			&dfltValue,
			&pk,
		); err != nil {
			continue
		}

		metadata.Columns = append(metadata.Columns, name)
		metadata.ColumnTypes = append(metadata.ColumnTypes, colType)

		if pk == 1 {
			metadata.PrimaryKeys = append(metadata.PrimaryKeys, name)
		}
	}

	// Fetch foreign keys
	fks, err := s.GetForeignKeys(tableName)
	if err == nil {
		metadata.ForeignKeys = fks
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

func (s *SQLiteConnection) GetTables() ([]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
		ORDER BY name
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

func (s *SQLiteConnection) GetViews() ([]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT name
		FROM sqlite_master
		WHERE type = 'view'
		ORDER BY name
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

func (s *SQLiteConnection) GetForeignKeys(tableName string) ([]ForeignKey, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	// Use PRAGMA foreign_key_list to get FK information
	query := fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	var foreignKeys []ForeignKey
	for rows.Next() {
		var id, seq int
		var table, from, to string
		var onUpdate, onDelete, match string

		if err := rows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match); err == nil {
			foreignKeys = append(foreignKeys, ForeignKey{
				Column:           from,
				ReferencedTable:  table,
				ReferencedColumn: to,
			})
		}
	}

	return foreignKeys, nil
}

func (s *SQLiteConnection) GetForeignKeysReferencingTable(tableName string) ([]ForeignKey, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	// Get all user tables
	tablesQuery := `
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`

	tablesRows, err := s.db.Query(tablesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer tablesRows.Close()

	var allTables []string
	for tablesRows.Next() {
		var tbl string
		if err := tablesRows.Scan(&tbl); err == nil {
			allTables = append(allTables, tbl)
		}
	}

	// Check each table for FKs referencing the target table
	var foreignKeys []ForeignKey
	for _, tbl := range allTables {
		query := fmt.Sprintf("PRAGMA foreign_key_list(%s)", tbl)
		rows, err := s.db.Query(query)
		if err != nil {
			continue
		}

		for rows.Next() {
			var id, seq int
			var table, from, to string
			var onUpdate, onDelete, match string

			if err := rows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match); err == nil {
				// Check if this FK references our target table
				if strings.EqualFold(table, tableName) {
					foreignKeys = append(foreignKeys, ForeignKey{
						Column:           from,
						ReferencedTable:  tbl,
						ReferencedColumn: to,
					})
				}
			}
		}
		rows.Close()
	}

	return foreignKeys, nil
}

func (s *SQLiteConnection) BuildUpdateStatement(
	tableName, columnName, currentValue, pkColumn, pkValue string,
) string {
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

func (s *SQLiteConnection) BuildDeleteStatement(
	tableName, primaryKeyCol, pkValue string,
) string {
	escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")

	return fmt.Sprintf(
		"-- SQLite DELETE statement\n-- WARNING: This will permanently delete data!\n-- Ensure the WHERE clause is correct.\n\nDELETE FROM %s\nWHERE %s = '%s';",
		tableName,
		primaryKeyCol,
		escapedPkValue,
	)
}

func (s *SQLiteConnection) GetPlaceholder(paramIndex int) string {
	return "?"
}
