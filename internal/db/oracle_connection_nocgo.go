//go:build !cgo

package db

import (
	"database/sql"
	"fmt"
)

type OracleConnection struct {
	*BaseConnection
	db *sql.DB
}

func NewOracleConnection(name, connStr string) (*OracleConnection, error) {
	return nil, fmt.Errorf("Oracle driver not available: this binary was built without CGO support. Please use a build with CGO enabled or choose a different database")
}

func (oc *OracleConnection) Open() error {
	return fmt.Errorf("Oracle driver not available: binary built without CGO")
}

func (oc *OracleConnection) Ping() error {
	return fmt.Errorf("Oracle driver not available: binary built without CGO")
}

func (oc *OracleConnection) Close() error {
	return nil
}

func (oc *OracleConnection) Query(queryName string, args ...any) (any, error) {
	return nil, fmt.Errorf("Oracle driver not available: binary built without CGO")
}

func (oc *OracleConnection) ExecQuery(sql string, args ...any) (*sql.Rows, error) {
	return nil, fmt.Errorf("Oracle driver not available: binary built without CGO")
}

func (oc *OracleConnection) Exec(sql string, args ...any) error {
	return fmt.Errorf("Oracle driver not available: binary built without CGO")
}

func (oc *OracleConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	return nil, fmt.Errorf("Oracle driver not available: binary built without CGO")
}

func (oc *OracleConnection) GetInfoSQL(infoType string) string {
	return ""
}

func (oc *OracleConnection) BuildUpdateStatement(tableName, columnName, currentValue, pkColumn, pkValue string) string {
	return "-- Oracle driver not available: binary built without CGO"
}

func (oc *OracleConnection) ApplyRowLimit(sql string, limit int) string {
	return sql
}

func (oc *OracleConnection) BuildDeleteStatement(tableName, primaryKeyCol, pkValue string) string {
	return "-- Oracle driver not available: binary built without CGO"
}

func (oc *OracleConnection) GetPlaceholder(paramIndex int) string {
	return fmt.Sprintf(":%d", paramIndex)
}

func (oc *OracleConnection) GetTables() ([]string, error) {
	return nil, fmt.Errorf("Oracle driver not available: binary built without CGO")
}

func (oc *OracleConnection) GetViews() ([]string, error) {
	return nil, fmt.Errorf("Oracle driver not available: binary built without CGO")
}

func (oc *OracleConnection) GetForeignKeys(table string) ([]ForeignKey, error) {
	return nil, fmt.Errorf("Oracle driver not available: binary built without CGO")
}

func (oc *OracleConnection) GetForeignKeysReferencingTable(table string) ([]ForeignKey, error) {
	return nil, fmt.Errorf("Oracle driver not available: binary built without CGO")
}

func (oc *OracleConnection) GetUniqueConstraints(table string) ([]string, error) {
	return nil, fmt.Errorf("Oracle driver not available: binary built without CGO")
}

func (oc *OracleConnection) GetName() string {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.Name
	}
	return ""
}

func (oc *OracleConnection) GetDbType() string {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.DbType
	}
	return ""
}

func (oc *OracleConnection) GetConnString() string {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.ConnString
	}
	return ""
}

func (oc *OracleConnection) GetSchema() string {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.Schema
	}
	return ""
}

func (oc *OracleConnection) GetQueries() map[string]Query {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.Queries
	}
	return nil
}

func (oc *OracleConnection) GetLastQuery() Query {
	if oc.BaseConnection != nil {
		return oc.BaseConnection.LastQuery
	}
	return Query{}
}

func (oc *OracleConnection) SetSchema(schema string) {
	if oc.BaseConnection != nil {
		oc.BaseConnection.Schema = schema
	}
}

func (oc *OracleConnection) SetLastQuery(query Query) {
	if oc.BaseConnection != nil {
		oc.BaseConnection.LastQuery = query
	}
}

func (oc *OracleConnection) SetQueries(queries map[string]Query) {
	if oc.BaseConnection != nil {
		oc.BaseConnection.Queries = queries
	}
}
