package db

import (
	"fmt"
	"regexp"
	"strings"
)

type TableMetadata struct {
	TableName    string
	PrimaryKeys  []string
	ColumnTypes  []string
	Columns      []string
	ForeignKeys  []ForeignKey
}

type ForeignKey struct {
	Column           string
	ReferencedTable  string
	ReferencedColumn string
}

func ExtractTableNameFromSQL(sqlQuery string) string {
	normalized := strings.Join(strings.Fields(strings.ToLower(sqlQuery)), " ")

	// Try to match: SELECT ...  FROM tablename
	patterns := []string{
		`from\s+([a-z_][a-z0-9_\. ]*)\s+(?:as\s+)? [a-z_]`,                   // FROM table alias (FIXED: removed space)
		`from\s+([a-z_][a-z0-9_\.]*)\s+(?:where|join|group|order|limit|;|$)`, // FROM table WHERE/JOIN/etc
		`from\s+([a-z_][a-z0-9_\.]*)`,                                        // FROM table (fallback)
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(normalized); len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// InferTableMetadata attempts to infer table metadata from a query
func InferTableMetadata(conn DatabaseConnection, query Query) (*TableMetadata, error) {
	if query.TableName != "" {
		metadata := &TableMetadata{
			TableName: query.TableName,
		}

		if len(query.PrimaryKeys) == 0 && conn != nil {
			if dbMeta, err := conn.GetTableMetadata(query.TableName); err == nil {
				metadata.PrimaryKeys = dbMeta.PrimaryKeys
			}
		} else {
			metadata.PrimaryKeys = query.PrimaryKeys
		}

		return metadata, nil
	}

	tableName := ExtractTableNameFromSQL(query.SQL)
	if tableName == "" {
		// Check if this is a JOIN query
		if HasJoinClause(query.SQL) {
			// Try to extract the primary table from the JOIN
			primaryTable := ExtractPrimaryTableFromJoin(query.SQL)
			if primaryTable != "" && conn != nil {
				// Get metadata for the primary table
				return conn.GetTableMetadata(primaryTable)
			}
			// Return empty metadata for complex JOINs
			return &TableMetadata{TableName: ""}, nil
		}
		return nil, fmt.Errorf("could not extract table name from query")
	}

	if conn != nil {
		return conn.GetTableMetadata(tableName)
	}

	return &TableMetadata{
		TableName: tableName,
	}, nil
}

// ExtractPrimaryTableFromJoin attempts to identify the primary/updateable table in a JOIN query
// For simple JOINs, returns the first table in the FROM clause
func ExtractPrimaryTableFromJoin(sqlQuery string) string {
	normalized := strings.Join(strings.Fields(strings.ToLower(sqlQuery)), " ")

	// Extract FROM clause
	fromPattern := regexp.MustCompile(`from\s+([^where^group^order^limit^;]+)`)
	matches := fromPattern.FindStringSubmatch(normalized)
	if len(matches) < 2 {
		return ""
	}

	fromClause := strings.TrimSpace(matches[1])

	// Split by JOIN to get the first table
	// Handle various JOIN types
	joinPattern := regexp.MustCompile(`\s+(?:inner|left|right|full|outer|cross)\s+join\s+`)
	tables := joinPattern.Split(fromClause, -1)

	// Get the first table (before any JOIN)
	firstTable := strings.TrimSpace(tables[0])

	// Remove any alias (e.g., "users u" or "users AS u")
	tableParts := strings.Fields(firstTable)
	if len(tableParts) > 0 {
		// The first part should be the table name
		tableName := tableParts[0]
		// Clean up schema prefix if present (e.g., "public.users" -> "users")
		if dotIdx := strings.LastIndex(tableName, "."); dotIdx != -1 {
			tableName = tableName[dotIdx+1:]
		}
		return tableName
	}

	return ""
}

func HasJoinClause(sqlQuery string) bool {
	normalized := strings.ToUpper(strings.Join(strings.Fields(sqlQuery), " "))

	joinKeywords := []string{
		" JOIN ",
		" INNER JOIN ",
		" LEFT JOIN ",
		" RIGHT JOIN ",
		" FULL JOIN ",
		" OUTER JOIN ",
		" CROSS JOIN ",
		" LEFT OUTER JOIN ",
		" RIGHT OUTER JOIN ",
		" FULL OUTER JOIN ",
	}

	for _, keyword := range joinKeywords {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}

	// Also check for comma-separated implicit joins (old style)
	fromPattern := regexp.MustCompile(`FROM\s+([^WHERE^GROUP^ORDER^LIMIT^;]+)`)
	if matches := fromPattern.FindStringSubmatch(normalized); len(matches) > 1 {
		tables := strings.Split(matches[1], ",")
		if len(tables) > 1 {
			return true
		}
	}

	return false
}
