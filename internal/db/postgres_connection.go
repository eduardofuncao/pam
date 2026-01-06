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
	
	if p.Schema != "" {
		setSchemaSQL := fmt.Sprintf("SET search_path TO %s", p.Schema)
		_, err = p.db.Exec(setSchemaSQL)
		if err != nil {
			p.db.Close()
			return fmt.Errorf("failed to set schema to '%s': %w", p.Schema, err)
		}
	}
	
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

func (p *PostgresConnection) ExecQuery(sql string, args ...any) (*sql.Rows, error) {
	return p.db.Query(sql, args...)
}

func (p *PostgresConnection) Exec(sql string, args ...any) error {
	_, err := p.db. Exec(sql, args...)
	return err
}

func (p *PostgresConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database is not open")
	}
	
	var currentSchema string
	schemaQuery := `SELECT current_schema()`
	row := p.db.QueryRow(schemaQuery)
	if err := row.Scan(&currentSchema); err != nil {
		// Fallback to configured schema or 'public'
		if p.Schema != "" {
			currentSchema = p.Schema
		} else {
			currentSchema = "public"
		}
	}
	
	pkQuery := `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		JOIN pg_class c ON c. oid = i.indrelid
		JOIN pg_namespace n ON n. oid = c.relnamespace
		WHERE c.relname = $1
		AND n.nspname = $2
		AND i.indisprimary
		ORDER BY a.attnum
		LIMIT 1
	`
	
	rows, err := p.db.Query(pkQuery, tableName, currentSchema)
	if err != nil {
		return nil, fmt. Errorf("failed to query postgres primary key: %w", err)
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
		SELECT column_name, 
		       CASE 
		           WHEN character_maximum_length IS NOT NULL 
		           THEN data_type || '(' || character_maximum_length || ')'
		           WHEN numeric_precision IS NOT NULL 
		           THEN data_type || '(' || numeric_precision || ',' || numeric_scale || ')'
		           ELSE data_type
		       END as full_type
		FROM information_schema.columns
		WHERE table_name = $1
		AND table_schema = $2
		ORDER BY ordinal_position
	`
	
	colRows, err := p.db. Query(colQuery, tableName, currentSchema)
	if err == nil {
		defer colRows. Close()
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

func (c *PostgresConnection) BuildDeleteStatement(tableName, primaryKeyCol, pkValue string) string {
	return fmt.Sprintf(
		"DELETE FROM %s\nWHERE %s = '%s';",
		tableName,
		primaryKeyCol,
		pkValue,
	)
}
