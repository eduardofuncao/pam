package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
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
		printError("Unknown edit type: %s.   Use 'config' or 'queries'", editType)
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
		log.Fatal("No active connection. Use 'pam switch <connection>' first")
	}

	conn, ok := a.config.Connections[a.config.CurrentConnection]
	if !ok {
		log.Fatalf("Connection %s not found", a.config.CurrentConnection)
	}

	tmpFile, err := os.CreateTemp("", "pam-queries-*.sql")
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	var content strings.Builder
	content.WriteString(fmt.Sprintf("-- Editing queries for connection: %s (%s)\n", 
		a.config.CurrentConnection, conn.DBType))
	content.WriteString("-- Format: -- queryname\n")
	content.WriteString("--         SQL query here\n")
	content.WriteString("-- Save and close to update\n\n")

	for _, query := range conn.Queries {
		content.WriteString(fmt.Sprintf("-- %s\n", query.Name))
		content.WriteString(strings.TrimSpace(query.SQL))
		content.WriteString("\n\n")
	}

	if _, err := tmpFile.Write([]byte(content.String())); err != nil {
		log.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cmd := exec.Command(editorCmd, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to open editor: %v", err)
	}

	editedData, err := os.ReadFile(tmpPath)
	if err != nil {
		log.Fatalf("Failed to read edited file: %v", err)
	}

	editedQueries, err := parseSQLQueriesFile(string(editedData))
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
	lines := strings.Split(content, "\n")
	
	var currentQueryName string
	var currentSQL strings.Builder
	queryID := 0

	for i := range lines {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if currentQueryName != "" {
				currentSQL.WriteString("\n")
			}
			continue
		}

		if strings.HasPrefix(trimmed, "--") {
			comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "--"))
			
			if strings.HasPrefix(comment, "Editing queries") || 
			   strings.HasPrefix(comment, "Format:") ||
			   strings.HasPrefix(comment, "SQL query") ||
			   strings.HasPrefix(comment, "Save and close") {
				continue
			}

			if currentQueryName != "" {
				sql := strings.TrimSpace(currentSQL.String())
				if sql != "" {
					queries[currentQueryName] = db.Query{
						Name: currentQueryName,
						SQL:  sql,
						Id:   queryID,
					}
					queryID++
				}
				currentSQL.Reset()
			}

			currentQueryName = comment
		} else if currentQueryName != "" {
			// Add line to current query SQL
			if currentSQL.Len() > 0 {
				currentSQL.WriteString("\n")
			}
			currentSQL.WriteString(line)
		}
	}

	if currentQueryName != "" {
		sql := strings.TrimSpace(currentSQL.String())
		if sql != "" {
			queries[currentQueryName] = db.Query{
				Name: currentQueryName,
				SQL:  sql,
				Id:   queryID,
			}
		}
	}

	return queries, nil
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
