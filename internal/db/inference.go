package db

import (
	"strings"
)

// InferDBType attempts to infer the database type from a connection string.
// Returns the detected type or empty string if unable to infer.
func InferDBType(connString string) string {
	conn := strings.TrimSpace(connString)

	// URL scheme detection
	if strings.HasPrefix(conn, "postgres://") || strings.HasPrefix(conn, "postgresql://") {
		return "postgres"
	}
	if strings.HasPrefix(conn, "mysql://") || strings.HasPrefix(conn, "mariadb://") {
		return "mysql"
	}
	if strings.HasPrefix(conn, "sqlserver://") || strings.HasPrefix(conn, "mssql://") {
		return "sqlserver"
	}
	if strings.HasPrefix(conn, "clickhouse://") {
		return "clickhouse"
	}
	if strings.HasPrefix(conn, "file://") {
		return "sqlite"
	}

	// SQLite file pattern detection
	if strings.HasSuffix(conn, ".db") ||
	   strings.HasSuffix(conn, ".sqlite") ||
	   strings.HasSuffix(conn, ".sqlite3") {
		return "sqlite"
	}

	return ""
}

// GetSupportedDBTypes returns a list of all supported database types.
func GetSupportedDBTypes() []string {
	return []string{
		"postgres",
		"mysql",
		"sqlite",
		"sqlserver",
		"clickhouse",
		"oracle",
		"firebird",
	}
}
