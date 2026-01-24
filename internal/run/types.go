package run

import "github.com/eduardofuncao/pam/internal/db"

// Flags represents command-line flags for the run command
type Flags struct {
	EditMode  bool
	LastQuery bool
	Selector  string
}

// ResolvedQuery represents a query that has been resolved from user input
type ResolvedQuery struct {
	Query    db.Query
	Saveable bool // will be saved to config file
}
