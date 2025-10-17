package db

import (
	"fmt"
	"strings"
)

func NewDatabaseAdapter(config DatabaseConfig) (DatabaseAdapter, error) {
	dbType := strings.ToLower(config.DBType)
	
	switch dbType {
	case "postgres", "postgresql":
		return NewPostgresAdapter(config), nil
	// case "mysql":
	// 	return NewMySQLAdapter(config), nil
	// case "sqlite", "sqlite3":
	// 	return NewSQLiteAdapter(config), nil
	// case "oracle":
	// 	return NewOracleAdapter(config), nil
	// case "sqlserver", "mssql":
	// 	return NewSQLServerAdapter(config), nil
	// case "mongodb":
	// 	return NewMongoDBAdapter(config), nil
	// case "redis":
	// 	return NewRedisAdapter(config), nil
	// case "cassandra":
	// 	return NewCassandraAdapter(config), nil
	// case "mariadb":
	// 	return NewMariaDBAdapter(config), nil
	// case "cockroachdb", "cockroach":
	// 	return NewCockroachDBAdapter(config), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

func GetSupportedDatabases() []string {
	return []string{
		"postgres",
		// "mysql",
		// "sqlite",
		// "oracle",
		// "sqlserver",
		// "mongodb",
		// "redis",
		// "mariadb",
		// "cockroachdb",
	}
}
