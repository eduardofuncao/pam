package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/run"
	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleExplore() {
	if len(os.Args) < 3 {
		a.listTablesAndViews()
		return
	}

	tableName := os.Args[2]
	limit := a.config.DefaultRowLimit
	if limit == 0 {
		limit = 1000
	}

	// Parse optional -l/--limit flag
	for i := 3; i < len(os.Args); i++ {
		if os.Args[i] == "-l" || os.Args[i] == "--limit" {
			if i+1 < len(os.Args) {
				parsed, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					printError("Invalid limit value: %s", os.Args[i+1])
				}
				limit = parsed
			}
			break
		}
	}

	if a.config.CurrentConnection == "" {
		printError("No active connection. Use 'pam switch <connection>' or 'pam init' first")
	}

	conn := config.FromConnectionYaml(a.config.Connections[a.config.CurrentConnection])

	if err := conn.Open(); err != nil {
		printError("Could not open connection: %v", err)
	}
	defer conn.Close()

	sql := fmt.Sprintf("SELECT * FROM %s", tableName)
	sql = conn.ApplyRowLimit(sql, limit)

	var onRerun func(string)
	onRerun = func(newSQL string) {
		run.ExecuteSelect(newSQL, tableName, run.ExecutionParams{
			Query:        db.Query{Name: tableName, SQL: newSQL},
			Connection:   conn,
			Config:       a.config,
			OnRerun:      onRerun,
		})
	}
	run.ExecuteSelect(sql, tableName, run.ExecutionParams{
		Query:        db.Query{Name: tableName, SQL: sql},
		Connection:   conn,
		Config:       a.config,
		OnRerun:      onRerun,
	})
}

func (a *App) listTablesAndViews() {
	if a.config.CurrentConnection == "" {
		printError(
			"No active connection. Use 'pam switch <connection>' or 'pam init' first",
		)
	}

	conn := config.FromConnectionYaml(
		a.config.Connections[a.config.CurrentConnection],
	)

	if err := conn.Open(); err != nil {
		printError(
			"Could not open connection to %s: %v",
			a.config.CurrentConnection,
			err,
		)
	}
	defer conn.Close()

	tables, err := conn.GetTables()
	if err != nil {
		printError("Could not list tables: %v", err)
	}

	views, err := conn.GetViews()
	if err != nil {
		printError("Could not list views: %v", err)
	}

	// Display tables
	if len(tables) > 0 {
		fmt.Printf("- tables %s\n", styles.Faint.Render(fmt.Sprintf("(%d)", len(tables))))
		a.formatTableList(tables)
		fmt.Println()
	}

	// Display views
	if len(views) > 0 {
		fmt.Printf("- views %s\n", styles.Faint.Render(fmt.Sprintf("(%d)", len(views))))
		a.formatTableList(views)
		fmt.Println()
	}
}

func (a *App) formatTableList(items []string) {
	if len(items) == 0 {
		return
	}

	// Filter out any empty or whitespace-only items
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	items = filtered

	// Find max item length for column width
	maxLen := 0
	for _, item := range items {
		if len(item) > maxLen {
			maxLen = len(item)
		}
	}

	// Add padding for spacing
	columnWidth := maxLen + 2

	// Simple terminal width detection (default to 120 if we can't detect)
	termWidth := 120

	// Calculate columns that fit
	numColumns := termWidth / columnWidth
	if numColumns < 1 {
		numColumns = 1
	}

	// Print in rows, filling horizontally
	for i, item := range items {
		fmt.Printf("%-*s", columnWidth, item)

		// New line after filling a row
		if (i+1)%numColumns == 0 {
			fmt.Println()
		}
	}

	// New line if last row wasn't complete
	if len(items)%numColumns != 0 {
		fmt.Println()
	}
}

