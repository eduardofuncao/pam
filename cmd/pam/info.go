package main

import (
	"fmt"
	"os"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/run"
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
		run.ExecuteSelect(sql, "<edited>", run.ExecutionParams{
			Query:        db.Query{Name: "<edited>", SQL: sql},
			Connection:   conn,
			Config:       a.config,
			OnRerun:      onRerun,
		})
	}
	run.ExecuteSelect(queryStr, fmt.Sprintf("info %s", infoType), run.ExecutionParams{
		Query:        db.Query{Name: fmt.Sprintf("info %s", infoType), SQL: queryStr},
		Connection:   conn,
		Config:       a.config,
		OnRerun:      onRerun,
	})
}
