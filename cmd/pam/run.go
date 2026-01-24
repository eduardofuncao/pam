package main

import (
	"fmt"
	"os"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/editor"
	"github.com/eduardofuncao/pam/internal/run"
)

func (a *App) handleRun() {
	if a.config.CurrentConnection == "" {
		printError("No active connection.   Use 'pam switch <connection>' or 'pam init' first")
	}

	flags := parseRunFlags()
	conn := config.FromConnectionYaml(a.config.Connections[a.config.CurrentConnection])

	resolved, err := run.ResolveQuery(flags, a.config, a.config.CurrentConnection, conn)
	if err != nil {
		printError("%v", err)
	}

	// Check if we need to create a new query via editor
	if run.ShouldCreateNewQuery(resolved) {
		newQuery := a.createNewQueryOrEdit()
		resolved.Query = newQuery
	}

	if flags.EditMode && !flags.LastQuery {
		resolved.Query = a.editQueryOrExit(resolved.Query)
	}

	a.saveIfNeeded(resolved)
	a.executeQuery(resolved.Query, conn)
}

func parseRunFlags() run.Flags {
	flags := run.Flags{}
	for _, arg := range os.Args[2:] {
		switch arg {
		case "--edit", "-e":
			flags.EditMode = true
		case "--last", "-l":
			flags.LastQuery = true
		default:
			flags.Selector = arg
		}
	}
	return flags
}

func (a *App) createNewQueryOrEdit() db.Query {
	instructions := `-- Enter your SQL run below
-- Save and exit to execute, or exit without saving to cancel
--
`
	editedSQL, err := editor.EditTempFileWithTemplate(instructions, "pam-run-")
	if err != nil {
		printError("Error opening editor: %v", err)
	}
	if editedSQL == "" {
		printError("Empty SQL, cancelled")
	}
	return db.Query{Name: "<runtime>", SQL: editedSQL, Id: -1}
}

func (a *App) editQueryOrExit(query db.Query) db.Query {
	editedSQL, err := editor.EditTempFile(query.SQL, "pam-run-")
	if err != nil {
		printError("Error opening editor: %v", err)
	}
	query.SQL = editedSQL
	return query
}

func (a *App) saveIfNeeded(resolved run.ResolvedQuery) {
	if !resolved.Saveable {
		return
	}

	// Save the query and update last query
	if err := a.config.SaveQueryAndLast(a.config.CurrentConnection, resolved.Query, true); err != nil {
		printError("Failed to save query: %v", err)
	}
}

func (a *App) executeQuery(query db.Query, conn db.DatabaseConnection) {
	originalQuery := query
	run.Execute(run.ExecutionParams{
		Query:        query,
		Connection:   conn,
		Config:       a.config,
		SaveCallback: a.saveQueryFromTable,
		OnRerun: func(editedSQL string) {
			// Re-run callback
			editedQuery := db.Query{
				Name: originalQuery.Name,
				SQL:  editedSQL,
				Id:   originalQuery.Id,
			}
			a.executeQuery(editedQuery, conn)
		},
	})
}

func (a *App) saveQueryFromTable(query db.Query) (db.Query, error) {
	connName := a.config.CurrentConnection
	if connName == "" {
		return db.Query{}, fmt.Errorf("no active connection")
	}

	// Save query with auto-ID generation
	savedQuery, err := a.config.SaveQueryToConnection(connName, query)
	if err != nil {
		return db.Query{}, err
	}

	// Update last query
	if err := a.config.UpdateLastQuery(connName, savedQuery); err != nil {
		return db.Query{}, err
	}

	return savedQuery, nil
}
