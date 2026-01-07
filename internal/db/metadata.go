package db

import (
	"fmt"
	"regexp"
	"strings"
)

type TableMetadata struct {
	TableName   string
	PrimaryKey  string
	ColumnTypes []string
	Columns     []string
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
	if HasJoinClause(query.SQL) {
		return nil, fmt.Errorf("update/delete operations are not supported for JOIN queries")
	}
	
	if query.TableName != "" {
		metadata := &TableMetadata{
			TableName:  query. TableName,
			PrimaryKey: query.PrimaryKey,
		}
		
		if metadata.PrimaryKey == "" && conn != nil {
			if dbMeta, err := conn.GetTableMetadata(query.TableName); err == nil {
				metadata. PrimaryKey = dbMeta. PrimaryKey
			}
		}
		
		return metadata, nil
	}
	
	tableName := ExtractTableNameFromSQL(query.SQL)
	if tableName == "" {
		return nil, fmt.Errorf("could not extract table name from query")
	}
	
	if conn != nil {
		return conn.GetTableMetadata(tableName)
	}
	
	return &TableMetadata{
		TableName: tableName,
	}, nil
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
	// Simple heuristic: more than one table in FROM clause
	fromPattern := regexp.MustCompile(`FROM\s+([^WHERE^GROUP^ORDER^LIMIT^;]+)`)
	if matches := fromPattern.FindStringSubmatch(normalized); len(matches) > 1 {
		tables := strings.Split(matches[1], ",")
		if len(tables) > 1 {
			return true
		}
	}

	return false
}
