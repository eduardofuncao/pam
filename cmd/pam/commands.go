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

func (a *App) handleInit() {
	if len(os.Args) < 5 {
		printError("Usage: pam create <name> <db-type> <connection-string> <user> <password>")
	}

	conn, err := db.CreateConnection(os.Args[2], os.Args[3], os.Args[4])
	if err != nil {
		printError("Could not create connection interface: %s/%s, %s", os.Args[3], os.Args[2], err)
	}

	err = conn.Open()
	if err != nil {
		printError("Could not establish connection to: %s/%s: %s",
			conn.GetDbType(), conn.GetName(), err)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		printError("Could not communicate with the database: %s/%s, %s", os.Args[3], os.Args[2], err)
	}

	a.config.CurrentConnection = conn.GetName()
	a.config.Connections[a.config.CurrentConnection] = config.ToConnectionYAML(conn)
	err = a.config.Save()
	if err != nil {
		printError("Could not save configuration file: %v", err)
	}

	fmt.Println(styles.Success.Render("✓ Connection created:"), styles.Title.Render(fmt.Sprintf("%s/%s", conn.GetDbType(), conn.GetName())))
}

func (a *App) handleSwitch() {
	if len(os.Args) < 3 {
		printError("Usage: pam switch/use <db-name>")
	}

	connName := os.Args[2]
	conn, ok := a.config.Connections[connName]
	if !ok {
		printError("Connection '%s' does not exist", connName)
	}
	a.config.CurrentConnection = connName

	err := a.config.Save()
	if err != nil {
		printError("Could not save configuration file: %v", err)
	}

	fmt.Println(styles.Success.Render("⇄ Switched to:"), styles.Title.Render(fmt.Sprintf("%s/%s", conn.DBType, connName)))
}

func (a *App) handleAdd() {
	if len(os.Args) < 3 {
		printError("Usage: pam add <query-name> [query]")
	}

	if a.config.CurrentConnection == "" {
		printError("No active connection.  Use 'pam switch <connection>' first")
	}

	_, ok := a.config.Connections[a.config.CurrentConnection]
	if !ok {
		a.config.Connections[a.config.CurrentConnection] = &config.ConnectionYAML{}
	}
	queries := a.config.Connections[a.config.CurrentConnection].Queries

	queryName := os.Args[2]
	var querySQL string

	if len(os.Args) >= 4 {
		querySQL = os.Args[3]
	} else {
		editorCmd := os.Getenv("EDITOR")
		if editorCmd == "" {
			editorCmd = "vim"
		}

		tmpFile, err := os.CreateTemp("", "pam-new-query-*. sql")
		if err != nil {
			printError("Failed to create temp file: %v", err)
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		header := fmt.Sprintf("-- Creating new query: %s\n", queryName)
		header += fmt.Sprintf("-- Connection: %s (%s)\n",
			a.config.CurrentConnection,
			a.config.Connections[a.config.CurrentConnection].DBType)
		header += "-- Write your SQL query below and save\n\n"

		if _, err := tmpFile.Write([]byte(header)); err != nil {
			printError("Failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		cmd := exec.Command(editorCmd, tmpPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			printError("Failed to open editor: %v", err)
		}

		editedData, err := os.ReadFile(tmpPath)
		if err != nil {
			printError("Failed to read edited file: %v", err)
		}

		querySQL = removeCommentLines(string(editedData))
		querySQL = strings.TrimSpace(querySQL)

		if querySQL == "" {
			printError("No SQL query provided.  Query not saved")
		}
	}

	queries[queryName] = db.Query{
		Name: queryName,
		SQL:  querySQL,
		Id:   db.GetNextQueryId(queries),
	}

	err := a.config.Save()
	if err != nil {
		printError("Could not save configuration file: %v", err)
	}

	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Added query '%s' with ID %d", queryName, queries[queryName].Id)))
}

func (a *App) handleRemove() {
	if len(os.Args) < 3 {
		printError("Usage: pam remove <query-name>")
	}

	conn := a.config.Connections[a.config.CurrentConnection]
	queries := conn.Queries

	query, exists := db.FindQueryWithSelector(queries, os.Args[2])
	if !exists {
		printError("Query '%s' could not be found", os.Args[2])
	}

	delete(conn.Queries, query.Name)

	err := a.config.Save()
	if err != nil {
		printError("Could not save configuration file: %v", err)
	}

	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Removed query '%s'", query.Name)))
}

func (a *App) handleQuery() {
	editMode := false
	selector := ""

	for _, arg := range os.Args[2:] {
		if arg == "--edit" || arg == "-e" {
			editMode = true
		} else {
			selector = arg
		}
	}

	if a.config.CurrentConnection == "" {
		printError("No active connection.  Use 'pam switch <connection>' first")
	}

	currConn := config.FromConnectionYaml(a.config. Connections[a.config.CurrentConnection])

	var query db.Query
	isInlineSQL := false
	
	// Check if selector looks like an inline SQL query
	if selector != "" && isLikelySQL(selector) {
		isInlineSQL = true
		query = db.Query{
			Name: "<inline>",
			SQL:   selector,
			Id:   -1,
		}
		
		// Allow editing inline SQL if requested
		if editMode {
			editedQuery, submitted, err := editor.EditQuery(query, true)
			if err != nil {
				printError("Error opening editor: %v", err)
			}
			if ! submitted {
				fmt.Println(styles. Faint.Render("Query execution cancelled"))
				return
			}
			query = editedQuery
		}
	} else if selector != "" {
		// Treat as saved query name/id lookup
		queries := currConn.GetQueries()
		q, found := db.FindQueryWithSelector(queries, selector)
		if !found {
			printError("Could not find query with name/id: %v", selector)
		}
		query = q
		
		// Edit the saved query if requested
		if editMode {
			editedQuery, submitted, err := editor.EditQuery(query, true)
			if err != nil {
				printError("Error opening editor: %v", err)
			}
			if submitted {
				a.config.Connections[a. config.CurrentConnection]. Queries[query.Name] = editedQuery
			}
			query = editedQuery
		}
	} else {
		// No selector provided, use last query
		query = a.config.Connections[a. config.CurrentConnection].LastQuery
		if query. Name == "" {
			printError("No last query found.  Usage: pam query/run <query-name|sql>")
		}
		
		if editMode {
			editedQuery, submitted, err := editor.EditQuery(query, true)
			if err != nil {
				printError("Error opening editor: %v", err)
			}
			if submitted && query.Name != "<inline>" {
				a.config.Connections[a.config. CurrentConnection].Queries[query. Name] = editedQuery
			}
			query = editedQuery
		}
	}
	
	// Save last query only if it's not inline
	if !isInlineSQL {
		a.config.Connections[a.config. CurrentConnection].LastQuery = query
		a.config.Save()
	}

	// Display the query that will be executed (if not inline)
	if !isInlineSQL {
		fmt.Printf("%s\n", styles.Title. Render("\n◆ "+query.Name))
		fmt.Println(editor.HighlightSQL(editor.FormatSQLWithLineBreaks(query. SQL)))
		fmt.Println(styles.Separator.Render("──────────────────────────────────────────────────────────"))
	}

	// Open database connection
	if err := currConn.Open(); err != nil {
		printError("Could not open the connection to %s/%s:  %s", 
			currConn. GetDbType(), currConn.GetName(), err)
	}
	defer currConn.Close()

	// Start timing and spinner
	start := time.Now()
	done := make(chan struct{})
	go spinner.Wait(done)

	// Determine query type and execute accordingly
	trimmedSQL := strings.TrimSpace(strings.ToUpper(query.SQL))
	isSelectQuery := strings.HasPrefix(trimmedSQL, "SELECT") ||
		strings.HasPrefix(trimmedSQL, "WITH") ||
		strings.HasPrefix(trimmedSQL, "SHOW") ||
		strings.HasPrefix(trimmedSQL, "DESCRIBE") ||
		strings.HasPrefix(trimmedSQL, "DESC") ||
		strings.HasPrefix(trimmedSQL, "EXPLAIN") ||
		strings.HasPrefix(trimmedSQL, "PRAGMA")

	if isSelectQuery {
		// Handle SELECT queries (returns rows)
		var rows any
		var err error
		
		if isInlineSQL {
			rows, err = currConn.ExecQuery(query.SQL)
		} else {
			rows, err = currConn.Query(query.Name)
		}
		
		if err != nil {
			done <- struct{}{}
			printError("Could not complete query: %v", err)
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
		} else {
			if ! isInlineSQL {
				fmt.Fprintf(os.Stderr, styles.Faint.Render("Warning: Could not extract table metadata:  %v\n"), err)
				fmt.Fprintf(os. Stderr, styles.Faint. Render("Update functionality will be limited.\n"))
			}
		}

		// Render the table view
		if err := table.Render(columns, data, elapsed, currConn, tableName, primaryKeyCol); err != nil {
			printError("Error rendering table: %v", err)
		}
	} else {
		// Handle INSERT/UPDATE/DELETE/CREATE etc. (no rows returned)
		var err error
		
		if isInlineSQL {
			err = currConn. Exec(query.SQL)
		} else {
			// For saved queries, we need to execute the SQL directly
			err = currConn.Exec(query.SQL)
		}
		
		done <- struct{}{}
		elapsed := time.Since(start)
		
		if err != nil {
			printError("Could not execute command: %v", err)
		}
		
		// Show success message
		fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Command executed successfully in %.2fs", elapsed.Seconds())))
		
		// For saved queries, show the SQL that was executed
		if !isInlineSQL {
			fmt. Println(styles. Faint.Render("\nExecuted SQL:"))
			fmt.Println(editor.HighlightSQL(query.SQL))
		}
	}
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

func (a *App) handleList() {
	var objectType string
	if len(os.Args) < 3 {
		objectType = "queries" // Default to queries
	} else {
		objectType = os.Args[2]
	}

	switch objectType {
	case "connections":
		if len(a.config.Connections) == 0 {
			fmt.Println(styles.Faint.Render("No connections configured"))
			return
		}
		for name, connection := range a.config.Connections {
			marker := "◆"
			if name == a.config.CurrentConnection {
				marker = styles.Success.Render("●") // Active connection
			} else {
				marker = styles.Faint.Render("◆")
			}
			fmt.Printf("%s %s %s\n", marker, styles.Title.Render(name), styles.Faint.Render(fmt.Sprintf("(%s)", connection.DBType)))
		}

	case "queries":
		if a.config.CurrentConnection == "" {
			printError("No active connection. Use 'pam switch <connection>' first")
		}
		conn := a.config.Connections[a.config.CurrentConnection]
		if len(conn.Queries) == 0 {
			fmt.Println(styles.Faint.Render("No queries saved"))
			return
		}
		for _, query := range conn.Queries {
			formatedItem := fmt.Sprintf("◆ %d/%s", query.Id, query.Name)
			fmt.Println(styles.Title.Render(formatedItem))
			fmt.Println(editor.HighlightSQL(editor.FormatSQLWithLineBreaks(query.SQL)))
			fmt.Println()
		}

	default:
		printError("Unknown list type: %s. Use 'queries' or 'connections'", objectType)
	}
}

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
		fmt.Println(styles.Success.Render("✓ Queries edited"))
	default:
		printError("Unknown edit type: %s.  Use 'config' or 'queries'", editType)
	}
}

func (a *App) handleStatus() {
	if a.config.CurrentConnection == "" {
		fmt.Println(styles.Faint.Render("No active connection"))
		return
	}
	currConn := a.config.Connections[a.config.CurrentConnection]
	fmt.Println(styles.Success.Render("● Currently using:"), styles.Title.Render(fmt.Sprintf("%s/%s", currConn.DBType, a.config.CurrentConnection)))
}

func (a *App) handleHistory() {
	fmt.Println(styles.Faint.Render("To be implemented in future releases... "))
}

func printError(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, styles.Error.Render("✗ Error:"), msg)
	os.Exit(1)
}

func (a *App) handleHelp() {

	if len(os.Args) == 2 {
		a.PrintGeneralHelp()
	} else {
		a.PrintCommandHelp()
	}
}
