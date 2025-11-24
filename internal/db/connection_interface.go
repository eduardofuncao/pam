package db

import "database/sql"

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
	GetDB() *sql.DB

	SetQueries(map[string]Query)
}

