package db

import "database/sql"

type DatabaseConnection interface {
	Open() error
	Ping() error
	Close() error
	Query(queryName string, args ...any) (any, error)
	ExecQuery(sql string, args ...any) (*sql.Rows, error)
	Exec(sql string, args ...any) error
	GetTableMetadata(tableName string) (*TableMetadata, error)
	GetInfoSQL(infoType string) string
	BuildUpdateStatement(tableName, columnName, currentValue, pkColumn, pkValue string) string
	BuildDeleteStatement(tableName, primaryKeyCol, pkValue string) string
	ApplyRowLimit(sql string, limit int) string

	GetName() string
	GetDbType() string
	GetConnString() string
	GetSchema() string
	GetQueries() map[string]Query
	GetLastQuery() Query

	SetSchema(string)
	SetLastQuery(Query)
	SetQueries(map[string]Query)
}
