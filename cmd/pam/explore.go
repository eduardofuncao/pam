package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/eduardofuncao/pam/internal/config"
)

func (a *App) handleExplore() {
	if len(os.Args) < 3 {
		fmt.Println("No table specified. Listing available tables:")
		fmt.Println()
		os.Args = append(os.Args, "tables")
		a.handleInfo()
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
		a.executeSelect(newSQL, tableName, conn, nil, false, onRerun)
	}
	a.executeSelect(sql, tableName, conn, nil, false, onRerun)
}
