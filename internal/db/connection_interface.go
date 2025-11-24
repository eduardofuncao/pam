package db

type DatabaseConnection interface {
	Open() error
	Ping() error
	Close() error
	Query(queryName string, args ...any) (any, error)
	QueryDirect(sql string, args ...any) (any, error)

	GetName() string
	GetDbType() string
	GetConnString() string
	GetQueries() map[string]Query

	SetQueries(map[string]Query)
}

