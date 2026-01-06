package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type BaseConnection struct {
	Name       string
	DbType     string
	ConnString string
	Schema     string
	Queries    map[string]Query
	LastQuery  Query
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
	return struct{}{}, errors.New("Close() not implemented for base connection")
}
func (b *BaseConnection) ExecQuery(sql string, args ...any) (*sql.Rows, error) {
	return nil, errors.New("ExecQuery() not implemented for base connection")
}
func (b *BaseConnection) Exec(sql string, args ...any) error {
	return errors. New("Exec() not implemented for base connection")
}
func (b *BaseConnection) GetTableMetadata(tableName string) (*TableMetadata, error) {
	return nil, errors. New("GetTableMetadata() not implemented for base connection")
}

func (b *BaseConnection) BuildUpdateStatement(tableName, columnName, currentValue, pkColumn, pkValue string) string {
	escapedValue := strings.ReplaceAll(currentValue, "'", "''")
	
	if pkColumn != "" && pkValue != "" {
		escapedPkValue := strings.ReplaceAll(pkValue, "'", "''")
		return fmt.Sprintf(
			"UPDATE %s\nSET %s = '%s'\nWHERE %s = '%s';",
			tableName,
			columnName,
			escapedValue,
			pkColumn,
			escapedPkValue,
		)
	}
	
	return fmt.Sprintf(
		"-- No primary key specified. Edit WHERE clause manually.\nUPDATE %s\nSET %s = '%s'\nWHERE <condition>;",
		tableName,
		columnName,
		escapedValue,
	)
}

func (b *BaseConnection) GetName() string                     { return b.Name }
func (b *BaseConnection) GetDbType() string                   { return b.DbType }
func (b *BaseConnection) GetConnString() string               { return b.ConnString }
func (b *BaseConnection) GetLastQuery() Query                 { return b.LastQuery }
func (b *BaseConnection) SetLastQuery(query Query)            { b.LastQuery = query }
func (b *BaseConnection) GetQueries() map[string]Query        { return b.Queries }
func (b *BaseConnection) SetQueries(queries map[string]Query) { b.Queries = queries }
func (b *BaseConnection) GetSchema() string        { return b.Schema }
func (b *BaseConnection) SetSchema(schema string)  { b.Schema = schema }
