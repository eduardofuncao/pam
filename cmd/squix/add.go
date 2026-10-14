package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/eduardofuncao/squix/internal/db"
	"github.com/eduardofuncao/squix/internal/editor"
	"github.com/eduardofuncao/squix/internal/styles"
)

func (a *App) handleAdd() {
	if len(os.Args) < 3 {
		printError("Usage: squix add <run-name> [query]")
	}

	if a.config.CurrentConnection == "" {
		printError("No active connection.  Use 'squix switch <connection>' or 'squix init' first")
	}

	queries := a.config.QueriesFor(a.config.CurrentConnection)

	queryName := os.Args[2]
	var querySQL string

	if len(os.Args) >= 4 {
		querySQL = os.Args[3]
	} else {
		header := fmt.Sprintf("-- Creating new run:  %s\n", queryName)
		header += fmt.Sprintf("-- Connection: %s (%s)\n",
			a.config.CurrentConnection,
			a.config.Connections[a.config.CurrentConnection].DBType)
		header += "-- Write your SQL run below and save\n\n"

		editedContent, cancelled, err := editor.EditTempFileWithTemplate(header, "squix-new-run-")
		if err != nil {
			printError("Failed to open editor: %v", err)
		}
		if cancelled {
			printError("Cancelled")
		}

		querySQL = strings.TrimSpace(editedContent)

		if querySQL == "" {
			printError("No SQL run provided.   Run not saved")
		}
	}

	queries[queryName] = db.Query{
		Name: queryName,
		SQL:  querySQL,
		Id:   db.GetNextQueryId(queries),
	}

	err := a.config.Save()
	if err != nil {
		printError("Could not save configuration file: %v", err)
	}

	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Added query '%s'", queryName)))
}
