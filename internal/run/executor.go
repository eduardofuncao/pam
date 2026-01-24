package run

import (
	"fmt"
	"os"
	"time"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/parser"
	"github.com/eduardofuncao/pam/internal/spinner"
	"github.com/eduardofuncao/pam/internal/styles"
	"github.com/eduardofuncao/pam/internal/table"
)

// SaveQueryCallback is a function that saves a query and returns the saved query
type SaveQueryCallback func(query db.Query) (db.Query, error)

// ExecutionParams holds all parameters needed for query execution
type ExecutionParams struct {
	Query        db.Query
	Connection   db.DatabaseConnection
	Config       *config.Config
	SaveCallback SaveQueryCallback
	OnRerun      func(editedSQL string)
}

// ExecuteSelect executes SELECT queries and renders results
// Used by both regular queries (with metadata) and info commands (without metadata)
func ExecuteSelect(sql, queryName string, params ExecutionParams) {
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

	// Execute the query
	rows, err := params.Connection.ExecQuery(sql)
	if err != nil {
		done <- struct{}{}
		printError("Could not execute query: %v", err)
		return
	}

	// Format the results
	columns, data, err := db.FormatTableData(rows)
	if err != nil {
		done <- struct{}{}
		printError("Could not format table data: %v", err)
		return
	}

	done <- struct{}{}
	elapsed := time.Since(start)

	// Check for empty results
	if len(data) == 0 {
		fmt.Println("No results found")
		return
	}

	// Create query object
	q := db.Query{
		Name: queryName,
		SQL:  sql,
	}
	if params.Query.Id != 0 {
		q.Id = params.Query.Id
	}

	// Render the TUI
	model, err := table.Render(columns, data, elapsed, params.Connection, tableName, primaryKey, q, params.Config.DefaultColumnWidth, params.SaveCallback)
	if err != nil {
		printError("Error rendering table: %v", err)
		return
	}

	// Handle re-run
	if model.ShouldRerunQuery() && params.OnRerun != nil {
		params.OnRerun(model.GetEditedQuery().SQL)
	}
}

// ExecuteNonSelect executes non-SELECT queries (INSERT, UPDATE, DELETE, etc.)
func ExecuteNonSelect(query db.Query, conn db.DatabaseConnection) {
	start := time.Now()
	done := make(chan struct{})
	go spinner.CircleWaitWithTimer(done)

	err := conn.Exec(query.SQL)
	done <- struct{}{}
	elapsed := time.Since(start)

	if err != nil {
		printError("Could not execute command: %v", err)
		return
	}

	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Command executed successfully in %.2fs", elapsed.Seconds())))
	fmt.Println(styles.Faint.Render("\nExecuted SQL:"))
	fmt.Println(parser.HighlightSQL(query.SQL))
}

// Execute routes to the appropriate executor based on query type
func Execute(params ExecutionParams) {
	if err := params.Connection.Open(); err != nil {
		printError("Could not open connection to %s/%s: %s", params.Connection.GetDbType(), params.Connection.GetName(), err)
		return
	}
	defer params.Connection.Close()

	if IsSelectQuery(params.Query.SQL) {
		ExecuteSelect(params.Query.SQL, params.Query.Name, params)
	} else {
		ExecuteNonSelect(params.Query, params.Connection)
	}
}

// extractMetadata extracts table metadata from a query
func extractMetadata(conn db.DatabaseConnection, query db.Query) (string, string) {
	metadata, err := db.InferTableMetadata(conn, query)
	if err == nil && metadata != nil {
		return metadata.TableName, metadata.PrimaryKey
	}

	fmt.Fprintf(os.Stderr, styles.Faint.Render("Warning: Could not extract table metadata %v\n"), err)
	return "", ""
}

func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
