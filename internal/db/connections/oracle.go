package connections

import (
	"database/sql"
	"fmt"

	_ "github.com/godror/godror"
)

type OracleConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewOracleConnection(name, connStr string) (*OracleConnection, error) {
	bc := &BaseConnection{
		Name:       name,
		DbType:     "oracle",
		ConnString: connStr,
	}
	return &OracleConnection{BaseConnection: bc}, nil
}

func (oc *OracleConnection) Open() error {
	if oc.db != nil {
		return nil
	}
	db, err := sql.Open("godror", oc.ConnString)
	if err != nil {
		return err
	}
	oc.db = db
	return nil
}

func (oc *OracleConnection) Ping() error {
	if oc.db == nil {
		return fmt.Errorf("database is not open")
	}
	return oc.db.Ping()
}

func (oc *OracleConnection) Close() error {
	if oc.db != nil {
		return oc.db.Close()
	}
	return nil
}

func (oc *OracleConnection) Query(queryName string, args ...any) (any, error) {
	query, exists := oc.Queries[queryName]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", queryName)
	}
	return oc.db.Query(query.SQL, args...)
}

func (oc *OracleConnection) QueryDirect(sql string, args ...any) (any, error) {
	return oc.db.Query(sql, args...)
}

func (oc *OracleConnection) GetDB() *sql.DB {
	return oc.db
}

func (oc *OracleConnection) QueryTableWithLimit(tableName string, limit int) (*sql.Rows, error) {
	if oc.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	query := fmt.Sprintf("SELECT * FROM %s FETCH FIRST %d ROWS ONLY", tableName, limit)
	return oc.db.Query(query)
}

func (oc *OracleConnection) ListTables() ([]string, error) {
	if oc.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	rows, err := oc.db.Query("SELECT table_name FROM user_tables ORDER BY table_name")
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
