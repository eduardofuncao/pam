package db
//
// import (
// 	"database/sql"
// 	"fmt"
// 	"strings"
//
// 	_ "github.com/marcboeker/go-duckdb"
// )
//
// type DuckDBConnection struct {
// 	*BaseConnection
// 	db *sql.DB
// }
//
// func NewDuckDBConnection(name, connStr string) (*DuckDBConnection, error) {
// 	bc := &BaseConnection{
// 		Name:       name,
// 		DbType:     "duckdb",
// 		ConnString: connStr,
// 	}
// 	return &DuckDBConnection{BaseConnection: bc}, nil
// }
//
// func (d *DuckDBConnection) Open() error {
// 	db, err := sql.Open("duckdb", d.ConnString)
// 	if err != nil {
// 		return fmt.Errorf("failed to open duckdb database: %w", err)
// 	}
// 	d.db = db
// 	return nil
// }
//
// func (d *DuckDBConnection) Ping() error {
// 	if d.db == nil {
// 		return fmt.Errorf("database is not open")
// 	}
// 	return d.db.Ping()
// }
//
// func (d *DuckDBConnection) Close() error {
// 	if d.db != nil {
// 		return d.db.Close()
// 	}
// 	return nil
// }
//
// func (d *DuckDBConnection) Query(queryName string, args ...any) (any, error) {
// 	query, exists := d.Queries[queryName]
// 	if !exists {
// 		return nil, fmt.Errorf("query not found: %s", queryName)
// 	}
// 	return d.db.Query(query.SQL, args...)
// }
//
// func (d *DuckDBConnection) ExecQuery(sql string, args ...any) (*sql.Rows, error) {
// 	return d.db.Query(sql, args...)
// }
//
// func (d *DuckDBConnection) Exec(sql string, args ...any) error {
// 	_, err := d.db.Exec(sql, args...)
// 	return err
// }
//
// func (d *DuckDBConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
// 	if d.db == nil {
// 		return nil, fmt.Errorf("database is not open")
// 	}
//
// 	metadata := &TableMetadata{
// 		TableName: tableName,
// 	}
//
// 	pkQuery := `
// 		SELECT column_name
// 		FROM information_schema.key_column_usage
// 		WHERE table_name = ?
// 		LIMIT 1
// 	`
//
// 	pkRows, err := d.db.Query(pkQuery, tableName)
// 	if err == nil {
// 		defer pkRows.Close()
// 		if pkRows.Next() {
// 			var pkColumn string
// 			if err := pkRows.Scan(&pkColumn); err == nil {
// 				metadata.PrimaryKey = pkColumn
// 			}
// 		}
// 	}
//
// 	colQuery := `
// 		SELECT column_name, data_type
// 		FROM information_schema.columns
// 		WHERE table_name = ?
// 		ORDER BY ordinal_position
// 	`
//
// 	colRows, err := d.db.Query(colQuery, tableName)
// 	if err != nil {
// 		return metadata, fmt.Errorf("failed to query duckdb column metadata: %w", err)
// 	}
// 	defer colRows.Close()
//
// 	for colRows.Next() {
// 		var colName, colType string
// 		if err := colRows.Scan(&colName, &colType); err != nil {
// 			continue
// 		}
// 		metadata.Columns = append(metadata.Columns, colName)
// 		metadata.ColumnTypes = append(metadata.ColumnTypes, colType)
// 	}
//
// 	return metadata, nil
// }
//
// func (d *DuckDBConnection) GetInfoSQL(infoType string) string {
// 	switch infoType {
// 	case "tables":
// 		return `SELECT table_schema as schema,
// 		       table_name as name
// 		FROM information_schema.tables
// 		WHERE table_type = 'BASE TABLE'
// 		ORDER BY table_schema, table_name`
// 	case "views":
// 		return `SELECT table_schema as schema,
// 		       table_name as name
// 		FROM information_schema.views
// 		ORDER BY table_schema, table_name`
// 	default:
// 		return ""
// 	}
// }
// func (d *DuckDBConnection) BuildUpdateStatement(tableName, columnName, currentValue, pkColumn, pkValue string) string {
// 	escapedValue := strings.ReplaceAll(currentValue, "'", "''")
//
// 	if pkColumn != "" && pkValue != "" {
// 		escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")
// 		return fmt.Sprintf(`-- DuckDB UPDATE statement
// UPDATE %s
// SET %s = '%s'
// WHERE %s = '%s';`,
// 			tableName,
// 			columnName,
// 			escapedValue,
// 			pkColumn,
// 			escapedPkValue,
// 		)
// 	}
//
// 	return fmt.Sprintf(`-- DuckDB UPDATE statement
// -- No primary key specified. Edit WHERE clause manually.
// UPDATE %s
// SET %s = '%s'
// WHERE <condition>;`,
// 		tableName,
// 		columnName,
// 		escapedValue,
// 	)
// }
//
// func (d *DuckDBConnection) BuildDeleteStatement(tableName, primaryKeyCol, pkValue string) string {
// 	escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")
//
// 	return fmt.Sprintf(`-- DuckDB DELETE statement
// -- WARNING: This will permanently delete data!
// -- Ensure the WHERE clause is correct.
//
// DELETE FROM %s
// WHERE %s = '%s';`,
// 		tableName,
// 		primaryKeyCol,
// 		escapedPkValue,
// 	)
// }
//
