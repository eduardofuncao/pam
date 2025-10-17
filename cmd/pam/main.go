package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/godror/godror"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/charmbracelet/lipgloss"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/editor"
	"github.com/eduardofuncao/pam/internal/spinner"
	"github.com/eduardofuncao/pam/internal/table"
)

func main() {
	cfg, err := config.LoadConfig(config.CfgFile)
	if err != nil {
		log.Fatal("Could not load config file", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("pam create <name> <db-type> <connection-string>")
		fmt.Println("pam switch <db-name>")
		fmt.Println("pam add <query-name> <query>")
		fmt.Println("pam query <query-name>")
		fmt.Println("pam get <db-type> <connection-string> <sql-query>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {

	case "init":
		if len(os.Args) < 5 {
			log.Fatal("Usage: pam create <name> <db-type> <connection-string> <user> <password>")
		}

		var conn *db.Connection
		if len(os.Args) < 7 { //no user/pass
			conn = db.NewConnection(os.Args[2], os.Args[3], os.Args[4], "", "")
		} else {
			conn = db.NewConnection(os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6])
		}

		err := conn.Open()
		if err != nil {
			log.Fatalf("Could not establish connection to: %s/%s: %s", conn.DBType, conn.Name, err)
		}
		defer conn.Close()

		cfg.CurrentConnection = conn.Name
		cfg.Connections[cfg.CurrentConnection] = conn
		cfg.Save()

	case "switch", "use":
		if len(os.Args) < 3 {
			log.Fatal("Usage: pam switch/use <db-name>")
		}

		_, ok := cfg.Connections[os.Args[2]]
		if !ok {
			log.Fatalf("Connection %s does not exist", os.Args[2])
		}
		cfg.CurrentConnection = os.Args[2]

		err := cfg.Save()
		if err != nil {
			log.Fatal("Could not save configuration file")
		}
		fmt.Printf("connected to: %s/%s\n", cfg.Connections[cfg.CurrentConnection].DBType, cfg.CurrentConnection)

	case "add":
		if len(os.Args) < 4 {
			log.Fatal("Usage: pam add <query-name> <query>")
		}

		_, ok := cfg.Connections[cfg.CurrentConnection]
		if !ok {
			cfg.Connections[cfg.CurrentConnection] = &db.Connection{}
		}
		cfg.Connections[cfg.CurrentConnection].Queries[os.Args[2]] = db.Query{
			Name: os.Args[2],
			SQL:  os.Args[3],
		}
		err := cfg.Save()
		if err != nil {
			log.Fatal("Could not save configuration file")
		}

	case "query", "run":
		if len(os.Args) < 3 {
			log.Fatal("Usage:pam query/run <query-name>")
		}

		editMode := false
		if len(os.Args) > 3 {
			if os.Args[3] == "--edit" || os.Args[3] == "-e" {
				editMode = true
			}
		}

		currConn := cfg.Connections[cfg.CurrentConnection]
		query := currConn.Queries[os.Args[2]]

		editedQuery, submitted, err := editor.EditQuery(query, editMode)
		if submitted {
			cfg.Connections[cfg.CurrentConnection].Queries[query.Name] = editedQuery
			cfg.Save()
		}

		err = currConn.Open()
		if err != nil {
			log.Fatalf("Could not open the connection to %s/%s: %s", currConn.DBType, currConn.Name, err)
		}

		start := time.Now() 
		done := make(chan struct{})
		go spinner.Wait(done)
		columns, data, err := currConn.Query(query.Name)
		done <- struct{}{}
		elapsed := time.Since(start)
		if err != nil {
			log.Fatal("Could not execute Query:", err)
		}

		if err := table.Render(columns, data, elapsed); err != nil {
			log.Fatalf("Error rendering table: %v", err)
		}

	case "list":
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
			for name, connection := range cfg.Connections {
				fmt.Printf("◆ %s (%s)\n", name, connection.ConnString)
			}

		case "", "queries":
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205"))

			for _, query := range cfg.Connections[cfg.CurrentConnection].Queries {
				fmt.Println(titleStyle.Render("\n◆ " + query.Name))
				fmt.Println(editor.HighlightSQL(editor.FormatSQLWithLineBreaks(query.SQL)))
			}
		}
	case "edit":
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		cmd := exec.Command(editor, config.CfgFile)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to open editor: %v", err)
		}

	case "status":
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("171")).
			Bold(true)
		currConn := cfg.Connections[cfg.CurrentConnection]
		fmt.Println(style.Render("✓ Now using:"), fmt.Sprintf("%s/%s", currConn.DBType, currConn.Name))

	case "history":
		fmt.Println("Not implemented")

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
