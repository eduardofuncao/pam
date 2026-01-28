package run

import "strings"

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
