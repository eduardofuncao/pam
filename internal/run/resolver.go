package run

import (
	"fmt"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
)

// ResolveQuery determines which query to run based on flags and config
// Priority:
//  1. Last query (if --last/-l flag)
//  2. Inline SQL (if selector looks like SQL)
//  3. Saved query by name/ID
//  4. Create new in editor (default)
func ResolveQuery(flags Flags, cfg *config.Config, currentConn string, conn db.DatabaseConnection) (ResolvedQuery, error) {
	// Priority 1: Last query with --last/-l flag
	if flags.LastQuery {
		if currentConn == "" {
			return ResolvedQuery{}, fmt.Errorf("no active connection")
		}
		lastQuery := cfg.Connections[currentConn].LastQuery
		if lastQuery.Name == "" {
			return ResolvedQuery{}, fmt.Errorf("no last query found. Run a query first, then use pam run --last")
		}
		return ResolvedQuery{
			Query:    lastQuery,
			Saveable: true,
		}, nil
	}

	// Priority 2: Inline SQL (pam run "select * from employees")
	if flags.Selector != "" && IsLikelySQL(flags.Selector) {
		return ResolvedQuery{
			Query:    db.Query{Name: "<inline>", SQL: flags.Selector, Id: -1},
			Saveable: false,
		}, nil
	}

	// Priority 3: Saved query by name/ID
	if flags.Selector != "" {
		q, found := db.FindQueryWithSelector(conn.GetQueries(), flags.Selector)
		if !found {
			return ResolvedQuery{}, fmt.Errorf("could not find query with name/id: %v", flags.Selector)
		}
		return ResolvedQuery{
			Query:    q,
			Saveable: true,
		}, nil
	}

	// Priority 4: Default - will create new query in editor (pam run with no args)
	// Return a placeholder that indicates caller should prompt for new query
	return ResolvedQuery{
		Query:    db.Query{Name: "<new>", SQL: "", Id: -1},
		Saveable: false,
	}, nil
}

// ShouldCreateNewQuery returns true if the resolved query indicates a new query should be created
func ShouldCreateNewQuery(resolved ResolvedQuery) bool {
	return resolved.Query.Name == "<new>" && resolved.Query.SQL == ""
}
