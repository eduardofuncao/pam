package commands

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/editor"
	"github.com/eduardofuncao/pam/internal/spinner"
	"github.com/eduardofuncao/pam/internal/table"
)

func Run(cfg *config.Config) {
	RunWithArgs(cfg, os.Args, false, nil)
}

func RunWithArgs(cfg *config.Config, args []string, fromTUI bool, cmdExec table.CommandExecutor) (*db.TableData, error) {
	currConn := config.FromConnectionYaml(cfg.Connections[cfg.CurrentConnection])

	if len(args) < 3 {
		if fromTUI {
			return nil, fmt.Errorf("usage: run <query-name|sql>")
		}
		log.Fatal("Usage: pam run <query-name> [--edit|-e]\n       pam run --edit|-e\n       pam run '<raw-sql>'")
	}

	editFlag := hasEditFlagArgs(args)

	if editFlag && len(args) == 3 {
		if fromTUI {
			return nil, fmt.Errorf("--edit flag not supported in TUI mode")
		}
		executeOneShot(currConn, cfg, cmdExec)
		return nil, nil
	}

	selector := args[2]
	queries := currConn.GetQueries()
	query, found := db.FindQueryWithSelector(queries, selector)

	if !found {
		rawSQL := strings.Join(args[2:], " ")
		if looksLikeSQL(rawSQL) {
			return executeRawSQLWithArgs(currConn, rawSQL, fromTUI, cfg, cmdExec)
		}
		if fromTUI {
			return nil, fmt.Errorf("could not find query: %v", selector)
		}
		log.Fatalf("Could not find query with name/id: %v", selector)
	}

	if editFlag && !fromTUI {
		editedQuery, submitted, _ := editor.EditQuery(query, editFlag)
		if submitted {
			cfg.Connections[cfg.CurrentConnection].Queries[query.Name] = editedQuery
			cfg.Save()
		}
	}

	err := currConn.Open()
	if err != nil {
		if fromTUI {
			return nil, fmt.Errorf("could not open connection: %w", err)
		}
		log.Fatalf("Could not open the connection to %s/%s: %s", currConn.GetDbType(), currConn.GetName(), err)
	}

	start := time.Now()
	var done chan struct{}
	if !fromTUI {
		done = make(chan struct{})
		go spinner.Wait(done)
	}

	rows, err := currConn.Query(query.Name)
	if err != nil {
		if !fromTUI {
			done <- struct{}{}
		}
		if fromTUI {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		log.Fatal("Could not complete query: ", err)
	}
	sqlRows, ok := rows.(*sql.Rows)
	if !ok {
		if fromTUI {
			return nil, fmt.Errorf("query did not return *sql.Rows")
		}
		log.Fatal("Query did not return *sql.Rows")
	}

	// Check if query returned any columns
	columns, err := sqlRows.Columns()
	if err != nil || len(columns) == 0 {
		// No columns = DML statement, just show success
		if !fromTUI {
			done <- struct{}{}
			elapsed := time.Since(start)
			fmt.Printf("\nQuery executed successfully (%.2fs)\n", elapsed.Seconds())
		}
		sqlRows.Close()
		return nil, nil
	}

	tableData, err := db.BuildTableData(sqlRows, query.SQL, currConn)
	if err != nil {
		if fromTUI {
			return nil, fmt.Errorf("error building table data: %w", err)
		}
		log.Fatalf("Error building table data: %v", err)
	}

	if !fromTUI {
		done <- struct{}{}
		elapsed := time.Since(start)
		if err := table.RenderWithExecutor(tableData, elapsed, cmdExec); err != nil {
			log.Fatalf("Error rendering table: %v", err)
		}
	}

	return tableData, nil
}

func hasEditFlag() bool {
	return hasEditFlagArgs(os.Args)
}

func hasEditFlagArgs(args []string) bool {
	for _, arg := range args[2:] {
		if arg == "--edit" || arg == "-e" {
			return true
		}
	}
	return false
}

func looksLikeSQL(s string) bool {
	keywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE",
		"CREATE", "ALTER", "DROP", "PRAGMA",
		"WITH", "EXPLAIN", "DESCRIBE", "SHOW",
	}
	upper := strings.ToUpper(strings.TrimSpace(s))
	for _, kw := range keywords {
		if strings.HasPrefix(upper, kw) {
			return true
		}
	}
	return false
}

func executeOneShot(currConn db.DatabaseConnection, cfg *config.Config, cmdExec table.CommandExecutor) {
	emptyQuery := db.Query{Name: "", SQL: ""}
	editedQuery, _, _ := editor.EditQuery(emptyQuery, true)

	if err := currConn.Open(); err != nil {
		log.Fatalf("Could not open the connection: %s", err)
	}

	start := time.Now()
	done := make(chan struct{})
	go spinner.Wait(done)

	rows, err := currConn.QueryDirect(editedQuery.SQL)
	if err != nil {
		done <- struct{}{}
		log.Fatal("Could not complete query: ", err)
	}
	sqlRows, ok := rows.(*sql.Rows)
	if !ok {
		log.Fatal("Query did not return *sql.Rows")
	}

	// Check if query returned any columns
	columns, err := sqlRows.Columns()
	if err != nil || len(columns) == 0 {
		// No columns = DML statement, just show success
		done <- struct{}{}
		elapsed := time.Since(start)
		fmt.Printf("\nQuery executed successfully (%.2fs)\n", elapsed.Seconds())
		sqlRows.Close()
		return
	}

	tableData, err := db.BuildTableData(sqlRows, editedQuery.SQL, currConn)
	if err != nil {
		log.Fatalf("Error building table data: %v", err)
	}

	done <- struct{}{}
	elapsed := time.Since(start)

	if err := table.RenderWithExecutor(tableData, elapsed, cmdExec); err != nil {
		log.Fatalf("Error rendering table: %v", err)
	}
}

func executeRawSQL(currConn db.DatabaseConnection, query string, cfg *config.Config) {
	executeRawSQLWithArgs(currConn, query, false, cfg, nil)
}

func executeRawSQLWithArgs(currConn db.DatabaseConnection, query string, fromTUI bool, cfg *config.Config, cmdExec table.CommandExecutor) (*db.TableData, error) {
	if err := currConn.Open(); err != nil {
		if fromTUI {
			return nil, fmt.Errorf("could not open connection: %w", err)
		}
		log.Fatalf("Could not open the connection: %s", err)
	}

	start := time.Now()
	var done chan struct{}
	if !fromTUI {
		done = make(chan struct{})
		go spinner.Wait(done)
	}

	rows, err := currConn.QueryDirect(query)
	if err != nil {
		if !fromTUI {
			done <- struct{}{}
		}
		if fromTUI {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		log.Fatal("Could not complete query: ", err)
	}
	sqlRows, ok := rows.(*sql.Rows)
	if !ok {
		if fromTUI {
			return nil, fmt.Errorf("query did not return *sql.Rows")
		}
		log.Fatal("Query did not return *sql.Rows")
	}

	// Check if query returned any columns
	columns, err := sqlRows.Columns()
	if err != nil || len(columns) == 0 {
		// No columns = DML statement, just show success
		if !fromTUI {
			done <- struct{}{}
			elapsed := time.Since(start)
			fmt.Printf("\nQuery executed successfully (%.2fs)\n", elapsed.Seconds())
		}
		sqlRows.Close()
		return nil, nil
	}

	tableData, err := db.BuildTableData(sqlRows, query, currConn)
	if err != nil {
		if fromTUI {
			return nil, fmt.Errorf("error building table data: %w", err)
		}
		log.Fatalf("Error building table data: %v", err)
	}

	if !fromTUI {
		done <- struct{}{}
		elapsed := time.Since(start)
		if err := table.RenderWithExecutor(tableData, elapsed, cmdExec); err != nil {
			log.Fatalf("Error rendering table: %v", err)
		}
	}

	return tableData, nil
}
