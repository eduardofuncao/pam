package connections

import (
	"database/sql"

	"github.com/eduardofuncao/pam/internal/db/types"
)

type DatabaseConnection interface {
	Open() error
	Ping() error
	Close() error
	Query(queryName string, args ...any) (any, error)
	QueryDirect(sql string, args ...any) (any, error)
	QueryTableWithLimit(tableName string, limit int) (*sql.Rows, error)
	ListTables() ([]string, error)

	GetName() string
	GetDbType() string
	GetConnString() string
	GetQueries() map[string]types.Query
	GetDB() *sql.DB

	SetQueries(map[string]types.Query)
}

