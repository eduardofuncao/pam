package main

import (
	"fmt"

	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleDisconnect() {
	if a.config.CurrentConnection == "" {
		fmt.Println(styles.Faint.Render("No active connection"))
		return
	}

	previousConnection := a.config.CurrentConnection
	a.config.CurrentConnection = ""

	if err := a.config.Save(); err != nil {
		printError("Could not save config: %v", err)
	}

	fmt.Println(
		styles.Success.Render(
			fmt.Sprintf("✓ Disconnected from '%s'", previousConnection),
		),
	)
	fmt.Println()
	fmt.Println(
		styles.Faint.Render(
			"Use 'pam switch <connection>' to connect to a database",
		),
	)
}
