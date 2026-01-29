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

func (p *PostgresConnection) ExecQuery(
	sql string,
	args ...any,
) (*sql.Rows, error) {
	return p.db.Query(sql, args...)
}

func (p *PostgresConnection) Exec(sql string, args ...any) error {
	_, err := p.db.Exec(sql, args...)
	return err
}

func (p *PostgresConnection) GetTableMetadata(
	tableName string,
) (*TableMetadata, error) {
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
	`

	rows, err := p.db.Query(pkQuery, tableName, currentSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to query postgres primary key: %w", err)
	}
	defer rows.Close()

	metadata := &TableMetadata{
		TableName: tableName,
	}

	for rows.Next() {
		var pkColumn string
		if err := rows.Scan(&pkColumn); err == nil {
			metadata.PrimaryKeys = append(metadata.PrimaryKeys, pkColumn)
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

	colRows, err := p.db.Query(colQuery, tableName, currentSchema)
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
	fks, err := p.GetForeignKeys(tableName)
	if err == nil {
		metadata.ForeignKeys = fks
	}

	return metadata, nil
}

func (p *PostgresConnection) GetInfoSQL(infoType string) string {
	schema := p.Schema
	if schema == "" {
		schema = "current_schema()"
	} else {
		schema = "'" + schema + "'"
	}

	switch infoType {
	case "tables":
		return fmt.Sprintf(`SELECT
			t.table_schema as schema,
			t.table_name as name,
			pg_get_userbyid(c.relowner) as owner
		FROM information_schema.tables t
		JOIN pg_class c ON c.relname = t.table_name
		WHERE t.table_schema = %s
		  AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_schema, t.table_name`, schema)
	case "views":
		return fmt.Sprintf(`SELECT
			table_schema as schema,
			table_name as name,
			pg_get_userbyid(c.relowner) as owner
		FROM information_schema.views v
		JOIN pg_class c ON c.relname = v.table_name
		WHERE table_schema = %s
		ORDER BY table_schema, table_name`, schema)
	default:
		return ""
	}
}

func (p *PostgresConnection) GetTables() ([]string, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var currentSchema string
	schemaQuery := `SELECT current_schema()`
	row := p.db.QueryRow(schemaQuery)
	if err := row.Scan(&currentSchema); err != nil {
		if p.Schema != "" {
			currentSchema = p.Schema
		} else {
			currentSchema = "public"
		}
	}

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := p.db.Query(query, currentSchema)
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

func (p *PostgresConnection) GetViews() ([]string, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var currentSchema string
	schemaQuery := `SELECT current_schema()`
	row := p.db.QueryRow(schemaQuery)
	if err := row.Scan(&currentSchema); err != nil {
		if p.Schema != "" {
			currentSchema = p.Schema
		} else {
			currentSchema = "public"
		}
	}

	query := `
		SELECT table_name
		FROM information_schema.views
		WHERE table_schema = $1
		ORDER BY table_name
	`

	rows, err := p.db.Query(query, currentSchema)
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

func (p *PostgresConnection) GetForeignKeys(tableName string) ([]ForeignKey, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var currentSchema string
	schemaQuery := `SELECT current_schema()`
	row := p.db.QueryRow(schemaQuery)
	if err := row.Scan(&currentSchema); err != nil {
		if p.Schema != "" {
			currentSchema = p.Schema
		} else {
			currentSchema = "public"
		}
	}

	query := `
		SELECT
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_name = $1
		  AND tc.table_schema = $2
		ORDER BY kcu.column_name
	`

	rows, err := p.db.Query(query, tableName, currentSchema)
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

func (p *PostgresConnection) GetForeignKeysReferencingTable(tableName string) ([]ForeignKey, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var currentSchema string
	schemaQuery := `SELECT current_schema()`
	row := p.db.QueryRow(schemaQuery)
	if err := row.Scan(&currentSchema); err != nil {
		if p.Schema != "" {
			currentSchema = p.Schema
		} else {
			currentSchema = "public"
		}
	}

	// Find all foreign keys that reference this table (reverse direction)
	query := `
		SELECT
			kcu.column_name,
			tc.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND ccu.table_name = $1
		  AND tc.table_schema = $2
		ORDER BY tc.table_name, kcu.column_name
	`

	rows, err := p.db.Query(query, tableName, currentSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to query referencing foreign keys: %w", err)
	}
	defer rows.Close()

	var foreignKeys []ForeignKey
	for rows.Next() {
		var fk ForeignKey
		// Note: We swap the meaning here
		// fk.Column = the FK column in the other table
		// fk.ReferencedTable = the other table (that has the FK)
		// fk.ReferencedColumn = the PK column in this table being referenced
		var otherColumn string
		if err := rows.Scan(&otherColumn, &fk.ReferencedTable, &fk.ReferencedColumn); err == nil {
			fk.Column = otherColumn
			foreignKeys = append(foreignKeys, fk)
		}
	}

	return foreignKeys, nil
}

func (p *PostgresConnection) GetUniqueConstraints(tableName string) ([]string, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var currentSchema string
	schemaQuery := `SELECT current_schema()`
	row := p.db.QueryRow(schemaQuery)
	if err := row.Scan(&currentSchema); err != nil {
		if p.Schema != "" {
			currentSchema = p.Schema
		} else {
			currentSchema = "public"
		}
	}

	query := `
		SELECT kcu.column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.constraint_type = 'UNIQUE'
		  AND tc.table_name = $1
		  AND tc.table_schema = $2
		ORDER BY kcu.column_name
	`

	rows, err := p.db.Query(query, tableName, currentSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to query unique constraints: %w", err)
	}
	defer rows.Close()

	var uniqueColumns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err == nil {
			uniqueColumns = append(uniqueColumns, column)
		}
	}

	return uniqueColumns, nil
}

func (p *PostgresConnection) BuildUpdateStatement(
	tableName, columnName, currentValue, pkColumn, pkValue string,
) string {
	escapedValue := strings.ReplaceAll(currentValue, "'", "''")

	if pkColumn != "" && pkValue != "" {
		escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")
		return fmt.Sprintf(
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

func (c *PostgresConnection) BuildDeleteStatement(
	tableName, primaryKeyCol, pkValue string,
) string {
	return fmt.Sprintf(
		"DELETE FROM %s\nWHERE %s = '%s';",
		tableName,
		primaryKeyCol,
		pkValue,
	)
}

func (p *PostgresConnection) GetPlaceholder(paramIndex int) string {
	return fmt.Sprintf("$%d", paramIndex)
}
