package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/editor"
	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleAdd() {
	if len(os.Args) < 3 {
		printError("Usage: pam add <run-name> [query]")
	}

	if a.config.CurrentConnection == "" {
		printError("No active connection.  Use 'pam switch <connection>' or 'pam init' first")
	}

	_, ok := a.config.Connections[a.config.CurrentConnection]
	if !ok {
		a.config.Connections[a. config.CurrentConnection] = &config.ConnectionYAML{}
	}
	queries := a.config.Connections[a.config.CurrentConnection].Queries

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

		editedContent, err := editor.EditTempFileWithTemplate(header, "pam-new-run-")
		if err != nil {
			printError("Failed to open editor: %v", err)
		}

		querySQL = removeCommentLines(editedContent)
		querySQL = strings.TrimSpace(querySQL)

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

	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Added query '%s' with ID %d", queryName, queries[queryName].Id)))
}

func removeCommentLines(content string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}
