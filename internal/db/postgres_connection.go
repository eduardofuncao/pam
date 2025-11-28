package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type PostgresConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewPostgresConnection(name, connStr string) (*PostgresConnection, error) {
	bc := &BaseConnection{
		Name:       name,
		DbType:     "postgres",
		ConnString: connStr,
	}
	return &PostgresConnection{BaseConnection: bc}, nil
}

func (p *PostgresConnection) Open() error {
	db, err := sql.Open("postgres", p.ConnString)
	if err != nil {
		return err
	}
	p.db = db
	return nil
}

func (oc *PostgresConnection) Ping() error {
	if oc.db == nil {
		return fmt.Errorf("database is not open")
	}
	return oc.db.Ping()
}

func (p *PostgresConnection) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *PostgresConnection) Query(queryName string, args ...any) (any, error) {
	query, exists := p.Queries[queryName]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", queryName)
	}
	return p.db.Query(query.SQL, args...)
}

func (p *PostgresConnection) Exec(sql string, args ...any) error {
	_, err := p.db. Exec(sql, args...)
	return err
}

func (p *PostgresConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database is not open")
	}
	
	pkQuery := `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = $1::regclass
		AND i.indisprimary
		ORDER BY a.attnum
		LIMIT 1
	`
	
	rows, err := p.db.Query(pkQuery, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query postgres primary key: %w", err)
	}
	defer rows.Close()
	
	metadata := &TableMetadata{
		TableName: tableName,
	}
	
	if rows.Next() {
		var pkColumn string
		if err := rows. Scan(&pkColumn); err == nil {
			metadata.PrimaryKey = pkColumn
		}
	}
	
	colQuery := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name = $1
		ORDER BY ordinal_position
	`
	
	colRows, err := p.db. Query(colQuery, tableName)
	if err == nil {
		defer colRows. Close()
		for colRows.Next() {
			var colName string
			if err := colRows.Scan(&colName); err == nil {
				metadata.Columns = append(metadata.Columns, colName)
			}
		}
	}
	
	return metadata, nil
}

func (p *PostgresConnection) BuildUpdateStatement(tableName, columnName, currentValue, pkColumn, pkValue string) string {
	escapedValue := strings. ReplaceAll(currentValue, "'", "''")
	
	if pkColumn != "" && pkValue != "" {
		escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")
		return fmt. Sprintf(
			"-- PostgreSQL UPDATE statement\nUPDATE %s\nSET %s = '%s'\nWHERE %s = '%s';",
			tableName,
			columnName,
			escapedValue,
			pkColumn,
			escapedPkValue,
		)
	}
	
	return fmt.Sprintf(
		"-- PostgreSQL UPDATE statement\n-- No primary key specified. Edit WHERE clause manually.\nUPDATE %s\nSET %s = '%s'\nWHERE <condition>;",
		tableName,
		columnName,
		escapedValue,
	)
}
