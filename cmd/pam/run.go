package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/editor"
	"github.com/eduardofuncao/pam/internal/params"
	"github.com/eduardofuncao/pam/internal/run"
)

func (a *App) handleRun() {
	if a.config.CurrentConnection == "" {
		printError("No active connection.   Use 'pam switch <connection>' or 'pam init' first")
	}

	flags := parseRunFlags()
	conn := config.FromConnectionYaml(a.config.Connections[a.config.CurrentConnection])

	resolved, err := run.ResolveQuery(flags, a.config, a.config.CurrentConnection, conn)
	if err != nil {
		printError("%v", err)
	}

	// Check if we need to create a new query via editor
	if run.ShouldCreateNewQuery(resolved) {
		newQuery := a.createNewQueryOrEdit()
		resolved.Query = newQuery
	}

	if flags.EditMode && !flags.LastQuery {
		resolved.Query = a.editQueryOrExit(resolved.Query)
	}

	a.saveIfNeeded(resolved)

	// Parse parameter flags and positional args
	paramFlags := parseParameterFlags()
	positionalArgsSlice := parsePositionalArgs(flags.Selector)

	positionalArgs := params.MapPositionalArgs(resolved.Query.SQL, positionalArgsSlice)

	a.executeQueryWithParams(resolved.Query, conn, paramFlags, positionalArgs)
}

func parseRunFlags() run.Flags {
	flags := run.Flags{}
	args := os.Args[2:]

	for i, arg := range args {
		// Skip parameter flags and their values
		if strings.HasPrefix(arg, "--") && arg != "--edit" && arg != "-e" && arg != "--last" && arg != "-l" {
			// This is a parameter flag, skip it and its value
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				continue
			}
		}

		switch arg {
		case "--edit", "-e":
			flags.EditMode = true
		case "--last", "-l":
			flags.LastQuery = true
		default:
			if !strings.HasPrefix(arg, "--") && flags.Selector == "" {
				flags.Selector = arg
			}
		}
	}
	return flags
}

func parseParameterFlags() map[string]string {
	paramValues := make(map[string]string)
	args := os.Args[2:]

	i := 0
	for i < len(args) {
		arg := args[i]

		// Skip known flags
		if arg == "--edit" || arg == "-e" || arg == "--last" || arg == "-l" {
			i++
			continue
		}

		// Check if it's a parameter flag (--param value)
		if strings.HasPrefix(arg, "--") && len(arg) > 2 {
			paramName := strings.TrimPrefix(arg, "--")

			// If next arg exists and doesn't start with --, it's the value
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				paramValues[paramName] = args[i+1]
				i += 2 // Skip next arg as we've consumed it
			} else {
				// Flag without value - set to empty string (will use default)
				paramValues[paramName] = ""
				i++
			}
		} else {
			// Skip non-flag args (query selector and positional params handled elsewhere)
			i++
		}
	}

	return paramValues
}

// parsePositionalArgs extracts positional parameter values from CLI
// Returns slice of values in order (after query selector)
func parsePositionalArgs(selector string) []string {
	args := os.Args[2:]
	var positionals []string
	selectorSeen := false

	i := 0
	for i < len(args) {
		arg := args[i]

		// Skip flags and their values
		if strings.HasPrefix(arg, "--") {
			// Skip the flag itself
			i++
			// Skip the value if it exists and isn't another flag
			if i < len(args) && !strings.HasPrefix(args[i], "--") {
				i++
			}
			continue
		}
		if arg == "--edit" || arg == "-e" || arg == "--last" || arg == "-l" {
			i++
			continue
		}

		// First non-flag arg is the query selector
		if !selectorSeen {
			selectorSeen = true
			i++
			continue
		}

		// Remaining non-flag args are positional parameter values
		positionals = append(positionals, arg)
		i++
	}

	return positionals
}

func (a *App) createNewQueryOrEdit() db.Query {
	instructions := `-- Enter your SQL run below
-- Save and exit to execute, or exit without saving to cancel
--
`
	editedSQL, err := editor.EditTempFileWithTemplate(instructions, "pam-run-")
	if err != nil {
		printError("Error opening editor: %v", err)
	}
	if editedSQL == "" {
		printError("Empty SQL, cancelled")
	}
	return db.Query{Name: "<runtime>", SQL: editedSQL, Id: -1}
}

func (a *App) editQueryOrExit(query db.Query) db.Query {
	editedSQL, err := editor.EditTempFile(query.SQL, "pam-run-")
	if err != nil {
		printError("Error opening editor: %v", err)
	}
	query.SQL = editedSQL
	return query
}

func (a *App) saveIfNeeded(resolved run.ResolvedQuery) {
	if !resolved.Saveable {
		return
	}

	// Save the query and update last query
	if err := a.config.SaveQueryAndLast(a.config.CurrentConnection, resolved.Query, true); err != nil {
		printError("Failed to save query: %v", err)
	}
}

func (a *App) executeQuery(query db.Query, conn db.DatabaseConnection) {
	originalQuery := query
	run.Execute(run.ExecutionParams{
		Query:        query,
		Connection:   conn,
		Config:       a.config,
		SaveCallback: a.saveQueryFromTable,
		OnRerun: func(editedSQL string) {
			// Re-run callback - for now, just execute directly without parameters
			// If user edits the SQL, they're modifying the substituted SQL
			editedQuery := db.Query{
				Name: originalQuery.Name,
				SQL:  editedSQL,
				Id:   originalQuery.Id,
			}
			run.Execute(run.ExecutionParams{
				Query:        editedQuery,
				Connection:   conn,
				Config:       a.config,
				SaveCallback: a.saveQueryFromTable,
			})
		},
	})
}

func (a *App) executeQueryWithParams(query db.Query, conn db.DatabaseConnection, paramFlags, positionalArgs map[string]string) {
	originalQuery := query

	// Process parameters
	sql, args, displaySQL := a.processParameters(query.SQL, conn, paramFlags, positionalArgs)

	// Create a modified query with processed SQL for execution
	processedQuery := db.Query{
		Name: query.Name,
		SQL:  sql,
		Id:   query.Id,
	}

	// Store original query metadata
	originalQuery.SQL = query.SQL

	run.Execute(run.ExecutionParams{
		Query:        processedQuery,
		Connection:   conn,
		Config:       a.config,
		SaveCallback: a.saveQueryFromTable,
		Args:         args,
		DisplaySQL:   displaySQL,
		OnRerun: func(editedSQL string) {
			// Re-run callback - if SQL contains placeholders, re-process parameters
			// Otherwise execute the edited SQL directly
			finalSQL := editedSQL
			finalArgs := []any{}
			finalDisplaySQL := ""

			if strings.Contains(editedSQL, ":") {
				// User kept the parameter syntax, re-process with same params
				finalSQL, finalArgs, finalDisplaySQL = a.processParameters(originalQuery.SQL, conn, paramFlags, positionalArgs)
			}

			editedQuery := db.Query{
				Name: originalQuery.Name,
				SQL:  finalSQL,
				Id:   originalQuery.Id,
			}
			processedQuery := db.Query{
				Name: editedQuery.Name,
				SQL:  finalSQL,
				Id:   editedQuery.Id,
			}

			run.Execute(run.ExecutionParams{
				Query:        processedQuery,
				Connection:   conn,
				Config:       a.config,
				SaveCallback: a.saveQueryFromTable,
				Args:         finalArgs,
				DisplaySQL:   finalDisplaySQL,
			})
		},
	})
}

// processParameters handles parameter extraction, validation, and substitution
func (a *App) processParameters(sql string, conn db.DatabaseConnection, cliValues, positionals map[string]string) (string, []any, string) {
	// Extract parameter definitions from SQL
	paramDefs := params.ExtractParameters(sql)

	if len(paramDefs) == 0 {
		return sql, []any{}, ""
	}

	// Map positional args to parameter names
	for k, v := range positionals {
		cliValues[k] = v
	}

	// Validate CLI values
	if err := params.ValidateCLIValues(cliValues, paramDefs); err != nil {
		printError("Parameter validation error: %v", err)
	}

	// Validate param names don't conflict with reserved flags
	if err := params.ValidateParamNames(paramDefs); err != nil {
		printError("Parameter name conflict: %v", err)
	}

	// Resolve parameters (CLI > defaults)
	paramValues := params.ResolveParameters(paramDefs, cliValues)

	// Check for missing required parameters
	missing := params.GetMissingRequired(paramDefs, paramValues)
	if len(missing) > 0 {
		// Launch interactive TUI
		collectedValues, err := params.CollectParameters(sql, missing, paramDefs)
		if err != nil {
			if err == params.ErrAborted {
				os.Exit(0)
			}
			printError("Error collecting parameters: %v", err)
		}
		// Merge collected values
		for k, v := range collectedValues {
			paramValues[k] = v
		}
	}

	// Substitute parameters with DB-specific placeholders
	finalSQL, args, err := params.SubstituteParameters(sql, paramValues, conn)
	if err != nil {
		printError("Error substituting parameters: %v", err)
	}

	// Generate display SQL with actual values for TUI
	displaySQL := params.GenerateDisplaySQL(sql, paramValues)

	// For Oracle, use literal substitution instead of prepared statements
	// This is a workaround for godror prepared statement issues
	if conn.GetDbType() == "oracle" && len(args) > 0 {
		finalSQL = substituteOracleLiterals(finalSQL, args)
		args = []any{}
	}

	return finalSQL, args, displaySQL
}

// substituteOracleLiterals replaces :1, :2 placeholders with actual values for Oracle
func substituteOracleLiterals(sql string, args []any) string {
	result := sql
	for i, arg := range args {
		placeholder := fmt.Sprintf(":%d", i+1)
		var value string
		switch v := arg.(type) {
		case string:
			// Escape single quotes in strings
			escaped := strings.ReplaceAll(v, "'", "''")
			value = fmt.Sprintf("'%s'", escaped)
		case int, int32, int64:
			value = fmt.Sprintf("%d", v)
		case float32, float64:
			value = fmt.Sprintf("%f", v)
		default:
			// For other types, try to convert to string
			value = fmt.Sprintf("'%v'", v)
		}
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

func (a *App) saveQueryFromTable(query db.Query) (db.Query, error) {
	connName := a.config.CurrentConnection
	if connName == "" {
		return db.Query{}, fmt.Errorf("no active connection")
	}

	// Save query with auto-ID generation
	savedQuery, err := a.config.SaveQueryToConnection(connName, query)
	if err != nil {
		return db.Query{}, err
	}

	// Update last query
	if err := a.config.UpdateLastQuery(connName, savedQuery); err != nil {
		return db.Query{}, err
	}

	return savedQuery, nil
}
