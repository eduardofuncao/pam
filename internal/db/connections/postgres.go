package connections

import (
	"database/sql"
	"fmt"

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

func (p *PostgresConnection) QueryDirect(sql string, args ...any) (any, error) {
	return p.db.Query(sql, args...)
}

func (p *PostgresConnection) GetDB() *sql.DB {
	return p.db
}

func (p *PostgresConnection) QueryTableWithLimit(tableName string, limit int) (*sql.Rows, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d", tableName, limit)
	return p.db.Query(query)
}

func (p *PostgresConnection) ListTables() ([]string, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	rows, err := p.db.Query("SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}
