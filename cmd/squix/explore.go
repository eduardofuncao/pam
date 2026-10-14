package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eduardofuncao/squix/internal/db"
	"github.com/eduardofuncao/squix/internal/explorer"
	"github.com/eduardofuncao/squix/internal/run"
	"github.com/eduardofuncao/squix/internal/spinner"
	"github.com/eduardofuncao/squix/internal/styles"
	"github.com/eduardofuncao/squix/internal/table"
)

type exploreFlags struct {
	oneline    bool
	tablesOnly bool
	viewsOnly  bool
	limit      int
}

func parseExploreFlags() (exploreFlags, []string) {
	flags := exploreFlags{}
	remainingArgs := []string{}
	args := os.Args[2:]

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--oneline", "-o":
			flags.oneline = true
		case "--tables":
			flags.tablesOnly = true
		case "--views":
			flags.viewsOnly = true
		case "--limit", "-l":
			if i+1 < len(args) {
				if parsed, err := strconv.Atoi(args[i+1]); err == nil {
					flags.limit = parsed
				}
				i++
			}
		default:
			if !strings.HasPrefix(arg, "-") {
				remainingArgs = append(remainingArgs, arg)
			}
		}
	}

	return flags, remainingArgs
}

func (a *App) handleExplore() {
	flags, args := parseExploreFlags()

	if len(args) > 0 && isSupportedFile(args[0]) {
		if _, err := os.Stat(args[0]); err != nil {
			printError("File not found: %s", args[0])
		}
		if err := a.handleExploreFile(args[0]); err != nil {
			printError("%v", err)
		}
		return
	}

	if a.config.CurrentConnection == "" {
		printError("No active connection. Use 'squix switch <connection>' or 'squix init' first")
	}

	conn := a.config.LiveConnection(a.config.CurrentConnection)

	if err := conn.Open(); err != nil {
		printError(
			"Could not open connection to %s: %v",
			a.config.CurrentConnection,
			err,
		)
	}
	defer conn.Close()

	// Direct table access: squix explore <table> [-l N]
	if len(args) > 0 {
		a.openTableResults(conn, args[0], flags.limit)
		return
	}

	if flags.oneline {
		a.printTablesAndViews(conn, !flags.viewsOnly, !flags.tablesOnly)
		return
	}

	a.runExplorerLoop(conn, !flags.viewsOnly, !flags.tablesOnly)
}

// runExplorerLoop opens the interactive schema explorer and re-opens it after
// each drill-down action until the user quits.
func (a *App) runExplorerLoop(conn db.DatabaseConnection, showTables, showViews bool) {
	cursor := 0
	for {
		done := make(chan struct{})
		go spinner.CircleWaitWithTimer(done)

		var tables, views []string
		var err error

		if showTables {
			tables, err = conn.GetTables()
			if err != nil {
				printError("Could not list tables: %v", err)
			}
		}
		if showViews {
			views, err = conn.GetViews()
			if err != nil {
				printError("Could not list views: %v", err)
			}
		}

		spinner.Stop(done)

		model, err := explorer.Render(
			conn,
			tables,
			views,
			cursor,
			a.config.UIVisibility,
			a.config.KeyMap,
		)
		if err != nil {
			printError("Error rendering explorer: %v", err)
		}

		cursor = model.Cursor()

		switch model.PendingAction() {
		case explorer.ActionSelectAll:
			a.openTableResults(conn, model.SelectedTable(), 0)
		case explorer.ActionColumns:
			a.openColumnsResults(conn, model.SelectedTable())
		default:
			return
		}
	}
}

// handleReplExplore handles the explore command inside the REPL. With an
// argument it opens the table directly; otherwise it opens the explorer.
func (a *App) handleReplExplore(conn db.DatabaseConnection, args []string) {
	if len(args) > 0 {
		a.openTableResults(conn, args[0], 0)
		return
	}
	a.runExplorerLoop(conn, true, true)
}

func (a *App) printTablesAndViews(conn db.DatabaseConnection, showTables, showViews bool) {
	if showTables {
		tables, err := conn.GetTables()
		if err != nil {
			printError("Could not list tables: %v", err)
		}
		for _, t := range tables {
			fmt.Println(strings.TrimSpace(t))
		}
	}

	if showViews {
		views, err := conn.GetViews()
		if err != nil {
			printError("Could not list views: %v", err)
		}
		for _, v := range views {
			fmt.Println(strings.TrimSpace(v))
		}
	}
}

func (a *App) openTableResults(conn db.DatabaseConnection, tableName string, limit int) {
	sqlQuery := fmt.Sprintf("SELECT * FROM %s", tableName)
	if limit <= 0 {
		limit = a.config.DefaultRowLimit
	}
	if limit > 0 {
		sqlQuery = conn.ApplyRowLimit(sqlQuery, limit)
	}

	query := db.Query{
		Name:      tableName,
		SQL:       sqlQuery,
		TableName: tableName,
		Id:        -1,
	}

	if metadata, err := conn.GetTableMetadata(tableName); err == nil && metadata != nil {
		query.PrimaryKeys = metadata.PrimaryKeys
	}

	var onRerun func(string) error
	onRerun = func(newSQL string) error {
		return run.ExecuteSelect(newSQL, tableName, run.ExecutionParams{
			Query:      db.Query{Name: tableName, SQL: newSQL},
			Connection: conn,
			Config:     a.config,
			OnRerun:    onRerun,
		})
	}

	if err := run.ExecuteSelect(sqlQuery, tableName, run.ExecutionParams{
		Query:        query,
		Connection:   conn,
		Config:       a.config,
		SaveCallback: a.saveQueryFromTable,
		OnRerun:      onRerun,
	}); err != nil {
		printError("%v", err)
	}
}

func (a *App) openColumnsResults(conn db.DatabaseConnection, name string) {
	metadata, err := conn.GetTableMetadata(name)
	if err != nil {
		fmt.Println(styles.Error.Render(fmt.Sprintf("Could not load columns for %s: %v", name, err)))
		return
	}
	if metadata == nil {
		fmt.Println(styles.Error.Render(fmt.Sprintf("No metadata available for %s", name)))
		return
	}

	columns := []string{"column", "type", "pk", "fk", "unique"}

	pkSet := make(map[string]bool, len(metadata.PrimaryKeys))
	for _, pk := range metadata.PrimaryKeys {
		pkSet[pk] = true
	}
	uniqueSet := make(map[string]bool, len(metadata.UniqueConstraints))
	for _, uc := range metadata.UniqueConstraints {
		uniqueSet[uc] = true
	}
	fkMap := make(map[string]string, len(metadata.ForeignKeys))
	for _, fk := range metadata.ForeignKeys {
		fkMap[strings.ToLower(fk.Column)] = fk.ReferencedTable + "." + fk.ReferencedColumn
	}

	data := make([][]string, 0, len(metadata.Columns))
	for i, col := range metadata.Columns {
		typ := ""
		if i < len(metadata.ColumnTypes) {
			typ = metadata.ColumnTypes[i]
		}

		pkMark := ""
		if pkSet[col] {
			pkMark = "✓"
		}

		fkRef := fkMap[strings.ToLower(col)]

		uniqueMark := ""
		if uniqueSet[col] {
			uniqueMark = "✓"
		}

		data = append(data, []string{col, typ, pkMark, fkRef, uniqueMark})
	}
	if len(data) == 0 {
		fmt.Println(styles.Faint.Render(fmt.Sprintf("No columns found for %s", name)))
		return
	}

	vis := a.config.UIVisibility
	vis.QuerySQL = false

	q := db.Query{Name: fmt.Sprintf("%s columns", name)}

	if _, err := table.Render(
		columns,
		nil,
		data,
		0,
		conn,
		"",
		"",
		q,
		a.config.DefaultColumnWidth,
		vis,
		a.config.KeyMap,
		nil,
	); err != nil {
		printError("Error rendering columns: %v", err)
	}
}

var supportedFileTypes = map[string]bool{
	".csv": true,
}

func isSupportedFile(arg string) bool {
	return supportedFileTypes[strings.ToLower(filepath.Ext(arg))]
}

func parseCSV(path string) (columns []string, data [][]string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	columns, err = reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("could not read CSV header: %w", err)
	}

	data, err = reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("could not read CSV data: %w", err)
	}

	return columns, data, nil
}

func (a *App) handleExploreFile(path string) error {
	columns, data, err := parseCSV(path)
	if err != nil {
		return err
	}

	columnTypes := make([]string, len(columns))
	_, err = table.Render(
		columns,
		columnTypes,
		data,
		0,
		nil,
		"",
		"",
		db.Query{},
		a.config.DefaultColumnWidth,
		a.config.UIVisibility,
		a.config.KeyMap,
		nil,
	)
	if err != nil {
		return fmt.Errorf("could not render table: %w", err)
	}

	return nil
}
