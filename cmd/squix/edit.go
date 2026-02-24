package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

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

	conn, ok := a.config.Connections[a.config.CurrentConnection]
	if !ok {
		log.Fatalf("Connection %s not found", a.config.CurrentConnection)
	}

	// Find the query
	query, exists := db.FindQueryWithSelector(conn.Queries, selector)
	if !exists {
		log.Fatalf("Query '%s' not found in connection '%s'", selector, a.config.CurrentConnection)
	}

	// Create temp file with the query SQL
	var content strings.Builder
	content.WriteString(fmt.Sprintf("-- %s\n", query.Name))
	content.WriteString(query.SQL)

	tmpFile, err := editor.CreateTempFile("squix-edit-query-", content.String())
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	tmpFile.Close()

	// Open editor
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
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

	newName, newSQL, err := parseSingleQueryFile(editedData)
	if err != nil {
		log.Fatalf("Failed to parse edited query: %v", err)
	}

	if newName != query.Name {
		if !a.confirmQueryRename(query.Name, newName) {
			fmt.Println(styles.Faint.Render("Aborted"))
			return
		}

		delete(conn.Queries, query.Name)
	}

	// Update query
	query.Name = newName
	query.SQL = newSQL
	conn.Queries[query.Name] = query
	a.config.Connections[a.config.CurrentConnection] = conn

	if err := a.config.Save(); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	if newName != query.Name {
		fmt.Printf("✓ Renamed and updated query '%s' → '%s'\n", query.Name, newName)
	} else {
		fmt.Printf("✓ Updated query '%s'\n", query.Name)
	}
}

func (a *App) editQueries() {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
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

	var content strings.Builder
	content.WriteString(fmt.Sprintf("-- Editing queries for connection: %s (%s)\n",
		a.config.CurrentConnection, conn.DBType))
	content.WriteString("-- Format: -- runname\n")
	content.WriteString("--         SQL run here\n")
	content.WriteString("-- Save and close to update\n\n")

	for _, query := range conn.Queries {
		content.WriteString(fmt.Sprintf("-- %s\n", query.Name))
		content.WriteString(strings.TrimSpace(query.SQL))
		content.WriteString("\n\n")
	}

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

	editedQueries, err := parseSQLQueriesFile(editedData)
	if err != nil {
		log.Fatalf("Failed to parse edited queries: %v", err)
	}

	conn.Queries = editedQueries
	a.config.Connections[a.config.CurrentConnection] = conn

	if err := a.config.Save(); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	fmt.Printf("✓ Updated queries for connection: %s\n", a.config.CurrentConnection)
}

// parseSingleQueryFile parses a file containing a single query
// Expected format:
//   -- queryname
//   SQL query here
func parseSingleQueryFile(content string) (string, string, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) == 0 {
		return "", "", fmt.Errorf("empty file")
	}

	// First non-empty line should be the query name comment
	var queryName string
	var sqlLines []string
	foundName := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines at the start
		if !foundName && trimmed == "" {
			continue
		}

		// Look for query name comment
		if !foundName {
			if comment, ok := strings.CutPrefix(trimmed, "--"); ok {
				queryName = strings.TrimSpace(comment)
				foundName = true
				continue
			} else {
				// First non-empty line is not a comment
				// Treat entire content as SQL with no name change
				queryName = "" // Will trigger error if not provided
				break
			}
		}

		// Rest is SQL
		if foundName {
			if !strings.HasPrefix(trimmed, "--") {
				sqlLines = append(sqlLines, line)
			}
		}
	}

	if queryName == "" {
		return "", "", fmt.Errorf("query name not found (expected '-- queryname' on first line)")
	}

	if len(sqlLines) == 0 {
		return queryName, "", fmt.Errorf("no SQL content found")
	}

	sql := strings.Join(sqlLines, "\n")
	return queryName, sql, nil
}

func (a *App) confirmQueryRename(oldName, newName string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(styles.Error.Render(fmt.Sprintf("Rename query '%s' → '%s'? [y/N]: ", oldName, newName)))

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// parseSQLQueriesFile parses a SQL file with the format:
// -- queryname
// SQL query here
func parseSQLQueriesFile(content string) (map[string]db.Query, error) {
	queries := make(map[string]db.Query)
	var name string
	var sql strings.Builder
	id := 1

	save := func() {
		if name != "" && sql.Len() > 0 {
			queries[name] = db.Query{Name: name, SQL: strings.TrimSpace(sql.String()), Id: id}
			id++
			sql.Reset()
		}
	}

	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)

		// Check for query name comment
		if comment, ok := strings.CutPrefix(trimmed, "--"); ok {
			comment = strings.TrimSpace(comment)

			// Skip help comments
			if strings.HasPrefix(comment, "Editing") || strings.HasPrefix(comment, "Format") ||
				strings.HasPrefix(comment, "SQL") || strings.HasPrefix(comment, "Save") {
				continue
			}

			save()
			name = comment
			continue
		}

		// Add SQL line
		if name != "" {
			if sql.Len() > 0 {
				sql.WriteString("\n")
			}
			sql.WriteString(line)
		}
	}

	save()
	return queries, nil
}
