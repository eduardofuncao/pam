package commands

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
)

func Add(cfg *config.Config) {
	if len(os.Args) < 3 {
		log.Fatal("Usage: pam add <query-name> [query]")
	}

	if cfg.CurrentConnection == "" {
		log.Fatal("No active connection. Use 'pam switch <connection>' first")
	}

	_, ok := cfg.Connections[cfg.CurrentConnection]
	if !ok {
		cfg.Connections[cfg.CurrentConnection] = config.ConnectionYAML{}
	}
	queries := cfg.Connections[cfg.CurrentConnection].Queries

	queryName := os.Args[2]
	var querySQL string

	if len(os.Args) >= 4 {
		querySQL = os.Args[3]
	} else {
		editorCmd := os.Getenv("EDITOR")
		if editorCmd == "" {
			editorCmd = "vim"
		}

		tmpFile, err := os.CreateTemp("", "pam-new-query-*.sql")
		if err != nil {
			log.Fatalf("Failed to create temp file: %v", err)
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		header := fmt.Sprintf("-- Creating new query: %s\n", queryName)
		header += fmt.Sprintf("-- Connection: %s (%s)\n",
			cfg.CurrentConnection,
			cfg.Connections[cfg.CurrentConnection].DBType)
		header += "-- Write your SQL query below and save\n\n"

		if _, err := tmpFile.Write([]byte(header)); err != nil {
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

		querySQL = removeCommentLines(string(editedData))
		querySQL = strings.TrimSpace(querySQL)

		if querySQL == "" {
			log.Fatal("No SQL query provided. Query not saved.")
		}
	}

	queries[queryName] = db.Query{
		Name: queryName,
		SQL:  querySQL,
		Id:   db.GetNextQueryId(queries),
	}

	err := cfg.Save()
	if err != nil {
		log.Fatal("Could not save configuration file")
	}

	fmt.Printf("✓ Added query '%s' with ID %d\n", queryName, queries[queryName].Id)
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
