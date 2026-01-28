package config

import (
	"fmt"

	"github.com/eduardofuncao/pam/internal/db"
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
// If the query already exists (by name), it returns an error
// If query.Id == -1, a new ID will be generated
func (c *Config) SaveQueryToConnection(connName string, query db.Query) (db.Query, error) {
	connData := c.Connections[connName]

	// Check if query with this name already exists (when creating new)
	if query.Id == -1 {
		if _, exists := connData.Queries[query.Name]; exists {
			return db.Query{}, fmt.Errorf("query '%s' already exists", query.Name)
		}
		// Generate new ID
		query.Id = GetNextQueryId(connData.Queries)
	}

	// Save the query
	connData.Queries[query.Name] = query

	// Save config
	if err := c.Save(); err != nil {
		return db.Query{}, err
	}

	return query, nil
}

func (c *Config) UpdateLastQuery(connName string, query db.Query) error {
	connData := c.Connections[connName]
	connData.LastQuery = query
	return c.Save()
}

func (c *Config) SaveQueryAndLast(connName string, query db.Query, saveAsLast bool) error {
	connData := c.Connections[connName]

	// Save the query (if it has a name and isn't inline)
	if query.Name != "<inline>" && query.Name != "" && query.SQL != "" {
		connData.Queries[query.Name] = query
	}

	// Update last query if requested
	if saveAsLast {
		connData.LastQuery = query
	}

	return c.Save()
}
