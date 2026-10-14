package main

import (
	"fmt"
	"log"
	"os"

	"github.com/eduardofuncao/squix/internal/config"
	"github.com/eduardofuncao/squix/internal/styles"
)

var Version = "dev" // overridden via -ldflags at build time

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

	if os.Args[1] == "--drivers" {
		a.handleDrivers()
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
	case "remove", "rm", "delete":
		a.handleRemove()
	case "query", "run":
		a.handleRun()
	case "shell", "repl":
		a.handleShell()
	case "list":
		a.handleList()
	case "ls":
		a.handleListConnections()
	case "edit":
		a.handleEdit()
	case "explore":
		a.handleExplore()
	case "status", "st", "test":
		a.handleStatus()
	case "history":
		a.handleHistory()
	case "disconnect", "clear", "unset":
		a.handleDisconnect()
	case "config":
		a.handleConfig()
	case "explain":
		a.handleExplain()
	case "example":
		a.handleExample()
	case "help":
		a.handleHelp()
	case "__complete":
		a.handleComplete()
	case "completion":
		a.handleCompletion()
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func (a *App) printUsage() {
	fmt.Println(styles.Title.Render("Squix's SQL Stash"))
	fmt.Println(styles.Faint.Render("Query manager for your databases"))
	fmt.Println()

	fmt.Println(styles.Title.Render("Quick Start"))
	fmt.Println(
		"  1. Create a connection: " + styles.Faint.Render(
			"squix init --name mydb --type postgres --conn \"postgres://localhost/db\"",
		),
	)
	fmt.Println(
		"  2. Add a query: " + styles.Faint.Render(
			"squix add <run-name> <sql>",
		),
	)
	fmt.Println("  3. Run it: " + styles.Faint.Render("squix run <run-name>"))
	fmt.Println()

	fmt.Println(styles.Title.Render("Common Commands"))
	fmt.Println(
		"  squix run <run>      " + styles.Faint.Render(
			"Execute a saved query",
		),
	)
	fmt.Println(
		"  squix explore          " + styles.Faint.Render(
			"Explore database schema",
		),
	)
	fmt.Println(
		"  squix explore <table>  " + styles.Faint.Render(
			"Query a table directly",
		),
	)
	fmt.Println(
		"  squix list queries     " + styles.Faint.Render("List saved queries"),
	)
	fmt.Println(
		"  squix shell            " + styles.Faint.Render("Interactive query REPL"),
	)
	fmt.Println(
		"  squix ls               " + styles.Faint.Render(
			"List database connections",
		),
	)
	fmt.Println(
		"  squix --drivers        " + styles.Faint.Render(
			"List database drivers in this build",
		),
	)
	fmt.Println(
		"  squix disconnect       " + styles.Faint.Render(
			"Disconnect from current database",
		),
	)
	fmt.Println()

	fmt.Println(styles.Title.Render("Help"))
	fmt.Println(
		"  squix help             " + styles.Faint.Render("Show all commands"),
	)
	fmt.Println(
		"  squix help <command>   " + styles.Faint.Render("Show command details"),
	)
	fmt.Println()
}

func (a *App) printVersion() {
	fmt.Println(styles.Title.Render("Squix's SQL Stash"))
	fmt.Println(styles.Faint.Render("version: " + Version))
}

func (a *App) handleListConnections() {
	// Set os.Args to simulate "squix list connections"
	originalArgs := os.Args
	os.Args = []string{os.Args[0], "list", "connections"}
	a.handleList()
	os.Args = originalArgs
}
