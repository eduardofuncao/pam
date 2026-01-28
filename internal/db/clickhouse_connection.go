package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewClickHouseConnection(name, connStr string) (*ClickHouseConnection, error) {
	bc := &BaseConnection{
		Name:       name,
		DbType:     "clickhouse",
		ConnString: connStr,
	}

	// Parse database from URL path if schema not already set
	// clickhouse://user:pass@host:9000/database
	parsedURL, err := url.Parse(connStr)
	if err == nil && parsedURL.Path != "" && parsedURL.Path != "/" {
		// Extract database from path (remove leading slash)
		database := strings.TrimPrefix(parsedURL.Path, "/")
		if database != "" && bc.Schema == "" {
			bc.Schema = database
		}
	}

	return &ClickHouseConnection{BaseConnection: bc}, nil
}

func (c *ClickHouseConnection) Open() error {
	db, err := sql.Open("clickhouse", c.ConnString)
	if err != nil {
		return err
	}
	c.db = db

	if c.Schema != "" {
		setDatabaseSQL := fmt.Sprintf("USE DATABASE %s", c.Schema)
		_, err = c.db.Exec(setDatabaseSQL)
		if err != nil {
			c.db.Close()
			return fmt.Errorf("failed to set database to '%s': %w", c.Schema, err)
		}
	}

	return nil
}

func (c *ClickHouseConnection) Ping() error {
	if c.db == nil {
		return fmt.Errorf("database is not open")
	}
	return c.db.Ping()
}

func (c *ClickHouseConnection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

func (c *ClickHouseConnection) Query(queryName string, args ...any) (any, error) {
	query, exists := c.Queries[queryName]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", queryName)
	}
	return c.db.Query(query.SQL, args...)
}

func (c *ClickHouseConnection) ExecQuery(sql string, args ...any) (*sql.Rows, error) {
	return c.db.Query(sql, args...)
}

func (c *ClickHouseConnection) Exec(sql string, args ...any) error {
	_, err := c.db.Exec(sql, args...)
	return err
}

func (c *ClickHouseConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	metadata := &TableMetadata{
		TableName: tableName,
	}

	// Query for primary key directly from system.tables
	pkQuery := `
		SELECT primary_key
		FROM system.tables
		WHERE name = ?
		  AND database = currentDatabase()
		LIMIT 1
	`

	row := c.db.QueryRow(pkQuery, tableName)
	var primaryKey string
	if err := row.Scan(&primaryKey); err == nil && primaryKey != "" {
		// Extract first column if comma-separated (for composite PKs)
		keys := strings.Split(primaryKey, ",")
		if len(keys) > 0 {
			// Remove quotes and trim whitespace
			pk := strings.TrimSpace(keys[0])
			pk = strings.Trim(pk, "`\"'")
			metadata.PrimaryKeys = append(metadata.PrimaryKeys, pk)
		}
	}

	// Query for column metadata
	colQuery := `
		SELECT name, type
		FROM system.columns
		WHERE table = ?
		  AND database = currentDatabase()
		ORDER BY position
	`

	colRows, err := c.db.Query(colQuery, tableName)
	if err != nil {
		return metadata, fmt.Errorf("failed to query clickhouse column metadata: %w", err)
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

	metadata.ForeignKeys = []ForeignKey{}

	return metadata, nil
}

func (c *ClickHouseConnection) GetForeignKeys(tableName string) ([]ForeignKey, error) {
	// Return empty list gracefully
	return []ForeignKey{}, nil
}

func (c *ClickHouseConnection) GetForeignKeysReferencingTable(tableName string) ([]ForeignKey, error) {
	// Return empty list gracefully
	return []ForeignKey{}, nil
}

func (c *ClickHouseConnection) GetInfoSQL(infoType string) string {
	database := c.Schema
	if database == "" {
		database = "currentDatabase()"
	} else {
		database = "'" + database + "'"
	}

	switch infoType {
	case "tables":
		return fmt.Sprintf(`SELECT database as schema,
		       name,
		       engine as owner
		FROM system.tables
		WHERE database = %s
		  AND engine != 'View'
		ORDER BY database, name`, database)
	case "views":
		return fmt.Sprintf(`SELECT database as schema,
		       name,
		       engine as owner
		FROM system.tables
		WHERE database = %s
		  AND engine = 'View'
		ORDER BY database, name`, database)
	default:
		return ""
	}
}

func (c *ClickHouseConnection) GetTables() ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT name
		FROM system.tables
		WHERE database = currentDatabase()
		  AND engine != 'View'
		ORDER BY name
	`

	rows, err := c.db.Query(query)
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

func (c *ClickHouseConnection) GetViews() ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not open")
	}

	query := `
		SELECT name
		FROM system.tables
		WHERE database = currentDatabase()
		  AND engine = 'View'
		ORDER BY name
	`

	rows, err := c.db.Query(query)
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

func (c *ClickHouseConnection) BuildUpdateStatement(tableName, columnName, currentValue, pkColumn, pkValue string) string {
	escapedValue := strings.ReplaceAll(currentValue, "'", "''")

	if pkColumn != "" && pkValue != "" {
		escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")
		return fmt.Sprintf(`-- ClickHouse UPDATE statement
-- Note: ClickHouse uses ALTER TABLE UPDATE for mutations
ALTER TABLE %s
UPDATE %s = '%s'
WHERE %s = '%s';`,
			tableName,
			columnName,
			escapedValue,
			pkColumn,
			escapedPkValue,
		)
	}

	return fmt.Sprintf(`-- ClickHouse UPDATE statement
-- No primary key specified. Edit WHERE clause manually.
-- Note: ClickHouse uses ALTER TABLE UPDATE for mutations
ALTER TABLE %s
UPDATE %s = '%s'
WHERE <condition>;`,
		tableName,
		columnName,
		escapedValue,
	)
}

func (c *ClickHouseConnection) BuildDeleteStatement(tableName, primaryKeyCol, pkValue string) string {
	escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")

	return fmt.Sprintf(`-- ClickHouse DELETE statement
-- WARNING: This will permanently delete data!
-- Note: ClickHouse uses ALTER TABLE DELETE for mutations
-- Ensure the WHERE clause is correct.

ALTER TABLE %s
DELETE
WHERE %s = '%s';`,
		tableName,
		primaryKeyCol,
		escapedPkValue,
	)
}

func (c *ClickHouseConnection) GetPlaceholder(paramIndex int) string {
	return "?"
}

func (c *ClickHouseConnection) ApplyRowLimit(sql string, limit int) string {
	// ClickHouse uses standard SQL LIMIT syntax
	trimmedSQL := strings.ToUpper(strings.TrimSpace(sql))

	// Only apply to SELECT statements
	if !strings.HasPrefix(trimmedSQL, "SELECT") {
		return sql
	}

	// Don't modify if already has LIMIT
	if strings.Contains(strings.ToUpper(sql), " LIMIT ") {
		return sql
	}

	return fmt.Sprintf("%s\nLIMIT %d", strings.TrimRight(sql, ";"), limit)
}
