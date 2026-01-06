package main

import (
	"fmt"
	"log"
	"os"

	"github.com/eduardofuncao/pam/internal/config"
)

type App struct {
	config *config.Config
}

func NewApp(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Run() {
	if len(os.Args) < 2 {
		a.printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "init":
		a.handleInit()
	case "switch", "use":
		a.handleSwitch()
	case "add", "save":
		a.handleAdd()
	case "remove", "delete":
		a.handleRemove()
	case "query", "run":
		a.handleQuery()
	case "list":
		a.handleList()
	case "edit":
		a.handleEdit()
	case "status":
		a.handleStatus()
	case "history":
		a.handleHistory()
	case "help":
		a.handleHelp()
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func (a *App) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("pam init <name> <db-type> <connection-string> [schema]")
	fmt. Println("pam switch <db-name>")
	fmt.Println("pam add <query-name> <query>")
	fmt.Println("pam run <query-name> [--edit]")
}
