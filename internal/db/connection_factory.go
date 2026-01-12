package db

import (
	"fmt"
)

func CreateConnection(name, dbType, connString string) (DatabaseConnection, error) {
	switch dbType {
	case "postgres", "postgresql":
		return NewPostgresConnection(name, connString)
	case "mysql", "mariadb":
		return NewMySQLConnection(name, connString)
	case "sqlite", "sqlite3":
		return NewSQLiteConnection(name, connString)
	case "sqlserver", "mssql":
		return NewSQLServerConnection(name, connString)
	case "duckdb":
		return NewDuckDBConnection(name, connString)
	case "godror", "oracle":
		return NewOracleConnection(name, connString)
	default:
		return nil, fmt.Errorf("driver not implemented for %s", dbType)
	}
}
