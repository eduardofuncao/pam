package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/styles"
)

type tablesFlags struct {
	oneline bool
}

func parseTablesFlags() (tablesFlags, []string) {
	flags := tablesFlags{}
	remainingArgs := []string{}
	args := os.Args[2:]

	for _, arg := range args {
		if arg == "--oneline" || arg == "-o" {
			flags.oneline = true
		} else if !strings.HasPrefix(arg, "-") {
			remainingArgs = append(remainingArgs, arg)
		}
	}

	return flags, remainingArgs
}

func (a *App) handleTables() {
	if a.config.CurrentConnection == "" {
		printError(
			"No active connection. Use 'pam switch <connection>' or 'pam init' first",
		)
	}

	flags, args := parseTablesFlags()
	conn := config.FromConnectionYaml(
		a.config.Connections[a.config.CurrentConnection],
	)

	if err := conn.Open(); err != nil {
		printError(
			"Could not open connection to %s/%s: %s",
			conn.GetDbType(),
			conn.GetName(),
			err,
		)
	}
	defer conn.Close()

	tables, err := conn.GetTables()
	if err != nil {
		printError("Could not retrieve tables: %v", err)
	}

	if len(tables) == 0 {
		fmt.Println(styles.Faint.Render("No tables found"))
		return
	}

	// If a table name is provided, run SELECT * FROM table
	if len(args) > 0 {
		tableName := args[0]
		found := false
		for _, t := range tables {
			if strings.EqualFold(t, tableName) {
				tableName = t
				found = true
				break
			}
		}

		if !found {
			printError("Table '%s' not found", args[0])
		}

		// Create a temporary query object with table metadata
		query := db.Query{
			Name:      tableName,
			SQL:       fmt.Sprintf("SELECT * FROM %s", tableName),
			TableName: tableName,
			Id:        -1,
		}

		// Try to get primary key from table metadata
		if metadata, err := conn.GetTableMetadata(
			tableName,
		); err == nil &&
			metadata != nil {
			query.PrimaryKey = metadata.PrimaryKey
		}

		a.executeSelect(
			query.SQL,
			query.Name,
			conn,
			&query,
			false,
			func(editedSQL string) {
				// Re-run query if edited
				editedQuery := db.Query{
					Name:      tableName,
					SQL:       editedSQL,
					TableName: tableName,
					Id:        -1,
				}
				a.executeSelect(
					editedSQL,
					tableName,
					conn,
					&editedQuery,
					true,
					nil,
				)
			},
		)
		return
	}

	// Display tables list
	if flags.oneline {
		for _, table := range tables {
			fmt.Println(table)
		}
	} else {
		fmt.Println(
			styles.Title.Render(fmt.Sprintf("Tables in %s", conn.GetName())),
		)
		fmt.Println()
		for i, table := range tables {
			fmt.Printf("%s %s\n",
				styles.Faint.Render(fmt.Sprintf("%d.", i+1)),
				styles.Title.Render(table),
			)
		}
		fmt.Println()
		fmt.Println(
			styles.Faint.Render(fmt.Sprintf("Total: %d tables", len(tables))),
		)
		fmt.Println()
		fmt.Println(
			styles.Faint.Render(
				"Run 'pam tables <table-name>' to query a table",
			),
		)
	}
}
