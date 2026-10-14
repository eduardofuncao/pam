package config

import (
	"fmt"

	"github.com/eduardofuncao/squix/internal/db"
)

func GetNextQueryId(queries map[string]db.Query) int {
	maxID := 0
	for _, q := range queries {
		if q.Id > maxID {
			maxID = q.Id
		}
	}
	return maxID + 1
}

// SaveQueryToConnection saves a query to a connection, generating an ID if needed
func (c *Config) SaveQueryToConnection(connName string, query db.Query) (db.Query, error) {
	queries := c.QueriesFor(connName)

	if query.Id == -1 {
		if _, exists := queries[query.Name]; exists {
			return db.Query{}, fmt.Errorf("query '%s' already exists", query.Name)
		}
		query.Id = GetNextQueryId(queries)
	}

	queries[query.Name] = query

	if err := c.Save(); err != nil {
		return db.Query{}, err
	}

	return query, nil
}
