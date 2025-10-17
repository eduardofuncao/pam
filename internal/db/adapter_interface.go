package db

import "database/sql"

type DatabaseAdapter interface {
	Connect() error
	
	Close() error
	
	Ping() error
	
	Query(query string, args ...any) ([]string, [][]string, error)
	
	Exec(query string, args ...any) (sql.Result, error)
	
	GetConnectionString() string
	
	GetDB() *sql.DB
}

type DatabaseConfig struct {
	Name       string
	DBType     string
	Host       string
	Port       int
	Database   string
	Username   string
	Password   string
	ConnString string // Optional: use this if provided, otherwise build it
	Options    map[string]string
}


