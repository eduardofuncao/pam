package commands

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/eduardofuncao/pam/internal/config"
)

func Status(cfg *config.Config) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("171")).
		Bold(true)
	currConn := cfg.Connections[cfg.CurrentConnection]
	fmt.Println(style.Render("✓ Now using:"), fmt.Sprintf("%s/%s", currConn.DBType, currConn.Name))
}
