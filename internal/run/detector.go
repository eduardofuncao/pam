package run

import "strings"

// IsSelectQuery detects if the SQL is a SELECT-type query (returns data)
func IsSelectQuery(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	keywords := []string{"SELECT", "WITH", "SHOW", "DESCRIBE", "DESC", "EXPLAIN", "PRAGMA"}

	for _, kw := range keywords {
		if upper == kw || strings.HasPrefix(upper, kw+" ") {
			return true
		}
	}
	return false
}

// IsLikelySQL uses heuristics to detect if a string is likely a SQL statement
func IsLikelySQL(s string) bool {
	upper := strings.ToUpper(strings.TrimSpace(s))
	keywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER", "TRUNCATE",
		"WITH", "SHOW", "DESCRIBE", "DESC", "EXPLAIN", "GRANT", "REVOKE",
		"BEGIN", "COMMIT", "ROLLBACK", "PRAGMA",
	}

	for _, kw := range keywords {
		if upper == kw || strings.HasPrefix(upper, kw+" ") {
			return true
		}
	}
	return false
}
