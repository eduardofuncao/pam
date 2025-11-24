package commands

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/editor"
)

func List(cfg *config.Config) {
	var objectType string
	if len(os.Args) < 3 {
		objectType = "queries"
	} else {
		objectType = os.Args[2]
	}

	switch objectType {
	case "connections":
		for name, connection := range cfg.Connections {
			fmt.Printf("◆ %s (%s)\n", name, connection.ConnString)
		}

	case "queries":
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

		for _, query := range cfg.Connections[cfg.CurrentConnection].Queries {
			formatedItem := fmt.Sprintf("\n◆ %d/%s", query.Id, query.Name)
			fmt.Println(titleStyle.Render(formatedItem))
			fmt.Println(editor.HighlightSQL(editor.FormatSQLWithLineBreaks(query.SQL)))
		}

	case "tables":
		ListTables(cfg)

	default:
		log.Fatalf("Unknown list type: %s. Use queries, connections, or tables", objectType)
	}
}

func ListTables(cfg *config.Config) {
	currConn := config.FromConnectionYaml(cfg.Connections[cfg.CurrentConnection])

	err := currConn.Open()
	if err != nil {
		log.Fatalf("Could not open connection: %v", err)
	}
	defer currConn.Close()

	dbType := currConn.GetDbType()
	var querySQL string

	switch dbType {
	case "sqlite3":
		querySQL = "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name"
	case "postgres":
		querySQL = "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename"
	case "mysql":
		querySQL = "SHOW TABLES"
	case "oracle":
		querySQL = "SELECT table_name FROM user_tables ORDER BY table_name"
	default:
		log.Fatalf("Unsupported database type for listing tables: %s", dbType)
	}

	queries := currConn.GetQueries()
	if queries == nil {
		queries = make(map[string]db.Query)
	}
	tempQueryName := "__list_tables_temp__"
	queries[tempQueryName] = db.Query{Name: tempQueryName, SQL: querySQL}
	currConn.SetQueries(queries)

	rows, err := currConn.Query(tempQueryName)
	if err != nil {
		log.Fatalf("Could not list tables: %v", err)
	}

	sqlRows, ok := rows.(*sql.Rows)
	if !ok {
		log.Fatal("Query did not return *sql.Rows")
	}
	defer sqlRows.Close()

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	fmt.Println(titleStyle.Render("\nTables:"))

	for sqlRows.Next() {
		var tableName string
		if err := sqlRows.Scan(&tableName); err != nil {
			log.Fatalf("Error scanning table name: %v", err)
		}
		fmt.Printf("◆ %s\n", tableName)
	}

	if err := sqlRows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}
}
