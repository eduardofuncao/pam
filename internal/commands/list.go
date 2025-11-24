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
			formatedItem := fmt.Sprintf("\n◆ %d/%s", query.Id, query.Name)
			fmt.Println(titleStyle.Render(formatedItem))
			fmt.Println(editor.HighlightSQL(editor.FormatSQLWithLineBreaks(query.SQL)))
		}
	}
}
