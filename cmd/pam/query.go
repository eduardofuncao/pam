package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/editor"
	"github.com/eduardofuncao/pam/internal/spinner"
	"github.com/eduardofuncao/pam/internal/styles"
	"github.com/eduardofuncao/pam/internal/table"
)

type queryFlags struct {
	editMode bool
	newQuery bool
	selector string
}

func (a *App) handleQuery() {
	flags := parseQueryFlags()
	
	if a.config.CurrentConnection == "" {
		printError("No active connection.  Use 'pam switch <connection>' first")
	}

	currConn := config. FromConnectionYaml(a. config.Connections[a.config.CurrentConnection])
	query, isInlineSQL := a.resolveQuery(flags, currConn)
	
	// Handle editing with external editor (skip if query was already edited via --new)
	if flags.editMode && !flags.newQuery {
		var err error
		query, err = a.openExternalEditor(query)
		if err != nil {
			return
		}
		
		// Save changes to saved queries (not inline)
		if query.Name != "<inline>" && query. SQL != "" {
			a. config.Connections[a.config.CurrentConnection].Queries[query.Name] = query
			a.config.Save()
		}
	}
	
	// Save last query only if it's not inline or runtime
	if !isInlineSQL {
		a. config.Connections[a.config.CurrentConnection].LastQuery = query
		a.config.Save()
	}

	// Execute query
	a.executeQuery(query, currConn, isInlineSQL)
}

func parseQueryFlags() queryFlags {
	flags := queryFlags{}
	
	for _, arg := range os. Args[2:] {
		switch arg {
		case "--edit", "-e":
			flags. editMode = true
		case "--new", "-n":
			flags.newQuery = true
		default:
			flags.selector = arg
		}
	}
	
	return flags
}

func (a *App) resolveQuery(flags queryFlags, currConn db.DatabaseConnection) (db.Query, bool) {
	// New query in editor
	if flags.newQuery {
		query := db.Query{
			Name: "<runtime>",
			SQL:   "",
			Id:   -1,
		}
		
		editedQuery, err := a.openExternalEditor(query)
		if err != nil {
			printError("Error opening editor:  %v", err)
		}
		return editedQuery, true
	}
	
	// Inline SQL query
	if flags.selector != "" && isLikelySQL(flags.selector) {
		return db.Query{
			Name: "<inline>",
			SQL:  flags.selector,
			Id:   -1,
		}, true
	}
	
	// Saved query lookup
	if flags.selector != "" {
		queries := currConn.GetQueries()
		q, found := db.FindQueryWithSelector(queries, flags.selector)
		if !found {
			printError("Could not find query with name/id:  %v", flags.selector)
		}
		return q, false
	}
	
	// Last query (default when no selector)
	query := a.config. Connections[a.config.CurrentConnection].LastQuery
	if query.Name == "" {
		printError("No last query found.  Usage: pam run <query-name|sql> or pam run -n for new query")
	}
	return query, false
}

func (a *App) openExternalEditor(query db.Query) (db.Query, error) {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
	}

	tmpFile, err := os.CreateTemp("", "pam-query-*.sql")
	if err != nil {
		printError("Failed to create temp file: %v", err)
		return db.Query{}, err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write current query to temp file
	if _, err := tmpFile.WriteString(query.SQL); err != nil {
		printError("Failed to write to temp file: %v", err)
		return db.Query{}, err
	}
	tmpFile.Close()

	// Open editor
	cmd := exec.Command(editorCmd, tmpPath)
	cmd.Stdin = os.Stdin
	cmd. Stdout = os.Stdout
	cmd.Stderr = os. Stderr
	
	if err := cmd. Run(); err != nil {
		printError("Failed to run editor: %v", err)
		return db.Query{}, err
	}

	// Read edited content
	editedData, err := os.ReadFile(tmpPath)
	if err != nil {
		printError("Failed to read edited file: %v", err)
		return db.Query{}, err
	}

	editedSQL := strings.TrimSpace(string(editedData))
	if editedSQL == "" {
		printError("No SQL query provided")
		return db.Query{}, fmt.Errorf("empty query")
	}

	query.SQL = editedSQL
	return query, nil
}

func (a *App) executeQuery(query db.Query, currConn db.DatabaseConnection, isInlineSQL bool) {
	// Open database connection
	if err := currConn.Open(); err != nil {
		printError("Could not open the connection to %s/%s: %s",
			currConn.GetDbType(), currConn.GetName(), err)
	}
	defer currConn.Close()

	// Start timing and spinner
	start := time.Now()
	done := make(chan struct{})
	go spinner.Wait(done)

	if isSelectQuery(query.SQL) {
		a.executeSelectQuery(query, currConn, isInlineSQL, done, start)
	} else {
		a.executeNonSelectQuery(query, currConn, isInlineSQL, done, start)
	}
}

func (a *App) executeSelectQuery(query db.Query, currConn db.DatabaseConnection, isInlineSQL bool, done chan struct{}, start time.Time) {
	var rows any
	var err error

	if isInlineSQL {
		rows, err = currConn.ExecQuery(query.SQL)
	} else {
		rows, err = currConn.Query(query.Name)
	}

	if err != nil {
		done <- struct{}{}
		printError("Could not complete query:  %v", err)
	}

	columns, data, err := db.FormatTableData(rows. (*sql.Rows))
	if err != nil {
		done <- struct{}{}
		printError("Could not format table data: %v", err)
	}

	done <- struct{}{}
	elapsed := time.Since(start)

	// Try to infer table metadata for update/delete functionality
	metadata, err := db.InferTableMetadata(currConn, query)
	tableName := ""
	primaryKeyCol := ""

	if err == nil && metadata != nil {
		tableName = metadata.TableName
		primaryKeyCol = metadata.PrimaryKey
	} else if ! isInlineSQL {
		fmt.Fprintf(os.Stderr, styles. Faint. Render("Warning: Could not extract table metadata:   %v\n"), err)
		fmt.Fprintf(os. Stderr, styles.Faint. Render("Update functionality will be limited.\n"))
	}

	// Render the table view
	model, err := table. Render(columns, data, elapsed, currConn, tableName, primaryKeyCol, query)
	if err != nil {
		printError("Error rendering table: %v", err)
	}

// Check if user edited and wants to re-run the query
if model.ShouldRerunQuery() {
	editedQuery := model.GetEditedQuery()
	a.executeQuery(editedQuery, currConn, true)
}
}

func (a *App) executeNonSelectQuery(query db.Query, currConn db. DatabaseConnection, isInlineSQL bool, done chan struct{}, start time.Time) {
	err := currConn. Exec(query.SQL)
	done <- struct{}{}
	elapsed := time.Since(start)

	if err != nil {
		printError("Could not execute command: %v", err)
	}

	// Show success message
fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Command executed successfully in %.2fs", elapsed.Seconds())))

	// For saved queries, show the SQL that was executed
	if ! isInlineSQL {
		fmt.Println(styles. Faint.Render("\nExecuted SQL:"))
		fmt.Println(editor.HighlightSQL(query.SQL))
	}
}

func isSelectQuery(sql string) bool {
	trimmedSQL := strings.TrimSpace(strings.ToUpper(sql))
	selectKeywords := []string{
		"SELECT", "WITH", "SHOW", "DESCRIBE", "DESC", "EXPLAIN", "PRAGMA",
	}
	
	for _, keyword := range selectKeywords {
		if strings.HasPrefix(trimmedSQL, keyword+" ") || trimmedSQL == keyword {
			return true
		}
	}
	return false
}

func isLikelySQL(s string) bool {
	upper := strings.ToUpper(strings.TrimSpace(s))

	// Check for common SQL keywords that start queries
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE",
		"CREATE", "DROP", "ALTER", "TRUNCATE",
		"WITH", "SHOW", "DESCRIBE", "DESC", "EXPLAIN",
		"GRANT", "REVOKE", "BEGIN", "COMMIT", "ROLLBACK",
		"PRAGMA", // SQLite specific
	}

	for _, keyword := range sqlKeywords {
		if strings.HasPrefix(upper, keyword+" ") || upper == keyword {
			return true
		}
	}

	return false
}
