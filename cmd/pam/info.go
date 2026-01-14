package main

import (
	"fmt"
	"os"

	"github.com/eduardofuncao/pam/internal/config"
)

func (a *App) handleInfo() {
	if len(os.Args) < 3 {
		printError("Usage: pam info <tables|views>")
	}

	infoType := os.Args[2]

	if infoType != "tables" && infoType != "views" {
		printError("Unknown info type: %s. Use 'tables' or 'views'", infoType)
	}

	if a.config.CurrentConnection == "" {
		printError("No active connection. Use 'pam switch <connection>' or 'pam init' first")
	}

	conn := config.FromConnectionYaml(a.config.Connections[a.config.CurrentConnection])

	queryStr := conn.GetInfoSQL(infoType)
	if queryStr == "" {
		printError("Could not get SQL for info type: %s", infoType)
	}

	if err := conn.Open(); err != nil {
		printError("Could not open connection: %v", err)
	}
	defer conn.Close()

	var onRerun func(string)
	onRerun = func(sql string) {
		a.executeSelect(sql, "<edited>", conn, nil, false, onRerun)
	}
	a.executeSelect(queryStr, fmt.Sprintf("info %s", infoType), conn, nil, false, onRerun)
}
