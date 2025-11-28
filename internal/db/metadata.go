package db

import (
	"fmt"
	"regexp"
	"strings"
)

type TableMetadata struct {
	TableName  string
	PrimaryKey string
	Columns    []string
}

func ExtractTableNameFromSQL(sqlQuery string) string {
	normalized := strings.Join(strings.Fields(strings.ToLower(sqlQuery)), " ")
	
	// Try to match: SELECT ...  FROM tablename
	patterns := []string{
		`from\s+([a-z_][a-z0-9_\. ]*)\s+(?:as\s+)? [a-z_]`,      // FROM table alias (FIXED: removed space)
		`from\s+([a-z_][a-z0-9_\.]*)\s+(?:where|join|group|order|limit|;|$)`, // FROM table WHERE/JOIN/etc
		`from\s+([a-z_][a-z0-9_\.]*)`,                         // FROM table (fallback)
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
			TableName:  query.TableName,
			PrimaryKey: query.PrimaryKey,
		}
		
		// If PrimaryKey is not set, try to fetch it from the database
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
