package main

import (
	"fmt"

	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleStatus() {
	if a.config.CurrentConnection == "" {
		fmt.Println(styles.Faint.Render("No active connection"))
		return
	}
	currConn := a.config. Connections[a.config.CurrentConnection]
	fmt. Println(styles.Success.Render("● Currently using: "), styles.Title.Render(fmt.Sprintf("%s/%s", currConn.DBType, a.config.CurrentConnection)))
}
