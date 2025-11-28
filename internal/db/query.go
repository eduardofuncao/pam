package db

import (
	"strconv"
)

type Query struct {
	Name       string            `yaml:"name"`
	Id         int               `yaml:"id"`
	SQL        string            `yaml:"sql"`
	TableName  string            `yaml:"table_name,omitempty"`
	PrimaryKey string            `yaml:"primary_key,omitempty"`
	Metadata   map[string]string `yaml:"metadata,omitempty"`
}

func FindQueryWithSelector(queries map[string]Query, selector string) (Query, bool) {
	if id, err := strconv.Atoi(selector); err == nil {
		for _, q := range queries {
			if q.Id == id {
				return q, true
			}
		}
		return Query{}, false
	}
	q, ok := queries[selector]
	return q, ok
}
