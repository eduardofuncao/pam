package run

import (
	"fmt"
	"os"
	"time"

	stdlib "database/sql"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/parser"
	"github.com/eduardofuncao/pam/internal/spinner"
	"github.com/eduardofuncao/pam/internal/styles"
	"github.com/eduardofuncao/pam/internal/table"
)

type SaveQueryCallback func(query db.Query) (db.Query, error)

type ExecutionParams struct {
	Query        db.Query
	Connection   db.DatabaseConnection
	Config       *config.Config
	SaveCallback SaveQueryCallback
	OnRerun      func(editedSQL string)
	Args         []any // Arguments for parameterized queries
	DisplaySQL   string // Human-readable SQL with values substituted (for TUI display)
}

func ExecuteSelect(sql, queryName string, params ExecutionParams) error {
	start := time.Now()
	done := make(chan struct{})
	go spinner.CircleWaitWithTimer(done)

	// Extract metadata if query provided
	var tableName, primaryKey string
	var applyRowLimit bool
	if params.Query.Id != 0 || params.Query.Name != "" {
		tableName, primaryKey = extractMetadata(params.Connection, params.Query)
		applyRowLimit = true
	}

	// Apply row limit if requested
	if applyRowLimit && params.Config.DefaultRowLimit > 0 {
		sql = params.Connection.ApplyRowLimit(sql, params.Config.DefaultRowLimit)
	}

	// Execute the query with or without parameters
	var err error
	var rows any
	if params.Args != nil && len(params.Args) > 0 {
		rows, err = params.Connection.ExecQuery(sql, params.Args...)
	} else {
		rows, err = params.Connection.ExecQuery(sql)
	}
	if err != nil {
		done <- struct{}{}
		return fmt.Errorf("query execution failed: %w", err)
	}

	// Format the results
	columns, columnTypes, data, err := db.FormatTableDataWithTypes(rows.(*stdlib.Rows))
	if err != nil {
		done <- struct{}{}
		return fmt.Errorf("formatting failed: %w", err)
	}

	done <- struct{}{}
	elapsed := time.Since(start)

	// Check for empty results
	if len(data) == 0 {
		fmt.Println("No results found")
		return nil
	}

	// Create query object
	q := db.Query{
		Name: queryName,
		SQL:  sql,
	}
	if params.Query.Id != 0 {
		q.Id = params.Query.Id
	}

	// Use DisplaySQL for TUI if available (shows actual values instead of placeholders)
	if params.DisplaySQL != "" {
		q.SQL = params.DisplaySQL
	}

	// Render the TUI
	model, err := table.Render(columns, columnTypes, data, elapsed, params.Connection, tableName, primaryKey, q, params.Config.DefaultColumnWidth, params.SaveCallback)
	if err != nil {
		return fmt.Errorf("error rendering table: %w", err)
	}

	// Handle re-run
	if model.ShouldRerunQuery() && params.OnRerun != nil {
		params.OnRerun(model.GetEditedQuery().SQL)
	}

	return nil
}

func ExecuteNonSelect(params ExecutionParams) {
	start := time.Now()
	done := make(chan struct{})
	go spinner.CircleWaitWithTimer(done)

	var err error
	if params.Args != nil && len(params.Args) > 0 {
		err = params.Connection.Exec(params.Query.SQL, params.Args...)
	} else {
		err = params.Connection.Exec(params.Query.SQL)
	}
	done <- struct{}{}
	elapsed := time.Since(start)

	if err != nil {
		printError("Could not execute command: %v", err)
		return
	}

	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Command executed successfully in %.2fs", elapsed.Seconds())))
	fmt.Println(styles.Faint.Render("\nExecuted SQL:"))
	fmt.Println(parser.HighlightSQL(params.Query.SQL))
}

func Execute(params ExecutionParams) error {
	if err := params.Connection.Open(); err != nil {
		return fmt.Errorf("could not open connection to %s/%s: %w", params.Connection.GetDbType(), params.Connection.GetName(), err)
	}
	defer params.Connection.Close()

	if IsSelectQuery(params.Query.SQL) {
		return ExecuteSelect(params.Query.SQL, params.Query.Name, params)
	} else {
		ExecuteNonSelect(params)
		return nil
	}
}

func extractMetadata(conn db.DatabaseConnection, query db.Query) (string, string) {
	metadata, err := db.InferTableMetadata(conn, query)
	if err == nil && metadata != nil {
		// Return first primary key if available
		pk := ""
		if len(metadata.PrimaryKeys) > 0 {
			pk = metadata.PrimaryKeys[0]
		}
		return metadata.TableName, pk
	}

	fmt.Fprintf(os.Stderr, styles.Faint.Render("Warning: Could not extract table metadata %v\n"), err)
	return "", ""
}

func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
