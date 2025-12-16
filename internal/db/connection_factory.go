package db

import(
	"errors"
	"fmt"

	"github.com/eduardofuncao/pam/internal/db/connections"
)

func CreateConnection(name, dbType, connString string) (connections.DatabaseConnection, error) {
	switch dbType{
	case "postgres":
		return connections.NewPostgresConnection(name, connString)
	case "sqlserver", "mssql":
		return connections.NewSqlServerConnection(name, connString)
	case "myslq", "mariadb":
		return nil, errors.New("mysql driver not implemented, check connection factory")
	case "sqlite", "sqlite3":
		return nil, errors.New("sqlite driver not implemented, check connection factory")
	case "godror", "oracle":
		return connections.NewOracleConnection(name, connString)
	default:
		return nil, errors.New(fmt.Sprintln("Driver not implemented for ", dbType, ". Check connection factory"))
	}
}
