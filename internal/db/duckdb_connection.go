//go:build cgo && !noduckdb

package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/duckdb/duckdb-go/v2"
)

func init() {
	registerDriver(func(name, connStr string) (DatabaseConnection, error) {
		return NewDuckDBConnection(name, connStr)
	}, "duckdb")
}

type DuckDBConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewDuckDBConnection(name, connStr string) (*DuckDBConnection, error) {
	bc := &BaseConnection{
		Name:       name,
		DbType:     "duckdb",
		ConnString: connStr,
	}
	return &DuckDBConnection{BaseConnection: bc}, nil
}

func (d *DuckDBConnection) Open() error {
	db, err := sql.Open("duckdb", d.ConnString)
	if err != nil {
		return fmt.Errorf("failed to open duckdb database: %w", err)
	}
	d.db = db
	return nil
}

func (d *DuckDBConnection) Ping() error {
	if d.db == nil {
		return fmt.Errorf("database is not open")
	}
	return d.db.Ping()
}

func (d *DuckDBConnection) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *DuckDBConnection) Query(queryName string, args ...any) (any, error) {
	query, exists := d.Queries[queryName]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", queryName)
	}
	return d.db.Query(query.SQL, args...)
}

func (d *DuckDBConnection) ExecQuery(sql string, args ...any) (*sql.Rows, error) {
	return d.db.Query(sql, args...)
}

func (d *DuckDBConnection) Exec(sql string, args ...any) (sql.Result, error) {
	return d.db.Exec(sql, args...)
}

func (d *DuckDBConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	metadata := &TableMetadata{
		TableName: tableName,
	}

	colQuery := `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = ?
		ORDER BY ordinal_position
	`

	colRows, err := d.db.Query(colQuery, tableName)
	if err != nil {
		return metadata, fmt.Errorf("failed to query duckdb column metadata: %w", err)
	}
	defer colRows.Close()

	for colRows.Next() {
		var colName, colType string
		if err := colRows.Scan(&colName, &colType); err != nil {
			continue
		}
		metadata.Columns = append(metadata.Columns, colName)
		metadata.ColumnTypes = append(metadata.ColumnTypes, colType)
	}

	pkQuery := `
		SELECT constraint_column_names::VARCHAR
		FROM duckdb_constraints()
		WHERE table_name = ?
		  AND constraint_type = 'PRIMARY KEY'
	`
	pkRows, err := d.db.Query(pkQuery, tableName)
	if err == nil {
		defer pkRows.Close()
		for pkRows.Next() {
			var colNames string
			if err := pkRows.Scan(&colNames); err == nil {
				metadata.PrimaryKeys = append(metadata.PrimaryKeys, parseDuckDBArray(colNames)...)
			}
		}
	}

	fks, err := d.GetForeignKeys(tableName)
	if err == nil {
		metadata.ForeignKeys = fks
	}

	return metadata, nil
}

func (d *DuckDBConnection) GetInfoSQL(infoType string) string {
	switch infoType {
	case "tables":
		return `SELECT table_schema as schema,
		       table_name as name
		FROM information_schema.tables
		WHERE table_type = 'BASE TABLE'
		  AND table_schema NOT IN ('information_schema', 'pg_catalog')
		ORDER BY table_schema, table_name`
	case "views":
		return `SELECT table_schema as schema,
		       table_name as name
		FROM information_schema.views
		WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
		ORDER BY table_schema, table_name`
	default:
		return ""
	}
}

func (d *DuckDBConnection) GetTables() ([]string, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_type = 'BASE TABLE'
		  AND table_schema NOT IN ('information_schema', 'pg_catalog')
		ORDER BY table_name
	`

	rows, err := d.db.Query(query)
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

func (d *DuckDBConnection) GetViews() ([]string, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT table_name
		FROM information_schema.views
		WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
		ORDER BY table_name
	`

	rows, err := d.db.Query(query)
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

func (d *DuckDBConnection) GetForeignKeys(tableName string) ([]ForeignKey, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT constraint_column_names::VARCHAR,
		       referenced_table,
		       referenced_column_names::VARCHAR
		FROM duckdb_constraints()
		WHERE table_name = ?
		  AND constraint_type = 'FOREIGN KEY'
	`

	rows, err := d.db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	var foreignKeys []ForeignKey
	for rows.Next() {
		var cols, refTable, refCols string
		if err := rows.Scan(&cols, &refTable, &refCols); err == nil {
			srcCols := parseDuckDBArray(cols)
			dstCols := parseDuckDBArray(refCols)
			for i := range srcCols {
				fk := ForeignKey{
					Column:          srcCols[i],
					ReferencedTable: refTable,
				}
				if i < len(dstCols) {
					fk.ReferencedColumn = dstCols[i]
				}
				foreignKeys = append(foreignKeys, fk)
			}
		}
	}

	return foreignKeys, nil
}

func (d *DuckDBConnection) GetForeignKeysReferencingTable(tableName string) ([]ForeignKey, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT table_name,
		       constraint_column_names::VARCHAR,
		       referenced_column_names::VARCHAR
		FROM duckdb_constraints()
		WHERE referenced_table = ?
		  AND constraint_type = 'FOREIGN KEY'
	`

	rows, err := d.db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query referencing foreign keys: %w", err)
	}
	defer rows.Close()

	var foreignKeys []ForeignKey
	for rows.Next() {
		var srcTable, srcCols, refCols string
		if err := rows.Scan(&srcTable, &srcCols, &refCols); err == nil {
			sCols := parseDuckDBArray(srcCols)
			dCols := parseDuckDBArray(refCols)
			for i := range sCols {
				fk := ForeignKey{
					Column:          sCols[i],
					ReferencedTable: srcTable,
				}
				if i < len(dCols) {
					fk.ReferencedColumn = dCols[i]
				}
				foreignKeys = append(foreignKeys, fk)
			}
		}
	}

	return foreignKeys, nil
}

func (d *DuckDBConnection) GetUniqueConstraints(tableName string) ([]string, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT constraint_column_names::VARCHAR
		FROM duckdb_constraints()
		WHERE table_name = ?
		  AND constraint_type = 'UNIQUE'
	`

	rows, err := d.db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query unique constraints: %w", err)
	}
	defer rows.Close()

	var uniqueColumns []string
	for rows.Next() {
		var cols string
		if err := rows.Scan(&cols); err == nil {
			uniqueColumns = append(uniqueColumns, parseDuckDBArray(cols)...)
		}
	}

	return uniqueColumns, nil
}

// parseDuckDBArray parses DuckDB VARCHAR array output like "[col1, col2]" or "[col1]"
func parseDuckDBArray(s string) []string {
	s = strings.Trim(s, "[]")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
