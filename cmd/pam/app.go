package main

import (
	"fmt"
	"log"
	"os"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/styles"
)

const Version = "v0.1.0"

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

	if os.Args[1] == "-v" || os.Args[1] == "--version" {
		a.printVersion()
		os.Exit(0)
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
	case "info":
		a.handleInfo()
	case "explore":
		a.handleExplore()
	case "status", "test":
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
	fmt.Println(styles.Title.Render("Pam's database drawer"))
	fmt.Println(styles.Faint.Render("Query manager for your databases"))
	fmt.Println()

	fmt.Println(styles.Title.Render("Quick Start"))
	fmt.Println("  1. Create a connection: " + styles.Faint.Render("pam init <name> <db-type> <connection-string>"))
	fmt.Println("  2. Add a query: " + styles.Faint.Render("pam add <query-name> <sql>"))
	fmt.Println("  3. Run it: " + styles.Faint.Render("pam run <query-name>"))
	fmt.Println()

	fmt.Println(styles.Title.Render("Common Commands"))
	fmt.Println("  pam run <query>      " + styles.Faint.Render("Execute a saved query"))
	fmt.Println("  pam list queries     " + styles.Faint.Render("List saved queries"))
	fmt.Println("  pam list connections " + styles.Faint.Render("List database connections"))
	fmt.Println()

	fmt.Println(styles.Title.Render("Help"))
	fmt.Println("  pam help             " + styles.Faint.Render("Show all commands"))
	fmt.Println("  pam help <command>   " + styles.Faint.Render("Show command details"))
	fmt.Println()
}

func (a *App) printVersion() {
	fmt.Println(styles.Title.Render("Pam's database drawer"))
	fmt.Println(styles.Faint.Render("version: " + Version))
}
