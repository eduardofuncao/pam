package run

import "github.com/eduardofuncao/pam/internal/db"

type Flags struct {
	EditMode  bool
	LastQuery bool
	Selector  string
}

type ResolvedQuery struct {
	Query    db.Query
	Saveable bool // will be saved to config file
}
