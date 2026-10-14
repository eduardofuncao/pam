package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/eduardofuncao/squix/internal/db"
)

// QueryGroupKey derives the shared query-library key for a connection name.
func QueryGroupKey(name string) string {
	i := strings.LastIndex(name, ":")
	if i <= 0 { // no colon, or colon at position 0 (empty service)
		return name
	}
	return name[:i]
}

// QueriesFor returns the shared query-library map for connName
func (c *Config) QueriesFor(connName string) map[string]db.Query {
	key := GroupKey(connName)
	if c.QueryGroups[key] == nil {
		c.QueryGroups[key] = make(map[string]db.Query)
	}
	return c.QueryGroups[key]
}

// SetQueriesFor replaces the library contents for connName in place
func (c *Config) SetQueriesFor(connName string, m map[string]db.Query) {
	queries := c.QueriesFor(connName)
	for k := range queries {
		delete(queries, k)
	}
	for k, v := range m {
		queries[k] = v
	}
}

// LiveConnection builds a runtime DatabaseConnection for connName with its
// Queries populated from the shared library
func (c *Config) LiveConnection(connName string) db.DatabaseConnection {
	conn := FromConnectionYaml(c.Connections[connName])
	conn.SetQueries(c.QueriesFor(connName))
	return conn
}

// MigrateQueriesToFiles lifts legacy per-connection inline queries into
// per-group .sql files (one-way). Called once at load; returns true if anything
// was migrated so the caller can persist the cleaned config.yaml. Idempotent:
// no-op once a group's file exists / inline queries are cleared.
func (c *Config) MigrateQueriesToFiles() bool {
	if c.QueryGroups == nil {
		c.QueryGroups = make(map[string]map[string]db.Query)
	}

	names := make([]string, 0, len(c.Connections))
	for name := range c.Connections {
		names = append(names, name)
	}
	sort.Strings(names)

	migrated := false
	for _, name := range names {
		conn := c.Connections[name]
		if conn == nil || len(conn.Queries) == 0 {
			continue
		}
		key := GroupKey(name)
		if existing, ok := c.QueryGroups[key]; ok && len(existing) > 0 {
			// Group already populated (from its on-disk file or an earlier
			// sibling): alphabetically-first wins, rest dropped with a warning.
			dropped := make([]string, 0, len(conn.Queries))
			for q := range conn.Queries {
				dropped = append(dropped, q)
			}
			sort.Strings(dropped)
			fmt.Fprintf(os.Stderr,
				"squix: query group %q already populated; dropping legacy queries from %q: %s\n",
				key, name, strings.Join(dropped, ", "))
			conn.Queries = nil
			continue
		}

		queries := conn.Queries
		AssignIDs(queries)
		if err := WriteGroupFile(key, queries); err != nil {
			// Keep inline queries so the next run retries; don't mark migrated.
			fmt.Fprintf(os.Stderr,
				"squix: failed to migrate queries for %q to %s: %v (keeping inline)\n",
				name, GroupFile(key), err)
			continue
		}
		c.QueryGroups[key] = queries
		conn.Queries = nil
		migrated = true
	}
	return migrated
}

// GroupHasOtherMembers reports whether any connection other than exceptConnName
// derives the same group key.
func (c *Config) GroupHasOtherMembers(key, exceptConnName string) bool {
	for name := range c.Connections {
		if name == exceptConnName {
			continue
		}
		if GroupKey(name) == key {
			return true
		}
	}
	return false
}
