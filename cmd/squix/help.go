package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/eduardofuncao/squix/internal/styles"
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
			"Squix's SQL Stash - query manager for your databases",
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
	fmt.Println(styles.Separator.Render("  squix <command> [arguments]"))
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
			"Remove a saved query by name/id, or remove a connection entirely (alias: delete)",
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
		"  explore     " + styles.Faint.Render("Explore database schema"),
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
			"Edit queries in your editor",
		),
	)
fmt.Println(
		"  config      " + styles.Faint.Render(
			"Edit the main configuration file",
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
		"  explain     " + styles.Faint.Render(
			"Show relationships between tables",
		),
	)
	fmt.Println(
		"  help        " + styles.Faint.Render(
			"Show help for squix or a specific command",
		),
	)
	fmt.Println()

	// Short help
	fmt.Println(styles.Title.Render("Help"))
	fmt.Println(
		"  squix help              " + styles.Faint.Render("Show this help"),
	)
	fmt.Println(
		"  squix help <command>    " + styles.Faint.Render(
			"Show detailed help for a specific command",
		),
	)
	fmt.Println()

	// Examples
	fmt.Println(styles.Title.Render("Examples"))
	fmt.Println(
		"  squix init dev postgres \"postgres://user:pass@localhost:5432/dbname\"",
	)
	fmt.Println(
		"  squix init prod sqlserver \"sqlserver://sa:password@localhost:1433?database=mydb\"",
	)
	fmt.Println("  squix switch dev")
	fmt.Println("  squix add list_users \"SELECT * FROM users\"")
	fmt.Println("  squix run list_users")
	fmt.Println("  squix run \"select * from users\"")
	fmt.Println("  squix list connections")
	fmt.Println("  squix list queries")
	fmt.Println("  squix edit config")
	fmt.Println("  squix edit queries")
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
		fmt.Println("  squix init [flags]")
		fmt.Println("  squix init <name> <connection-string>          # type auto-inferred")
		fmt.Println("  squix init <name> <db-type> <connection-string> [schema]  # legacy")
		fmt.Println()
		section("Flags")
		fmt.Println("  --name, -n              Connection name")
		fmt.Println("  --type, -t              Database type (optional, auto-inferred from conn)")
		fmt.Println("  --conn-string, --conn, -c  Connection string")
		fmt.Println("  --schema, -s            Schema (optional)")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  - Opens and pings the database to verify the connection.",
		)
		fmt.Println("  - Saves the configuration if everything succeeds.")
		fmt.Println(
			"  - If any required parameter is missing, launches interactive TUI.",
		)
		fmt.Println(
			"  - Database type is auto-inferred from connection string when possible",
		)
		fmt.Println("    (works in flag mode and 2-arg positional mode).")
		fmt.Println("  - For Oracle databases, optionally specify a schema to set as default.")
		fmt.Println()
		section("Examples")
		fmt.Println("  # Flag mode with auto-inference")
		fmt.Println(
			"  squix init --name dev --conn \"postgres://user:pass@localhost:5432/dbname\"",
		)
		fmt.Println()
		fmt.Println("  # 2-arg positional with auto-inference")
		fmt.Println(
			"  squix init dev \"postgres://user:pass@localhost:5432/dbname\"",
		)
		fmt.Println()
		fmt.Println("  # 3-arg positional (legacy, explicit type)")
		fmt.Println(
			"  squix init dev postgres \"postgres://user:pass@localhost:5432/dbname\"",
		)
		fmt.Println()
		fmt.Println("  # Interactive mode")
		fmt.Println("  squix init")
		fmt.Println()
		fmt.Println("  # With schema")
		fmt.Println(
			"  squix init prod sqlserver \"sqlserver://sa:password@localhost:1433?database=mydb\" --schema public",
		)
		fmt.Println(
			"  squix init staging mysql \"user:pass@tcp(127.0.0.1:3306)/dbname\"",
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
		fmt.Println("  squix switch/use <connection-name>")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  - Sets the connection to be used by 'add', 'run', 'list queries', etc.",
		)
		fmt.Println()
		section("Examples")
		fmt.Println("  squix switch dev")
		fmt.Println("  squix use prod")

	case "add", "save":
		section("Command: add")
		fmt.Println(
			styles.Faint.Render(
				"Save a new named query under the current connection.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  squix add <run-name> [query]")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  - If [query] is omitted, squix opens $EDITOR (default: vim) so you",
		)
		fmt.Println("    can write the query interactively.")
		fmt.Println("  - Each query gets a numeric ID as well as a name.")
		fmt.Println("  - Requires an active connection (use 'squix switch').")
		fmt.Println()
		section("Examples")
		fmt.Println("  squix add list_users \"SELECT * FROM users\"")
		fmt.Println("  squix add update_status    # opens editor to write SQL")

	case "remove", "delete":
		section("Command: remove")
		fmt.Println(styles.Faint.Render("Remove a saved query by name/id, or remove a connection entirely."))
		fmt.Println()
		section("Usage")
		fmt.Println("  squix remove <run-name-or-id>              # Remove query")
		fmt.Println("  squix remove --connection <conn-name>    # Remove connection")
		fmt.Println("  squix remove -c <conn-name>             # Remove connection (short)")
		fmt.Println()
		section("Examples")
		fmt.Println("  squix remove list_users                    # Remove query")
		fmt.Println("  squix remove 3                             # Remove query by ID")
		fmt.Println("  squix remove --connection dev              # Remove connection")
		fmt.Println("  squix remove -c prod                         # Remove connection (short)")

	case "run", "query":
		section("Command: run")
		fmt.Println(
			styles.Faint.Render(
				"Execute a saved query against the current connection and show results in an interactive table view.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  squix run <query-name-or-id> [--edit | -e] [--last | -l]")
		fmt.Println("  squix run                      " + styles.Faint.Render("# Opens the editor to build sql query"))
		fmt.Println()
		section("Description")
		fmt.Println(
			"  - Looks up a saved query by name or numeric ID and runs it against",
		)
		fmt.Println("    the current connection.")
		fmt.Println("  - If no selector is provided, squix will open the editor to build sql query")
		fmt.Println("  - The result is rendered as an interactive table in your terminal.")
		fmt.Println("  - With '--edit' or '-e', squix opens the query in your $EDITOR before")
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
		fmt.Println("  squix run list_users")
		fmt.Println("  squix run \"select * from orders\"")
		fmt.Println("  squix run 2 --edit")
		fmt.Println("  squix run --last")
		fmt.Println("  squix query list_users")

	case "list":
		section("Command: list")
		fmt.Println(
			styles.Faint.Render(
				"List connections or queries. Defaults to queries.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  squix list [connections | queries] [search-term]")
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
			"  squix list                      # lists queries for the current connection",
		)
		fmt.Println("  squix list queries")
		fmt.Println("  squix list queries emp          # list queries containing 'emp'")
		fmt.Println("  squix list queries employees    # list queries containing 'employees'")
		fmt.Println("  squix list queries --oneline    # list each query in one separate line")
		fmt.Println("  squix list connections")

	case "tables":
		section("Command: tables")
		fmt.Println(
			styles.Faint.Render(
				"List all tables in the current database or query a specific table.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  squix tables [table-name] [--oneline | -o]")
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
		fmt.Println("  squix tables              # list all tables")
		fmt.Println("  squix tables users        # query the users table")
		fmt.Println("  squix tables --oneline    # list tables in oneline format")

	case "disconnect":
		section("Command: disconnect")
		fmt.Println(
			styles.Faint.Render(
				"Disconnect from the current active database connection.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  squix disconnect")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  Clears the current active connection. You will need to use 'squix switch'",
		)
		fmt.Println("  to select a connection before running queries again.")
		fmt.Println()
		section("Examples")
		fmt.Println("  squix disconnect")

	case "edit":
		section("Command: edit")
		fmt.Println(
			styles.Faint.Render(
				"Edit queries in your editor.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  squix edit [<query-name-or-id>]")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  - Opens the editor to modify queries for the current connection.",
		)
		fmt.Println("    - With no arguments: opens all queries in one file")
		fmt.Println("    - With query name/id: edits a single query")
		fmt.Println("    - Query name can be changed by editing the '-- queryname' header")
		fmt.Println("  - Requires an active connection (use 'squix switch').")
		fmt.Println()
		section("Examples")
		fmt.Println("  squix edit                    # edit all queries")
		fmt.Println("  squix edit list_users         # edit single query")
		fmt.Println("  squix edit 3                  # edit query by ID")
	case "config":
		section("Command: config")
		fmt.Println(
			styles.Faint.Render(
				"Edit the main configuration file.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  squix config")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  Opens the configuration file (~/.config/squix/config.yaml) in your editor.",
		)
		fmt.Println("  Allows you to edit connections, color schemes, and other settings.")
		fmt.Println()
		section("Examples")
		fmt.Println("  squix config")

	case "explore":
		section("Command: explore")
		fmt.Println(
			styles.Faint.Render(
				"Explore your database schema and query tables interactively.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  squix explore")
		fmt.Println("  squix explore <table> [--limit | -l N]")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  Without arguments, lists all tables and views in multi-column format.",
		)
		fmt.Println(
			"  With a table name, queries the table and shows results in an",
		)
		fmt.Println("  interactive table view (similar to 'squix run').")
		fmt.Println()
		fmt.Println(
			"  --limit, -l N  " + styles.Faint.Render(
				"Limit number of rows returned (default: from config or 1000)",
			),
		)
		fmt.Println()
		section("Examples")
		fmt.Println("  squix explore                  # list all tables and views")
		fmt.Println("  squix explore employees        # query employees table")
		fmt.Println("  squix explore orders -l 50     # query with 50 row limit")

	case "explain":
		section("Command: explain")
		fmt.Println(
			styles.Faint.Render(
				"Visualize foreign key relationships between tables.",
			),
		)
		fmt.Println()
		fmt.Println(
			styles.Faint.Render(
				"  Note: This command is currently a work in progress and may change.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  squix explain <table> [--depth | -d N]")
		fmt.Println()
		section("Description")
		fmt.Println(
			"  Shows a tree visualization of foreign key relationships for a table.",
		)
		fmt.Println("  Displays both 'belongs to' and 'has many' relationships.")
		fmt.Println()
		fmt.Println(
			"  --depth, -d N  " + styles.Faint.Render(
				"Show relationships up to N levels deep (default: 1)",
			),
		)
		fmt.Println()
		section("Relationship types")
		fmt.Println(
			"  belongs to [N:1]  " + styles.Faint.Render(
				"FK from this table to another table",
			),
		)
		fmt.Println(
			"  has many [1:N]   " + styles.Faint.Render(
				"FK from other tables to this table",
			),
		)
		fmt.Println()
		section("Examples")
		fmt.Println("  squix explain employees")
		fmt.Println("  squix explain employees --depth 2")
		fmt.Println("  squix explain departments -d 3")

	case "info":
		section("Command: info")
		fmt.Println(
			styles.Faint.Render(
				"Show all tables or views in the current database connection.",
			),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  squix info <tables | views>")
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
		fmt.Println("  squix info tables")
		fmt.Println("  squix info views")

	case "status":
		section("Command: status")
		fmt.Println(styles.Faint.Render("Show the current active connection."))
		fmt.Println()
		section("Usage")
		fmt.Println("  squix status")

	case "history":
		section("Command: history")
		fmt.Println(
			styles.Faint.Render("Show query execution history (placeholder)."),
		)
		fmt.Println()
		section("Usage")
		fmt.Println("  squix history")
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
		fmt.Println("  squix help [command]")
		fmt.Println()
		section("Examples")
		fmt.Println("  squix help")
		fmt.Println("  squix help run")
		fmt.Println("  squix help list")

	default:
		fmt.Printf("%s %q\n\n", styles.Error.Render("Unknown command"), cmd)
		a.PrintGeneralHelp()
	}
}
