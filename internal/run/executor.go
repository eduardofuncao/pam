package run

import (
	"fmt"
	"os"
	"strings"
	"time"

	stdlib "database/sql"

	"github.com/eduardofuncao/squix/internal/config"
	"github.com/eduardofuncao/squix/internal/db"
	"github.com/eduardofuncao/squix/internal/parser"
	"github.com/eduardofuncao/squix/internal/spinner"
	"github.com/eduardofuncao/squix/internal/styles"
	"github.com/eduardofuncao/squix/internal/table"
)

type SaveQueryCallback func(query db.Query) (db.Query, error)

type ExecutionParams struct {
	Query        db.Query
	Connection   db.DatabaseConnection
	Config       *config.Config
	SaveCallback SaveQueryCallback
	OnRerun      func(editedSQL string) error
	Args         []any  // Arguments for parameterized queries
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
	if len(params.Args) > 0 {
		rows, err = params.Connection.ExecQuery(sql, params.Args...)
	} else {
		rows, err = params.Connection.ExecQuery(sql)
	}
	if err != nil {
		spinner.Stop(done)
		return fmt.Errorf("query execution failed: %w", err)
	}

	// Format the results
	columns, columnTypes, data, err := db.FormatTableDataWithTypes(rows.(*stdlib.Rows))
	if err != nil {
		spinner.Stop(done)
		return fmt.Errorf("formatting failed: %w", err)
	}

	spinner.Stop(done)
	elapsed := time.Since(start)
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

	statusMessage := ""

	for {
		model, err := table.Render(columns, columnTypes, data, elapsed, params.Connection, tableName, primaryKey, q, params.Config.DefaultColumnWidth, params.Config.UIVisibility, params.Config.KeyMap, params.SaveCallback, statusMessage)
		if err != nil {
			return fmt.Errorf("error rendering table: %w", err)
		}

		if !model.ShouldRerunQuery() || params.OnRerun == nil {
			return nil
		}

		err = params.OnRerun(model.GetEditedQuery().SQL)
		if err == nil {
			return nil
		}

		// The rerun failed: track the attempted edit as the current query so the
		// next `e` opens it (and the view shows what was tried), not the last success.
		q.SQL = model.GetEditedQuery().SQL
		statusMessage = formatQueryError(err)
	}
}

func ExecuteNonSelect(params ExecutionParams) {
	start := time.Now()
	done := make(chan struct{})
	go spinner.CircleWaitWithTimer(done)

	var result stdlib.Result
	var err error
	if len(params.Args) > 0 {
		result, err = params.Connection.Exec(params.Query.SQL, params.Args...)
	} else {
		result, err = params.Connection.Exec(params.Query.SQL)
	}
	spinner.Stop(done)
	elapsed := time.Since(start)

	if err != nil {
		printError("Could not execute command: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	noun := "row"
	if rowsAffected != 1 {
		noun = "rows"
	}
	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ %d %s affected in %.2fs", rowsAffected, noun, elapsed.Seconds())))
	fmt.Println(styles.Faint.Render("\nExecuted SQL:"))
	fmt.Println(parser.HighlightSQL(params.Query.SQL))
}

func Execute(params ExecutionParams) error {
	if err := params.Connection.Open(); err != nil {
		return fmt.Errorf("could not open connection to %s/%s: %w", params.Connection.GetDbType(), params.Connection.GetName(), err)
	}
	defer params.Connection.Close()

	return ExecuteWithOpenConn(params)
}

func ExecuteExport(params ExecutionParams, format string) error {
	if err := params.Connection.Open(); err != nil {
		return fmt.Errorf("could not open connection to %s/%s: %w", params.Connection.GetDbType(), params.Connection.GetName(), err)
	}
	defer params.Connection.Close()

	return ExecuteExportWithOpenConn(params, format)
}

func ExecuteExportWithOpenConn(params ExecutionParams, format string) error {
	if IsSelectQuery(params.Query.SQL) {
		return executeExportSelect(params.Query.SQL, params, format)
	}

	// Non-SELECT
	start := time.Now()
	var err error
	if len(params.Args) > 0 {
		_, err = params.Connection.Exec(params.Query.SQL, params.Args...)
	} else {
		_, err = params.Connection.Exec(params.Query.SQL)
	}
	elapsed := time.Since(start)

	if err != nil {
		return fmt.Errorf("could not execute command: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Command executed successfully in %.2fs\n", elapsed.Seconds())
	return nil
}

func executeExportSelect(sql string, params ExecutionParams, format string) error {
	var tableName string
	if params.Query.Id != 0 || params.Query.Name != "" {
		tableName, _ = extractMetadata(params.Connection, params.Query)
	}

	// Apply row limit
	if params.Config.DefaultRowLimit > 0 {
		sql = params.Connection.ApplyRowLimit(sql, params.Config.DefaultRowLimit)
	}

	// Execute query
	var err error
	var rows any
	if len(params.Args) > 0 {
		rows, err = params.Connection.ExecQuery(sql, params.Args...)
	} else {
		rows, err = params.Connection.ExecQuery(sql)
	}
	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}

	columns, _, data, err := db.FormatTableDataWithTypes(rows.(*stdlib.Rows))
	if err != nil {
		return fmt.Errorf("formatting failed: %w", err)
	}

	if len(data) == 0 {
		fmt.Fprintln(os.Stderr, "No results found")
		return nil
	}

	opts := table.FormatOptions{
		QueryName: params.Query.Name,
		DbType:    params.Connection.GetDbType(),
		DbName:    params.Connection.GetName(),
		TableName: tableName,
	}

	content, err := table.FormatExport(columns, data, format, opts)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	fmt.Print(content)
	return nil
}

func ExecuteWithOpenConn(params ExecutionParams) error {
	if IsSelectQuery(params.Query.SQL) {
		return ExecuteSelect(params.Query.SQL, params.Query.Name, params)
	}
	ExecuteNonSelect(params)
	return nil
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

func formatQueryError(err error) string {
	msg := err.Error()
	msg = strings.TrimPrefix(msg, "query execution failed: ")
	msg = strings.TrimPrefix(msg, "error rendering table: ")
	return styles.Error.Render("✗ Query failed: " + msg)
}
