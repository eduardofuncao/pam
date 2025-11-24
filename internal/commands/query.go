package commands

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/editor"
	"github.com/eduardofuncao/pam/internal/spinner"
	"github.com/eduardofuncao/pam/internal/table"
)

func Query(cfg *config.Config) {
	if len(os.Args) < 3 {
		log.Fatal("Usage:pam query/run <query-name>")
	}

	editMode := false
	if len(os.Args) > 3 {
		if os.Args[3] == "--edit" || os.Args[3] == "-e" {
			editMode = true
		}
	}

	currConn := config.FromConnectionYaml(cfg.Connections[cfg.CurrentConnection])

	queries := currConn.GetQueries()
	selector := os.Args[2]
	query, found := db.FindQueryWithSelector(queries, selector)
	if !found {
		log.Fatalf("Could not find query with name/id: %v", selector)
	}

	editedQuery, submitted, err := editor.EditQuery(query, editMode)
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
