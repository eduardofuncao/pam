
package db

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
