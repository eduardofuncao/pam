package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/editor"
	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleEdit() {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
	}

	editType := "config"
	if len(os.Args) >= 3 {
		editType = os.Args[2]
	}

	switch editType {
	case "config":
		a.editConfig(editorCmd)
		fmt.Println(styles.Success.Render("✓ Config file edited"))
	case "queries":
		a.editQueries(editorCmd)
		fmt.Println(styles.Success. Render("✓ Queries edited"))
	default:
		printError("Unknown edit type: %s. Use 'config' or 'queries'", editType)
	}
}

func (a *App) editConfig(editorCmd string) {
	cmd := exec.Command(editorCmd, config.CfgFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to open editor: %v", err)
	}

	cfg, err := config.LoadConfig(config.CfgFile)
	if err != nil {
		log.Printf("Warning: Could not reload config: %v", err)
	} else {
		a.config = cfg
	}
}

func (a *App) editQueries(editorCmd string) {
	if a.config.CurrentConnection == "" {
		log.Fatal("No active connection. Use 'pam switch <connection>' or 'pam init' first")
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

	tmpFile, err := editor.CreateTempFile("pam-queries-", content.String())
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

// parseSQLQueriesFile parses a SQL file with the format:
// -- queryname
// SQL query here
func parseSQLQueriesFile(content string) (map[string]db.Query, error) {
	queries := make(map[string]db.Query)
	var name string
	var sql strings. Builder
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

