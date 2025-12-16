package connections

import (
	"database/sql"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"
)

type SqlServerConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewSqlServerConnection(name, connStr string) (*SqlServerConnection, error) {
	bc := &BaseConnection{
		Name:       name,
		DbType:     "sqlserver",
		ConnString: connStr,
	}
	return &SqlServerConnection{BaseConnection: bc}, nil
}

func (s *SqlServerConnection) Open() error {
	db, err := sql.Open("sqlserver", s.ConnString)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *SqlServerConnection) Ping() error {
	if s.db == nil {
		return fmt.Errorf("database is not open")
	}
	return s.db.Ping()
}

func (s *SqlServerConnection) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SqlServerConnection) Query(queryName string, args ...any) (any, error) {
	query, exists := s.Queries[queryName]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", queryName)
	}
	return s.db.Query(query.SQL, args...)
}

func (s *SqlServerConnection) QueryDirect(sql string, args ...any) (any, error) {
	return s.db.Query(sql, args...)
}

func (s *SqlServerConnection) GetDB() *sql.DB {
	return s.db
}

func (s *SqlServerConnection) QueryTableWithLimit(tableName string, limit int) (*sql.Rows, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	query := fmt.Sprintf("SELECT TOP %d * FROM %s", limit, tableName)
	return s.db.Query(query)
}

func (s *SqlServerConnection) ListTables() ([]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	rows, err := s.db.Query("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME")
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
