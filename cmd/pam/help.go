package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleHelp() {
	if len(os.Args) == 2 {
		a.PrintGeneralHelp()
	} else {
		a.PrintCommandHelp()
	}
}

func (a *App) PrintGeneralHelp() {
	// Header
	fmt.Println(
		styles.Title.Render(
			"Pam's database drawer - query manager for your databases",
		),
	)
	fmt.Println(
		styles.Faint.Render(
			"Save, edit, and run named SQL queries across connections.",
		),
	)
	fmt.Println()

	// Usage
	fmt.Println(styles.Title.Render("Usage"))
	fmt.Println(styles.Separator.Render("  pam <command> [arguments]"))
	fmt.Println()

	// Commands
	fmt.Println(styles.Title.Render("Commands"))
	fmt.Println(
		"  init        " + styles.Faint.Render(
			"Create and Save a new database connection",
		),
	)
	fmt.Println(
		"  switch      " + styles.Faint.Render(
			"Switch the active connection (alias: use)",
		),
	)
	fmt.Println(
		"  disconnect  " + styles.Faint.Render(
			"Disconnect from the current database",
		),
	)
	fmt.Println(
		"  add         " + styles.Faint.Render(
			"Save a new named query (alias: save)",
		),
	)
	fmt.Println(
		"  remove      " + styles.Faint.Render(
			"Remove a saved query by name or id (alias: delete)",
		),
	)
	fmt.Println(
		"  run         " + styles.Faint.Render(
			"Run a saved query by name or id (alias: query)",
		),
	)
	fmt.Println(
		"  tables      " + styles.Faint.Render("List or query database tables"),
	)
	fmt.Println(
		"  list        " + styles.Faint.Render("List connections or queries"),
	)
	fmt.Println(
		"  info        " + styles.Faint.Render(
			"Show tables or views in current connection",
		),
	)
	fmt.Println(
		"  edit        " + styles.Faint.Render(
			"Open config or queries in your editor",
		),
	)
	fmt.Println(
		"  status      " + styles.Faint.Render(
			"Show the current active connection",
		),
	)
	fmt.Println(
		"  history     " + styles.Faint.Render(
			"Show query history (not implemented yet)",
		),
	)
	fmt.Println(
		"  help        " + styles.Faint.Render(
			"Show help for pam or a specific command",
		),
	)
	fmt.Println()

	// Short help
	fmt.Println(styles.Title.Render("Help"))
	fmt.Println(
		"  pam help              " + styles.Faint.Render("Show this help"),
	)
	fmt.Println(
		"  pam help <command>    " + styles.Faint.Render(
			"Show detailed help for a specific command",
		),
	)
	fmt.Println()

	// Examples
	fmt.Println(styles.Title.Render("Examples"))
	fmt.Println(
		"  pam init dev postgres \"postgres://user:pass@localhost:5432/dbname\"",
	)
	fmt.Println(
		"  pam init prod sqlserver \"sqlserver://sa:password@localhost:1433?database=mydb\"",
	)
	fmt.Println("  pam switch dev")
	fmt.Println("  pam add list_users \"SELECT * FROM users\"")
	fmt.Println("  pam run list_users")
	fmt.Println(" pam run \"select * from users\"")
	fmt.Println("  pam list connections")
	fmt.Println("  pam list queries")
	fmt.Println("  pam edit config")
	fmt.Println("  pam edit queries")
}

func (a *App) PrintCommandHelp() {
	if len(os.Args) < 3 {
		a.PrintGeneralHelp()
		return
	}

	cmd := strings.ToLower(os.Args[2])

	section := func(title string) {
		fmt.Println(styles.Title.Render(title))
	}

	switch cmd {
	case "init", "create":
		section("Command:  init")
		fmt.Println(
			styles.Faint.Render(
				"Create and validate a new database connection configuration. ",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  pam init <name> <db-type> <connection-string> [schema]")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  - Opens and pings the database to verify the connection.",
		)
		fmt.Println("  - Saves the configuration if everything succeeds.")
		fmt.Println(
			"  - For Oracle databases, optionally specify a schema to set as default.",
		)
		fmt.Println()
		section("Examples")
		fmt.Println(
			"  pam init dev postgres \"postgres://user:pass@localhost:5432/dbname\"",
		)
		fmt.Println(
			"  pam init prod sqlserver \"sqlserver://sa:password@localhost:1433?database=mydb\"",
		)
		fmt.Println(
			"  pam init staging mysql \"user:pass@tcp(127.0.0.1:3306)/dbname\"",
		)

	case "switch", "use":
		section("Command: switch")
		fmt.Println(
			styles.Faint.Render(
				"Switch the active connection used by other commands.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  pam switch/use <connection-name>")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  - Sets the connection to be used by 'add', 'run', 'list queries', etc.",
		)
		fmt.Println()
		section("Examples")
		fmt.Println("  pam switch dev")
		fmt.Println("  pam use prod")

	case "add", "save":
		section("Command: add")
		fmt.Println(
			styles.Faint.Render(
				"Save a new named query under the current connection.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  pam add <query-name> [query]")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  - If [query] is omitted, pam opens $EDITOR (default: vim) so you",
		)
		fmt.Println("    can write the query interactively.")
		fmt.Println("  - Each query gets a numeric ID as well as a name.")
		fmt.Println("  - Requires an active connection (use 'pam switch').")
		fmt.Println()
		section("Examples")
		fmt.Println("  pam add list_users \"SELECT * FROM users\"")
		fmt.Println("  pam add update_status    # opens editor to write SQL")

	case "remove", "delete":
		section("Command: remove")
		fmt.Println(styles.Faint.Render("Remove a saved query by name or ID."))
		fmt.Println()
		section("Usage")
		fmt.Println("  pam remove <query-name-or-id>")
		fmt.Println()
		section("Examples")
		fmt.Println("  pam remove list_users")
		fmt.Println("  pam remove 3")

	case "run", "query":
		section("Command: run")
		fmt.Println(
			styles.Faint.Render(
				"Execute a saved query against the current connection and show results in an interactive table view.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  pam run <query-name-or-id> [--edit | -e] [--last | -l]")
		fmt.Println("  pam run                      " + styles.Faint.Render("# Opens the editor to build sql query"))
		fmt.Println()
		section("Description")
		fmt.Println(
			"  - Looks up a saved query by name or numeric ID and runs it against",
		)
		fmt.Println("    the current connection.")
		fmt.Println("  - If no selector is provided, pam will open the editor to build sql query")
		fmt.Println("  - The result is rendered as an interactive table in your terminal.")
		fmt.Println("  - With '--edit' or '-e', pam opens the query in your $EDITOR before")
		fmt.Println("    running it and saves any changes back to the configuration.")
		fmt.Println("  - With '--last' or '-l', runs the last used query")
		fmt.Println()
		section("Interactive table view")
		fmt.Println(
			styles.Faint.Render(
				"When results are shown, you can interact with the table using the keyboard:",
			),
		)
		fmt.Println()
		fmt.Println("  Arrow keys / h j k l  " + styles.Faint.Render("Move selection around the table"))
		fmt.Println("  PageUp / Ctrl+u       " + styles.Faint.Render("Scroll by a page up"))
		fmt.Println("  PageDown / Ctrl+d     " + styles.Faint.Render("Scroll by a page down"))
		fmt.Println("  Home / 0 / _          " + styles.Faint.Render("Jump to first row"))
		fmt.Println("  End / $               " + styles.Faint.Render("Jump to last row"))
		fmt.Println("  g / G                 " + styles.Faint.Render("Jump to top / bottom"))
		fmt.Println("  y / Enter             " + styles.Faint.Render("Copy current cell value to clipboard (if supported)"))
		fmt.Println("  v                     " + styles.Faint.Render("Start multi-selection mode"))
		fmt.Println("  u                     " + styles.Faint.Render("Update selected cell"))
		fmt.Println("  d                     " + styles.Faint.Render("Delete current row (requires WHERE clause)"))
		fmt.Println("  e                     " + styles.Faint.Render("Open the editor to update and rerun query"))
		fmt.Println("  s                     " + styles.Faint.Render("Save current query"))
		fmt.Println("  Esc /Ctrl+c           " + styles.Faint.Render("Quit the table view"))
		fmt.Println()
		fmt.Println(
			styles.Faint.Render(
				"Exact keys may vary depending on how the table component is wired,",
			),
		)
		fmt.Println(
			styles.Faint.Render(
				"but the basic navigation, search/filtering, and quit commands are available.",
			),
		)
		fmt.Println()
		section("Examples")
		fmt.Println("  pam run list_users")
		fmt.Println("  pam run \"select * from orders\"")
		fmt.Println("  pam run 2 --edit")
		fmt.Println("  pam run --last")
		fmt.Println("  pam query list_users")

	case "list":
		section("Command: list")
		fmt.Println(
			styles.Faint.Render(
				"List connections or queries. Defaults to queries.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  pam list [connections | queries] [search-term]")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  connections  " + styles.Faint.Render(
				"List all configured connections; active one is highlighted.",
			),
		)
		fmt.Println(
			"  queries      " + styles.Faint.Render(
				"List all saved queries for the current connection, with SQL.",
			),
		)
		fmt.Println(
			"               " + styles.Faint.Render(
				"Optionally filter by search term (searches name and SQL).",
			),
		)
		fmt.Println()
		section("Examples")
		fmt.Println(
			"  pam list                      # lists queries for the current connection",
		)
		fmt.Println("  pam list queries")
		fmt.Println("  pam list queries emp          # list queries containing 'emp'")
		fmt.Println("  pam list queries employees    # list queries containing 'employees'")
		fmt.Println("  pam list queries --oneline    # list each query in one separate line")
		fmt.Println("  pam list connections")

	case "tables":
		section("Command: tables")
		fmt.Println(
			styles.Faint.Render(
				"List all tables in the current database or query a specific table.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  pam tables [table-name] [--oneline | -o]")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  Without arguments, lists all tables in the current database connection.",
		)
		fmt.Println(
			"  With a table name, executes 'SELECT * FROM <table>' and displays results",
		)
		fmt.Println("  in an interactive table view.")
		fmt.Println()
		fmt.Println(
			"  --oneline, -o  " + styles.Faint.Render(
				"Display table names one per line (useful for scripting)",
			),
		)
		fmt.Println()
		section("Examples")
		fmt.Println("  pam tables              # list all tables")
		fmt.Println("  pam tables users        # query the users table")
		fmt.Println("  pam tables --oneline    # list tables in oneline format")

	case "disconnect":
		section("Command: disconnect")
		fmt.Println(
			styles.Faint.Render(
				"Disconnect from the current active database connection.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  pam disconnect")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  Clears the current active connection. You will need to use 'pam switch'",
		)
		fmt.Println("  to select a connection before running queries again.")
		fmt.Println()
		section("Examples")
		fmt.Println("  pam disconnect")

	case "edit":
		section("Command: edit")
		fmt.Println(
			styles.Faint.Render(
				"Open pam's configuration or queries in your editor.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  pam edit [config | queries]")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  config   " + styles.Faint.Render(
				"Edit the main configuration file (connections etc.).",
			),
		)
		fmt.Println(
			"  queries  " + styles.Faint.Render(
				"Edit all queries for the current connection in one file.",
			),
		)
		fmt.Println()
		section("Examples")
		fmt.Println("  pam edit           # defaults to 'config'")
		fmt.Println("  pam edit config")
		fmt.Println("  pam edit queries")

	case "info":
		section("Command: info")
		fmt.Println(
			styles.Faint.Render(
				"Show all tables or views in the current database connection.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  pam info <tables | views>")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  tables  " + styles.Faint.Render(
				"List all tables in the current connection/schema.",
			),
		)
		fmt.Println(
			"  views   " + styles.Faint.Render(
				"List all views in the current connection/schema.",
			),
		)
		fmt.Println()
		section("Columns displayed")
		fmt.Println("  - schema (if supported by database)")
		fmt.Println("  - name")
		fmt.Println("  - owner (if supported by database)")
		fmt.Println()
		section("Examples")
		fmt.Println("  pam info tables")
		fmt.Println("  pam info views")

	case "status":
		section("Command: status")
		fmt.Println(styles.Faint.Render("Show the current active connection."))
		fmt.Println()
		section("Usage")
		fmt.Println("  pam status")

	case "history":
		section("Command: history")
		fmt.Println(
			styles.Faint.Render("Show query execution history (placeholder)."),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  pam history")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  This command is currently a placeholder and will be implemented",
		)
		fmt.Println("  in a future release.")

	case "help":
		section("Command: help")
		fmt.Println(
			styles.Faint.Render(
				"Show general help or detailed help for a specific command.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  pam help [command]")
		fmt.Println()
		section("Examples")
		fmt.Println("  pam help")
		fmt.Println("  pam help run")
		fmt.Println("  pam help list")

	default:
		fmt.Printf("%s %q\n\n", styles.Error.Render("Unknown command"), cmd)
		a.PrintGeneralHelp()
	}
}
