package commands

import (
	"database/sql"
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
	currConn := config.FromConnectionYaml(cfg.Connections[cfg.CurrentConnection])

	if len(os.Args) < 3 {
		log.Fatal("Usage: pam run <query-name> [--edit|-e]\n       pam run --edit|-e\n       pam run '<raw-sql>'")
	}

	editFlag := hasEditFlag()

	if editFlag && len(os.Args) == 3 {
		executeOneShot(currConn)
		return
	}

	selector := os.Args[2]
	queries := currConn.GetQueries()
	query, found := db.FindQueryWithSelector(queries, selector)

	if !found {
		rawSQL := strings.Join(os.Args[2:], " ")
		if looksLikeSQL(rawSQL) {
			executeRawSQL(currConn, rawSQL)
			return
		}
		log.Fatalf("Could not find query with name/id: %v", selector)
	}

	editedQuery, submitted, err := editor.EditQuery(query, editFlag)
	if submitted {
		cfg.Connections[cfg.CurrentConnection].Queries[query.Name] = editedQuery
		cfg.Save()
	}

	err = currConn.Open()
	if err != nil {
		log.Fatalf("Could not open the connection to %s/%s: %s", currConn.GetDbType(), currConn.GetName(), err)
	}

	start := time.Now()
	done := make(chan struct{})
	go spinner.Wait(done)

	rows, err := currConn.Query(query.Name)
	if err != nil {
		log.Fatal("Could not complete query: ", err)
	}
	sqlRows, ok := rows.(*sql.Rows)
	if !ok {
		log.Fatal("Query did not return *sql.Rows")
	}
	columns, data, err := db.FormatTableData(sqlRows)

	done <- struct{}{}
	elapsed := time.Since(start)

	if err := table.Render(columns, data, elapsed); err != nil {
		log.Fatalf("Error rendering table: %v", err)
	}
}

func hasEditFlag() bool {
	for _, arg := range os.Args[2:] {
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

func executeOneShot(currConn db.DatabaseConnection) {
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
		log.Fatal("Could not complete query: ", err)
	}
	sqlRows, ok := rows.(*sql.Rows)
	if !ok {
		log.Fatal("Query did not return *sql.Rows")
	}
	columns, data, err := db.FormatTableData(sqlRows)

	done <- struct{}{}
	elapsed := time.Since(start)

	if err := table.Render(columns, data, elapsed); err != nil {
		log.Fatalf("Error rendering table: %v", err)
	}
}

func executeRawSQL(currConn db.DatabaseConnection, query string) {
	if err := currConn.Open(); err != nil {
		log.Fatalf("Could not open the connection: %s", err)
	}

	start := time.Now()
	done := make(chan struct{})
	go spinner.Wait(done)

	rows, err := currConn.QueryDirect(query)
	if err != nil {
		log.Fatal("Could not complete query: ", err)
	}
	sqlRows, ok := rows.(*sql.Rows)
	if !ok {
		log.Fatal("Query did not return *sql.Rows")
	}
	columns, data, err := db.FormatTableData(sqlRows)

	done <- struct{}{}
	elapsed := time.Since(start)

	if err := table.Render(columns, data, elapsed); err != nil {
		log.Fatalf("Error rendering table: %v", err)
	}
}
