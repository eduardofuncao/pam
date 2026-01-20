package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/spinner"
	"github.com/eduardofuncao/pam/internal/table"
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

	// If a table name is provided, run SELECT * FROM table
	if len(args) > 0 {
		tableName := args[0]

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

	// Display tables list using GetInfoSQL
	queryStr := conn.GetInfoSQL("tables")
	if queryStr == "" {
		printError("Could not get tables SQL for this database type")
	}

	// Extract just the name column by wrapping in a subquery
	nameOnlyQuery := fmt.Sprintf(
		"SELECT name FROM (%s) AS tables_info ORDER BY name",
		queryStr,
	)

	if flags.oneline {
		// For oneline mode, execute query and print just the table names
		rows, err := conn.ExecQuery(nameOnlyQuery)
		if err != nil {
			printError("Could not retrieve tables: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				printError("Could not scan row: %v", err)
			}
			fmt.Println(tableName)
		}

		if err := rows.Err(); err != nil {
			printError("Error iterating tables: %v", err)
		}
	} else {
		// For normal mode, use the interactive table viewer with name-only query
		a.showTablesInteractive(conn, nameOnlyQuery)
	}
}

func (a *App) showTablesInteractive(
	conn db.DatabaseConnection,
	queryStr string,
) {
	for {
		start := time.Now()
		done := make(chan struct{})
		go spinner.CircleWaitWithTimer(done)

		rows, err := conn.ExecQuery(queryStr)
		if err != nil {
			done <- struct{}{}
			printError("Could not retrieve tables: %v", err)
		}

		columns, data, err := db.FormatTableData(rows)
		if err != nil {
			done <- struct{}{}
			printError("Could not format table data: %v", err)
		}

		done <- struct{}{}
		elapsed := time.Since(start)

		if len(data) == 0 {
			fmt.Println("No tables found")
			return
		}

		q := db.Query{
			Name: "tables",
			SQL:  queryStr,
		}

		model, err := table.RenderTablesList(
			columns,
			data,
			elapsed,
			conn,
			q,
			a.config.DefaultColumnWidth,
		)
		if err != nil {
			printError("Error rendering tables: %v", err)
		}

		// Check if user selected a table
		selectedTable := model.GetSelectedTableName()
		if selectedTable != "" {
			// User pressed Enter on a table - query it
			query := db.Query{
				Name:      selectedTable,
				SQL:       fmt.Sprintf("SELECT * FROM %s", selectedTable),
				TableName: selectedTable,
				Id:        -1,
			}

			// Try to get primary key from table metadata
			if metadata, err := conn.GetTableMetadata(
				selectedTable,
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
					editedQuery := db.Query{
						Name:      selectedTable,
						SQL:       editedSQL,
						TableName: selectedTable,
						Id:        -1,
					}
					a.executeSelect(
						editedSQL,
						selectedTable,
						conn,
						&editedQuery,
						true,
						nil,
					)
				},
			)
			// After returning from table view, go back to tables list
			continue
		}

		// Check if user wants to re-run (edited the query)
		if model.ShouldRerunQuery() {
			queryStr = model.GetEditedQuery().SQL
			continue
		}

		// User quit without selecting anything
		break
	}
}
