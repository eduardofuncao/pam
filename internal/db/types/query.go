package types

import (
	"strconv"
)

type Query struct {
	Name string
	Id   int
	SQL  string
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
