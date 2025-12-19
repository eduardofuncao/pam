package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleAdd() {
	if len(os.Args) < 3 {
		printError("Usage: pam add <query-name> [query]")
	}

	if a.config.CurrentConnection == "" {
		printError("No active connection.  Use 'pam switch <connection>' first")
	}

	_, ok := a.config.Connections[a.config.CurrentConnection]
	if !ok {
		a.config.Connections[a. config.CurrentConnection] = &config.ConnectionYAML{}
	}
	queries := a.config.Connections[a.config.CurrentConnection]. Queries

	queryName := os.Args[2]
	var querySQL string

	if len(os.Args) >= 4 {
		querySQL = os.Args[3]
	} else {
		editorCmd := os.Getenv("EDITOR")
		if editorCmd == "" {
			editorCmd = "vim"
		}

		tmpFile, err := os.CreateTemp("", "pam-new-query-*.  sql")
		if err != nil {
			printError("Failed to create temp file: %v", err)
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		header := fmt.Sprintf("-- Creating new query:  %s\n", queryName)
		header += fmt.Sprintf("-- Connection: %s (%s)\n",
			a.config.CurrentConnection,
			a.config.Connections[a.config.CurrentConnection]. DBType)
		header += "-- Write your SQL query below and save\n\n"

		if _, err := tmpFile.Write([]byte(header)); err != nil {
			printError("Failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		cmd := exec.Command(editorCmd, tmpPath)
		cmd.Stdin = os.Stdin
		cmd. Stdout = os.Stdout
		cmd.Stderr = os. Stderr
		if err := cmd.Run(); err != nil {
			printError("Failed to open editor: %v", err)
		}

		editedData, err := os.ReadFile(tmpPath)
		if err != nil {
			printError("Failed to read edited file: %v", err)
		}

		querySQL = removeCommentLines(string(editedData))
		querySQL = strings.TrimSpace(querySQL)

		if querySQL == "" {
			printError("No SQL query provided.   Query not saved")
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
