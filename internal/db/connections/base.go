package connections

import (
	"database/sql"
	"errors"

	"github.com/eduardofuncao/pam/internal/db/types"
)

type BaseConnection struct {
	Name       string
	DbType     string
	ConnString string
	Queries    map[string]types.Query
}

func (b *BaseConnection) Open() error {
	return errors.New("Open() not implemented for base connection")
}
func (b *BaseConnection) Ping() error {
	return errors.New("Ping() not implemented for base connection")
}
func (b *BaseConnection) Close() error {
	return errors.New("Close() not implemented for base connection")
}
func (b *BaseConnection) Query(name string, args ...any) (any, error) {
	return struct{}{}, errors.New("Query() not implemented for base connection")
}
func (b *BaseConnection) QueryDirect(sql string, args ...any) (any, error) {
	return struct{}{}, errors.New("QueryDirect() not implemented for base connection")
}
func (b *BaseConnection) QueryTableWithLimit(tableName string, limit int) (*sql.Rows, error) {
	return nil, errors.New("QueryTableWithLimit() not implemented for base connection")
}
func (b *BaseConnection) ListTables() ([]string, error) {
	return nil, errors.New("ListTables() not implemented for base connection")
}

func (b *BaseConnection) GetName() string                     { return b.Name }
func (b *BaseConnection) GetDbType() string                   { return b.DbType }
func (b *BaseConnection) GetConnString() string               { return b.ConnString }
func (b *BaseConnection) GetQueries() map[string]types.Query        { return b.Queries }
func (b *BaseConnection) SetQueries(queries map[string]types.Query) { b.Queries = queries }
