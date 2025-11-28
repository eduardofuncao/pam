package db

type DatabaseConnection interface {
	Open() error
	Ping() error
	Close() error
	Query(queryName string, args ...any) (any, error)
	Exec(sql string, args ...any) error
	GetTableMetadata(tableName string) (*TableMetadata, error)
	BuildUpdateStatement(tableName, columnName, currentValue, pkColumn, pkValue string) string

	GetName() string
	GetDbType() string
	GetConnString() string
	GetQueries() map[string]Query
	GetLastQuery() Query

	SetLastQuery(Query)
	SetQueries(map[string]Query)
}
