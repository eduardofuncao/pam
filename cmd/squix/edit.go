package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/eduardofuncao/squix/internal/config"
	"github.com/eduardofuncao/squix/internal/db"
	"github.com/eduardofuncao/squix/internal/editor"
	"github.com/eduardofuncao/squix/internal/styles"
)

func (a *App) handleEdit() {
	if len(os.Args) >= 3 {
		querySelector := os.Args[2]
		if querySelector == "config" {
			printError("Config editing moved to 'squix config' command")
			return
		}
		a.editSingleQuery(querySelector)
	} else {
		a.editQueries()
	}
}

func (a *App) editSingleQuery(selector string) {
	if a.config.CurrentConnection == "" {
		log.Fatal("No active connection. Use 'squix switch <connection>' or 'squix init' first")
	}

	if _, ok := a.config.Connections[a.config.CurrentConnection]; !ok {
		log.Fatalf("Connection %s not found", a.config.CurrentConnection)
	}
	queries := a.config.QueriesFor(a.config.CurrentConnection)

	// Find the query
	query, exists := db.FindQueryWithSelector(queries, selector)
	if !exists {
		log.Fatalf("Query '%s' not found in connection '%s'", selector, a.config.CurrentConnection)
	}

	// Render the single query block (round-trips table/pk metadata too).
	tmpFile, err := editor.CreateTempFile("squix-edit-query-", config.RenderQuery(query))
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	tmpFile.Close()

	editorCmd, err := editor.CheckEditor()
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(editorCmd, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to open editor: %v", err)
	}

	editedData, err := editor.ReadTempFile(tmpPath)
	if err != nil {
		log.Fatalf("Failed to read edited file: %v", err)
	}

	parsed, err := config.ParseGroupFile([]byte(editedData))
	if err != nil {
		log.Fatalf("Failed to parse edited query: %v", err)
	}
	if len(parsed) != 1 {
		log.Fatalf("Expected exactly one query block, got %d", len(parsed))
	}
	var edited db.Query
	for _, q := range parsed {
		edited = q
	}

	oldName := query.Name
	if edited.Name != oldName {
		if !a.confirmQueryRename(oldName, edited.Name) {
			fmt.Println(styles.Faint.Render("Aborted"))
			return
		}
		delete(queries, oldName)
	}

	// Update query (preserve id; reassign the rest from the edited block).
	query.Name = edited.Name
	query.SQL = edited.SQL
	query.TableName = edited.TableName
	query.PrimaryKeys = edited.PrimaryKeys
	queries[query.Name] = query

	if err := a.config.Save(); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	if edited.Name != oldName {
		fmt.Printf("✓ Renamed and updated query '%s' → '%s'\n", oldName, edited.Name)
	} else {
		fmt.Printf("✓ Updated query '%s'\n", query.Name)
	}
}

func (a *App) editQueries() {
	editorCmd, err := editor.CheckEditor()
	if err != nil {
		log.Fatal(err)
	}

	a.editQueriesWithEditor(editorCmd)
	fmt.Println(styles.Success.Render("✓ Queries edited"))
}

func (a *App) editQueriesWithEditor(editorCmd string) {
	if a.config.CurrentConnection == "" {
		log.Fatal("No active connection. Use 'squix switch <connection>' or 'squix init' first")
	}

	conn, ok := a.config.Connections[a.config.CurrentConnection]
	if !ok {
		log.Fatalf("Connection %s not found", a.config.CurrentConnection)
	}
	queries := a.config.QueriesFor(a.config.CurrentConnection)

	var content strings.Builder
	fmt.Fprintf(&content, "-- Editing queries for connection: %s (%s)\n",
		a.config.CurrentConnection, conn.DBType)
	content.WriteString("-- Format: -- @query <name>  (+ optional -- @table / -- @pk)\n")
	content.WriteString("-- Save and close to update\n\n")
	content.WriteString(config.RenderGroup(queries))

	tmpFile, err := editor.CreateTempFile("squix-queries-", content.String())
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	tmpFile.Close()

	cmd := exec.Command(editorCmd, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to open editor: %v", err)
	}

	editedData, err := editor.ReadTempFile(tmpPath)
	if err != nil {
		log.Fatalf("Failed to read edited file: %v", err)
	}

	editedQueries, err := config.ParseGroupFile([]byte(editedData))
	if err != nil {
		log.Fatalf("Failed to parse edited queries: %v", err)
	}
	config.AssignIDs(editedQueries)

	// Replace the shared library in place so sibling connections see the change.
	a.config.SetQueriesFor(a.config.CurrentConnection, editedQueries)

	if err := a.config.Save(); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	fmt.Printf("✓ Updated queries for connection: %s\n", a.config.CurrentConnection)
}

func (a *App) confirmQueryRename(oldName, newName string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(styles.Error.Render(fmt.Sprintf("Rename query '%s' → '%s'? [y/N]: ", oldName, newName)))

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
