package db

import(
	"errors"
	"fmt"
)

func CreateConnection(name, dbType, connString string) (DatabaseConnection, error) {
	switch dbType{
	case "postgres":
		return NewPostgresConnection(name, connString)
	case "sqlserver", "mssql":
		return NewSqlServerConnection(name, connString)
	case "myslq", "mariadb":
		return nil, errors.New("mysql driver not implemented, check connection factory")
	case "sqlite", "sqlite3":
		return nil, errors.New("sqlite driver not implemented, check connection factory")
	case "godror", "oracle":
		return NewOracleConnection(name, connString)
	default:
		return nil, errors.New(fmt.Sprintln("Driver not implemented for ", dbType, ". Check connection factory"))
	}
}
