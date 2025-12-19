package main

import (
	"fmt"
	"os"

	"github.com/eduardofuncao/pam/internal/editor"
	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleList() {
	var objectType string
	if len(os.Args) < 3 {
		objectType = "queries" // Default to queries
	} else {
		objectType = os.Args[2]
	}

	switch objectType {
	case "connections":
		if len(a.config.Connections) == 0 {
			fmt.Println(styles. Faint.Render("No connections configured"))
			return
		}
		for name, connection := range a.config.Connections {
			marker := "◆"
			if name == a.config.CurrentConnection {
				marker = styles.Success.Render("●") // Active connection
			} else {
				marker = styles.Faint.Render("◆")
			}
			fmt.Printf("%s %s %s\n", marker, styles.Title.Render(name), styles.Faint.Render(fmt.Sprintf("(%s)", connection.DBType)))
		}

	case "queries":
		if a.config.CurrentConnection == "" {
			printError("No active connection.  Use 'pam switch <connection>' first")
		}
		conn := a.config.Connections[a.config.CurrentConnection]
		if len(conn.Queries) == 0 {
			fmt. Println(styles.Faint. Render("No queries saved"))
			return
		}
		for _, query := range conn.Queries {
			formatedItem := fmt.Sprintf("◆ %d/%s", query.Id, query.Name)
			fmt. Println(styles.Title.Render(formatedItem))
			fmt.Println(editor. HighlightSQL(editor.FormatSQLWithLineBreaks(query.SQL)))
			fmt.Println()
		}

	default: 
		printError("Unknown list type: %s.  Use 'queries' or 'connections'", objectType)
	}
}
