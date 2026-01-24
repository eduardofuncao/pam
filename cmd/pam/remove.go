package main

import (
	"fmt"
	"os"

	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleRemove() {
	if len(os.Args) < 3 {
		printError("Usage:  pam remove <run-name>")
	}

	conn := a.config.Connections[a.config.CurrentConnection]
	queries := conn.Queries

	query, exists := db.FindQueryWithSelector(queries, os.Args[2])
	if !exists {
		printError("Query '%s' could not be found", os.Args[2])
	}

	delete(conn.Queries, query.Name)

	err := a.config.Save()
	if err != nil {
		printError("Could not save configuration file: %v", err)
	}

	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Removed run '%s'", query.Name)))
}
