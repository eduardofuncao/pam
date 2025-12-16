package commands

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/db/connections"
	"github.com/eduardofuncao/pam/internal/db/types"
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
		executeExternalEditor(currConn, cfg, cmdExec)
		return nil, nil
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
	query, found := types.FindQueryWithSelector(queries, selector)

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

	start := time.Now()
	var done chan struct{}
	if !fromTUI {
		done = make(chan struct{})
		go spinner.Wait(done)
	}

	desc := &db.QueryDescriptor{
		Type:       "saved_query",
		QueryName:  query.Name,
		SQL:        query.SQL,
		Connection: currConn,
	}

	tableData, err := db.ExecuteQuery(desc)
	if err != nil {
		if !fromTUI {
			done <- struct{}{}
		}
		if fromTUI {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		log.Fatal("Could not complete query: ", err)
	}

	if tableData == nil {
		if !fromTUI {
			done <- struct{}{}
			elapsed := time.Since(start)
			fmt.Printf("\nQuery executed successfully (%.2fs)\n", elapsed.Seconds())
		}
		return nil, nil
	}

	if !fromTUI {
		done <- struct{}{}
		elapsed := time.Since(start)
		if err := table.RenderWithDescriptor(tableData, elapsed, cmdExec, desc); err != nil {
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

func executeExternalEditor(currConn connections.DatabaseConnection, cfg *config.Config, cmdExec table.CommandExecutor) {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		log.Fatal("$EDITOR not set")
	}

	tmpFile, err := os.CreateTemp("", "pam-query-*.sql")
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	tmpFile.Close()

	cmd := exec.Command(editorCmd, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run editor: %v", err)
	}

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		log.Fatalf("Failed to read temp file: %v", err)
	}

	query := strings.TrimSpace(string(content))
	if query == "" {
		fmt.Println("Empty query, nothing to execute")
		return
	}

	executeRawSQLWithArgs(currConn, query, false, cfg, cmdExec)
}

func executeOneShot(currConn connections.DatabaseConnection, cfg *config.Config, cmdExec table.CommandExecutor) {
	emptyQuery := types.Query{Name: "", SQL: ""}
	editedQuery, _, _ := editor.EditQuery(emptyQuery, true)

	start := time.Now()
	done := make(chan struct{})
	go spinner.Wait(done)

	desc := &db.QueryDescriptor{
		Type:       "direct_sql",
		SQL:        editedQuery.SQL,
		Connection: currConn,
	}

	tableData, err := db.ExecuteQuery(desc)
	if err != nil {
		done <- struct{}{}
		log.Fatal("Could not complete query: ", err)
	}

	if tableData == nil {
		done <- struct{}{}
		elapsed := time.Since(start)
		fmt.Printf("\nQuery executed successfully (%.2fs)\n", elapsed.Seconds())
		return
	}

	done <- struct{}{}
	elapsed := time.Since(start)

	if err := table.RenderWithExecutor(tableData, elapsed, cmdExec); err != nil {
		log.Fatalf("Error rendering table: %v", err)
	}
}

func executeRawSQL(currConn connections.DatabaseConnection, query string, cfg *config.Config) {
	executeRawSQLWithArgs(currConn, query, false, cfg, nil)
}

func executeRawSQLWithArgs(currConn connections.DatabaseConnection, query string, fromTUI bool, cfg *config.Config, cmdExec table.CommandExecutor) (*db.TableData, error) {
	start := time.Now()
	var done chan struct{}
	if !fromTUI {
		done = make(chan struct{})
		go spinner.Wait(done)
	}

	desc := &db.QueryDescriptor{
		Type:       "direct_sql",
		SQL:        query,
		Connection: currConn,
	}

	tableData, err := db.ExecuteQuery(desc)
	if err != nil {
		if !fromTUI {
			done <- struct{}{}
		}
		if fromTUI {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		log.Fatal("Could not complete query: ", err)
	}

	if tableData == nil {
		if !fromTUI {
			done <- struct{}{}
			elapsed := time.Since(start)
			fmt.Printf("\nQuery executed successfully (%.2fs)\n", elapsed.Seconds())
		}
		return nil, nil
	}

	if !fromTUI {
		done <- struct{}{}
		elapsed := time.Since(start)
		if err := table.RenderWithDescriptor(tableData, elapsed, cmdExec, desc); err != nil {
			log.Fatalf("Error rendering table: %v", err)
		}
	}

	return tableData, nil
}
