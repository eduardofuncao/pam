package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/eduardofuncao/pam/internal/config"
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

	tables, err := currConn.ListTables()
	if err != nil {
		log.Fatalf("Could not list tables: %v", err)
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	fmt.Println(titleStyle.Render("\nTables:"))

	for _, tableName := range tables {
		fmt.Printf("◆ %s\n", tableName)
	}
}
