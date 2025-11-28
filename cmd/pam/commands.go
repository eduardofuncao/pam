package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/editor"
	"github.com/eduardofuncao/pam/internal/spinner"
	"github.com/eduardofuncao/pam/internal/table"
)

func (a *App) handleInit() {
	if len(os.Args) < 5 {
		log.Fatal("Usage: pam create <name> <db-type> <connection-string> <user> <password>")
	}

	conn, err := db.CreateConnection(os.Args[2], os.Args[3], os.Args[4])
	if err != nil {
		log.Fatalf("Could not create connection interface: %s/%s, %s", os.Args[3], os.Args[2], err)
	}

	err = conn.Open()
	if err != nil {
		log.Fatalf("Could not establish connection to: %s/%s: %s",
			conn.GetDbType(), conn.GetName(), err)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		log.Fatalf("Could not communicate with the database: %s/%s, %s", os.Args[3], os.Args[2], err)
	}

	a.config.CurrentConnection = conn.GetName()
	a.config.Connections[a.config.CurrentConnection] = config.ToConnectionYAML(conn)
	a.config.Save()
}

func (a *App) handleSwitch() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: pam switch/use <db-name>")
	}

	_, ok := a.config.Connections[os.Args[2]]
	if !ok {
		log.Fatalf("Connection %s does not exist", os.Args[2])
	}
	a.config.CurrentConnection = os.Args[2]

	err := a.config.Save()
	if err != nil {
		log.Fatal("Could not save configuration file")
	}
	fmt.Printf("connected to: %s/%s\n", a.config.Connections[a.config.CurrentConnection].DBType, a.config.CurrentConnection)
}

func (a *App) handleAdd() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: pam add <query-name> [query]")
	}

	if a.config.CurrentConnection == "" {
		log.Fatal("No active connection. Use 'pam switch <connection>' first")
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

		tmpFile, err := os.CreateTemp("", "pam-new-query-*.sql")
		if err != nil {
			log.Fatalf("Failed to create temp file: %v", err)
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		header := fmt.Sprintf("-- Creating new query: %s\n", queryName)
		header += fmt.Sprintf("-- Connection: %s (%s)\n",
			a.config.CurrentConnection,
			a.config.Connections[a.config.CurrentConnection].DBType)
		header += "-- Write your SQL query below and save\n\n"

		if _, err := tmpFile.Write([]byte(header)); err != nil {
			log.Fatalf("Failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		cmd := exec.Command(editorCmd, tmpPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to open editor: %v", err)
		}

		editedData, err := os.ReadFile(tmpPath)
		if err != nil {
			log.Fatalf("Failed to read edited file: %v", err)
		}

		querySQL = removeCommentLines(string(editedData))
		querySQL = strings.TrimSpace(querySQL)

		if querySQL == "" {
			log.Fatal("No SQL query provided. Query not saved.")
		}
	}

	queries[queryName] = db.Query{
		Name: queryName,
		SQL:  querySQL,
		Id:   db.GetNextQueryId(queries),
	}

	err := a.config.Save()
	if err != nil {
		log.Fatal("Could not save configuration file")
	}

	fmt.Printf("✓ Added query '%s' with ID %d\n", queryName, queries[queryName].Id)
}

func (a *App) handleRemove() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: pam remove <query-name>")
	}

	conn := a.config.Connections[a.config.CurrentConnection]
	queries := conn.Queries

	query, exists := db.FindQueryWithSelector(queries, os.Args[2])
	if exists {
		delete(conn.Queries, query.Name)
	} else {
		log.Fatalf("Query %s could not be found", os.Args[2])
	}
	err := a.config.Save()
	if err != nil {
		log.Fatal("Could not save configuration file")
	}
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

	currConn := config.FromConnectionYaml(a.config.Connections[a.config.CurrentConnection])

	var query db.Query
	if selector != "" {
		queries := currConn.GetQueries()
		q, found := db.FindQueryWithSelector(queries, selector)
		if !found {
			log. Fatalf("Could not find query with name/id: %v", selector)
		}
		query = q
	} else {
		query = a.config.Connections[a. config.CurrentConnection].LastQuery
		if query.Name == "" {
			log.Fatal("No last query found.  Usage: pam query/run <query-name>")
		}
	}

	editedQuery, submitted, err := editor.EditQuery(query, editMode)
	if submitted {
		a. config.Connections[a.config.CurrentConnection].Queries[query.Name] = editedQuery
	}
	a.config. Connections[a.config.CurrentConnection].LastQuery = editedQuery
	a.config.Save()

	if err = currConn.Open(); err != nil {
		log.Fatalf("Could not open the connection to %s/%s: %s", currConn.GetDbType(), currConn.GetName(), err)
	}

	start := time.Now()
	done := make(chan struct{})
	go spinner.Wait(done)

	rows, err := currConn.Query(query.Name)
	if err != nil {
		log.Fatal("Could not complete query: ", err)
	}
	columns, data, err := db.FormatTableData(rows. (*sql.Rows))

	done <- struct{}{}
	elapsed := time.Since(start)

	metadata, err := db.InferTableMetadata(currConn, query)
	tableName := ""
	primaryKeyCol := ""
	
	if err == nil && metadata != nil {
		tableName = metadata.TableName
		primaryKeyCol = metadata.PrimaryKey
	} else {
		fmt. Fprintf(os.Stderr, "Warning: Could not extract table metadata: %v\n", err)
		fmt.Fprintf(os.Stderr, "Update functionality will be limited.\n")
	}

	if err := table.Render(columns, data, elapsed, currConn, tableName, primaryKeyCol); err != nil {
		log.Fatalf("Error rendering table: %v", err)
	}
}

func (a *App) handleList() {
	if len(os.Args) < 3 {
		log.Fatal("Usage:pam list [queries/connections]")
	}

	var objectType string
	if len(os.Args) < 3 {
		objectType = ""
	} else {
		objectType = os.Args[2]
	}

	switch objectType {
	case "connections":
		for name, connection := range a.config.Connections {
			fmt.Printf("◆ %s (%s)\n", name, connection.ConnString)
		}

	case "", "queries":
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

		for _, query := range a.config.Connections[a.config.CurrentConnection].Queries {
			formatedItem := fmt.Sprintf("\n◆ %d/%s", query.Id, query.Name)
			fmt.Println(titleStyle.Render(formatedItem))
			fmt.Println(editor.HighlightSQL(editor.FormatSQLWithLineBreaks(query.SQL)))
		}
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
	case "queries":
		a.editQueries(editorCmd)
	default:
		log.Fatalf("Unknown edit type: %s. Use 'config', 'queries'", editType)
	}
}

func (a *App) handleStatus() {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("171")).
		Bold(true)
	currConn := a.config.Connections[a.config.CurrentConnection]
	fmt.Println(style.Render("✓ Now using:"), fmt.Sprintf("%s/%s", currConn.DBType, currConn.Name))
}

func (a *App) handleHistory() {
	fmt.Println("To be implemented in future releases...")
}
