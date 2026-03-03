//go:build !cgo

package db

import (
	"database/sql"
	"fmt"
)

type SQLiteConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewSQLiteConnection(name, connStr string) (*SQLiteConnection, error) {
	return nil, fmt.Errorf("SQLite driver not available: this binary was built without CGO support. Please use a build with CGO enabled or choose a different database")
}

func (oc *SQLiteConnection) Open() error {
	return fmt.Errorf("SQLite driver not available: binary built without CGO")
}

func (oc *SQLiteConnection) Ping() error {
	return fmt.Errorf("SQLite driver not available: binary built without CGO")
}

func (oc *SQLiteConnection) Close() error {
	return nil
}

func (oc *SQLiteConnection) Query(queryName string, args ...any) (any, error) {
	return nil, fmt.Errorf("SQLite driver not available: binary built without CGO")
}

func (oc *SQLiteConnection) ExecQuery(sql string, args ...any) (*sql.Rows, error) {
	return nil, fmt.Errorf("SQLite driver not available: binary built without CGO")
}

func (oc *SQLiteConnection) Exec(sql string, args ...any) error {
	return fmt.Errorf("SQLite driver not available: binary built without CGO")
}

func (oc *SQLiteConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	return nil, fmt.Errorf("SQLite driver not available: binary built without CGO")
}

func (oc *SQLiteConnection) GetInfoSQL(infoType string) string {
	return ""
}

func (oc *SQLiteConnection) BuildUpdateStatement(tableName, columnName, currentValue, pkColumn, pkValue string) string {
	return "-- SQLite driver not available: binary built without CGO"
}

func (oc *SQLiteConnection) ApplyRowLimit(sql string, limit int) string {
	return sql
}

func (oc *SQLiteConnection) BuildDeleteStatement(tableName, primaryKeyCol, pkValue string) string {
	return "-- SQLite driver not available: binary built without CGO"
}

func (oc *SQLiteConnection) GetPlaceholder(paramIndex int) string {
	return "?"
}

func (oc *SQLiteConnection) GetTables() ([]string, error) {
	return nil, fmt.Errorf("SQLite driver not available: binary built without CGO")
}

func (oc *SQLiteConnection) GetViews() ([]string, error) {
	return nil, fmt.Errorf("SQLite driver not available: binary built without CGO")
}

func (oc *SQLiteConnection) GetForeignKeys(table string) ([]ForeignKey, error) {
	return nil, fmt.Errorf("SQLite driver not available: binary built without CGO")
}

func (oc *SQLiteConnection) GetForeignKeysReferencingTable(table string) ([]ForeignKey, error) {
	return nil, fmt.Errorf("SQLite driver not available: binary built without CGO")
}

func (oc *SQLiteConnection) GetUniqueConstraints(table string) ([]string, error) {
	return nil, fmt.Errorf("SQLite driver not available: binary built without CGO")
}

func (oc *SQLiteConnection) GetName() string {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.Name
	}
	return ""
}

func (oc *SQLiteConnection) GetDbType() string {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.DbType
	}
	return ""
}

func (oc *SQLiteConnection) GetConnString() string {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.ConnString
	}
	return ""
}

func (oc *SQLiteConnection) GetSchema() string {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.Schema
	}
	return ""
}

func (oc *SQLiteConnection) GetQueries() map[string]Query {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.Queries
	}
	return nil
}

func (oc *SQLiteConnection) GetLastQuery() Query {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.LastQuery
	}
	return Query{}
}

func (oc *SQLiteConnection) SetSchema(schema string) {
	if oc.BaseConnection != nil {
		oc.BaseConnection.Schema = schema
	}
}

func (oc *SQLiteConnection) SetLastQuery(query Query) {
	if oc.BaseConnection != nil {
		oc.BaseConnection.LastQuery = query
	}
}

func (oc *SQLiteConnection) SetQueries(queries map[string]Query) {
	if oc.BaseConnection != nil {
		oc.BaseConnection.Queries = queries
	}
}
