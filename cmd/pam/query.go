package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/parser"
	"github.com/eduardofuncao/pam/internal/spinner"
	"github.com/eduardofuncao/pam/internal/styles"
	"github.com/eduardofuncao/pam/internal/table"
)

type queryFlags struct {
	editMode  bool
	lastQuery bool
	selector  string
}

type resolvedQuery struct {
	query    db.Query
	saveable bool // will be saved to config file
}

func (a *App) handleQuery() {
	if a.config.CurrentConnection == "" {
		printError("No active connection.   Use 'pam switch <connection>' or 'pam init' first")
	}

	flags := parseQueryFlags()
	conn := config.FromConnectionYaml(a.config.Connections[a.config.CurrentConnection])

	resolved := a.resolveQuery(flags, conn)

	if flags.editMode && !flags.lastQuery {
		resolved.query = a.editQueryOrExit(resolved.query)
	}

	a.saveIfNeeded(resolved)
	a.executeQuery(resolved.query, conn, !resolved.saveable)
}

func parseQueryFlags() queryFlags {
	flags := queryFlags{}
	for _, arg := range os.Args[2:] {
		switch arg {
		case "--edit", "-e":
			flags.editMode = true
		case "--last", "-l":
			flags.lastQuery = true
		default:
			flags.selector = arg
		}
	}
	return flags
}

func (a *App) resolveQuery(flags queryFlags, conn db.DatabaseConnection) resolvedQuery {
	// Priority 1: Last query with --last/-l flag
	if flags.lastQuery {
		lastQuery := a.config.Connections[a.config.CurrentConnection].LastQuery
		if lastQuery.Name == "" {
			printError("No last query found. Run a query first, then use pam run --last")
		}
		return resolvedQuery{
			query:    lastQuery,
			saveable: true,
		}
	}

	// Priority 2: Inline SQL (pam run "select * from employees")
	if flags.selector != "" && isLikelySQL(flags.selector) {
		return resolvedQuery{
			query:    db.Query{Name: "<inline>", SQL: flags.selector, Id: -1},
			saveable: false,
		}
	}

	// Priority 3: Saved query by name/ID
	if flags.selector != "" {
		q, found := db.FindQueryWithSelector(conn.GetQueries(), flags.selector)
		if !found {
			printError("Could not find query with name/id: %v", flags.selector)
		}
		return resolvedQuery{
			query:    q,
			saveable: true,
		}
	}

	//  Default - create new query in editor (pam run with no args)
	return resolvedQuery{
		query:    a.createNewQueryOrExit(),
		saveable: false,
	}
}

func (a *App) createNewQueryOrExit() db.Query {
	instructions := `-- Enter your SQL query below
-- Save and exit to execute, or exit without saving to cancel
--
`
	query := db.Query{Name: "<runtime>", SQL: instructions, Id: -1}
	edited, err := a.openExternalEditor(query)
	if err != nil {
		printError("Error opening editor: %v", err)
	}
	return edited
}

func (a *App) editQueryOrExit(query db.Query) db.Query {
	edited, err := a.openExternalEditor(query)
	if err != nil {
		printError("Error opening editor: %v", err)
	}
	return edited
}

func (a *App) saveIfNeeded(resolved resolvedQuery) {
	if !resolved.saveable {
		return
	}

	connData := a.config.Connections[a.config.CurrentConnection]

	// Save edited query back to config
	if resolved.query.Name != "<inline>" && resolved.query.SQL != "" {
		connData.Queries[resolved.query.Name] = resolved.query
	}

	// Save as last query
	connData.LastQuery = resolved.query
	a.config.Save()
}

func (a *App) openExternalEditor(query db.Query) (db.Query, error) {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
	}

	tmpFile, err := os.CreateTemp("", "pam-query-*.sql")
	if err != nil {
		return db.Query{}, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(query.SQL); err != nil {
		return db.Query{}, fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command(editorCmd, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return db.Query{}, fmt.Errorf("run editor: %w", err)
	}

	editedData, err := os.ReadFile(tmpPath)
	if err != nil {
		return db.Query{}, fmt.Errorf("read edited file: %w", err)
	}

	editedSQL := strings.TrimSpace(string(editedData))

	// Strip instructions if present (for new queries)
	if strings.HasPrefix(editedSQL, "-- Enter your SQL query below") {
		lines := strings.Split(editedSQL, "\n")
		var sqlLines []string
		foundSeparator := false
		for _, line := range lines {
			if foundSeparator {
				sqlLines = append(sqlLines, line)
			} else if line == "--" {
				foundSeparator = true
			}
		}
		editedSQL = strings.TrimSpace(strings.Join(sqlLines, "\n"))
	}

	if editedSQL == "" {
		return db.Query{}, fmt.Errorf("empty query")
	}

	query.SQL = editedSQL
	return query, nil
}

func (a *App) executeQuery(query db.Query, conn db.DatabaseConnection, isInline bool) {
	if err := conn.Open(); err != nil {
		printError("Could not open connection to %s/%s: %s", conn.GetDbType(), conn.GetName(), err)
	}
	defer conn.Close()

	if isSelectQuery(query.SQL) {
		originalQuery := query
		a.executeSelect(query.SQL, query.Name, conn, &query, isInline, func(editedSQL string) {
			editedQuery := db.Query{
				Name: originalQuery.Name,
				SQL:  editedSQL,
				Id:   originalQuery.Id,
			}
			a.executeQuery(editedQuery, conn, true)
		})
	} else {
		a.executeNonSelect(query, conn, isInline)
	}
}

// executeSelect executes SELECT queries and renders results
// Used by both regular queries (with metadata) and info commands (without metadata)
func (a *App) executeSelect(sql, queryName string, conn db.DatabaseConnection, query *db.Query, isInline bool, onRerun func(string)) {
	start := time.Now()
	done := make(chan struct{})
	go spinner.CircleWaitWithTimer(done)

	// Extract metadata if query provided
	var tableName, primaryKey string
	var applyRowLimit bool
	if query != nil {
		tableName, primaryKey = a.extractMetadata(conn, *query, isInline)
		applyRowLimit = true
	}

	// Apply row limit if requested
	if applyRowLimit && a.config.DefaultRowLimit > 0 {
		sql = conn.ApplyRowLimit(sql, a.config.DefaultRowLimit)
	}

	// Execute the query
	rows, err := conn.ExecQuery(sql)
	if err != nil {
		done <- struct{}{}
		printError("Could not execute query: %v", err)
	}

	// Format the results
	columns, data, err := db.FormatTableData(rows)
	if err != nil {
		done <- struct{}{}
		printError("Could not format table data: %v", err)
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
	if query != nil {
		q.Id = query.Id
	}

	// Render the TUI
	model, err := table.Render(columns, data, elapsed, conn, tableName, primaryKey, q, a.config.DefaultColumnWidth, a.saveQueryFromTable)
	if err != nil {
		printError("Error rendering table: %v", err)
	}

	// Handle re-run
	if model.ShouldRerunQuery() {
		onRerun(model.GetEditedQuery().SQL)
	}
}

func (a *App) extractMetadata(conn db.DatabaseConnection, query db.Query, isInline bool) (string, string) {
	metadata, err := db.InferTableMetadata(conn, query)
	if err == nil && metadata != nil {
		return metadata.TableName, metadata.PrimaryKey
	}

	if !isInline {
		fmt.Fprintf(os.Stderr, styles.Faint.Render("Warning: Could not extract table metadata %v\n"), err)
	}

	return "", ""
}

func (a *App) executeNonSelect(query db.Query, conn db.DatabaseConnection, isInline bool) {
	start := time.Now()
	done := make(chan struct{})
	go spinner.CircleWaitWithTimer(done)

	err := conn.Exec(query.SQL)
	done <- struct{}{}
	elapsed := time.Since(start)

	if err != nil {
		printError("Could not execute command: %v", err)
	}

	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Command executed successfully in %.2fs", elapsed.Seconds())))

	if !isInline {
		fmt.Println(styles.Faint.Render("\nExecuted SQL:"))
		fmt.Println(parser.HighlightSQL(query.SQL))
	}
}

func isSelectQuery(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	keywords := []string{"SELECT", "WITH", "SHOW", "DESCRIBE", "DESC", "EXPLAIN", "PRAGMA"}

	for _, kw := range keywords {
		if upper == kw || strings.HasPrefix(upper, kw+" ") {
			return true
		}
	}
	return false
}

func isLikelySQL(s string) bool {
	upper := strings.ToUpper(strings.TrimSpace(s))
	keywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER", "TRUNCATE",
		"WITH", "SHOW", "DESCRIBE", "DESC", "EXPLAIN", "GRANT", "REVOKE",
		"BEGIN", "COMMIT", "ROLLBACK", "PRAGMA",
	}

	for _, kw := range keywords {
		if upper == kw || strings.HasPrefix(upper, kw+" ") {
			return true
		}
	}
	return false
}

func (a *App) saveQueryFromTable(query db.Query) (db.Query, error) {
	connName := a.config.CurrentConnection
	if connName == "" {
		return db.Query{}, fmt.Errorf("no active connection")
	}

	connData := a.config.Connections[connName]

	// Check if query with this name already exists (and we're creating new)
	if query.Id == -1 {
		if _, exists := connData.Queries[query.Name]; exists {
			return db.Query{}, fmt.Errorf("query '%s' already exists", query.Name)
		}
		// Get next ID
		maxID := 0
		for _, q := range connData.Queries {
			if q.Id > maxID {
				maxID = q.Id
			}
		}
		query.Id = maxID + 1
	}

	// Save the query
	connData.Queries[query.Name] = query

	// Update last query
	connData.LastQuery = query

	// Save config
	if err := a.config.Save(); err != nil {
		return db.Query{}, err
	}

	return query, nil
}
