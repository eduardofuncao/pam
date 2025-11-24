package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/gohelp-luar"
)

func PrintHelp() {
	gohelp.PrintHeader("Usage")
	fmt.Println("  pam <command> [arguments]")
	fmt.Println()

	gohelp.PrintHeader("Database Connection")
	gohelp.Item("init <name> <engine> <conn-string>", "Initialize a new database connection")
	gohelp.Item("switch <db-name>", "Switch to a different connection")
	gohelp.Item("use <db-name>", "Alias for switch")
	gohelp.Item("status", "Show current active connection")
	fmt.Println()

	gohelp.PrintHeader("Query Management")
	gohelp.Item("add <query-name> [sql]", "Add a saved query (opens editor if no SQL)")
	gohelp.Item("save <query-name> [sql]", "Alias for add")
	gohelp.Item("remove <query-name>", "Remove a saved query")
	gohelp.Item("delete <query-name>", "Alias for remove")
	gohelp.Item("list queries", "List all saved queries for current connection")
	fmt.Println()

	gohelp.PrintHeader("Query Execution")
	gohelp.Item("query <query-name>", "Execute a saved query in view mode")
	gohelp.Item("query <query-name> --edit", "Execute with editor (save changes)")
	gohelp.Item("query <query-name> -e", "Short flag for --edit")
	gohelp.Item("run <query-name>", "Alias for query")
	fmt.Println()

	gohelp.PrintHeader("Configuration")
	gohelp.Item("list connections", "List all database connections")
	gohelp.Item("edit", "Edit config file in $EDITOR")
	gohelp.Item("edit config", "Edit config file directly")
	gohelp.Item("edit queries", "Edit queries as SQL file")
	fmt.Println()

	gohelp.PrintHeader("Database Engines")
	gohelp.Item("sqlite3", "SQLite database")
	gohelp.Item("postgres", "PostgreSQL database")
	gohelp.Item("mysql", "MySQL database")
	gohelp.Item("oracle", "Oracle database (via godror)")
	fmt.Println()

	gohelp.PrintHeader("Table Viewer Controls")
	gohelp.Item("Arrow keys / hjkl", "Navigate cells")
	gohelp.Item("Home/End / g/G", "Jump to start/end")
	gohelp.Item("c", "Copy cell value to clipboard")
	gohelp.Item("q / Ctrl+C", "Quit viewer")
	fmt.Println()

	gohelp.PrintHeader("Examples")
	fmt.Println("  pam init mydb sqlite3 ./data.db")
	fmt.Println("  pam add users \"SELECT * FROM users\"")
	fmt.Println("  pam run users")
	fmt.Println("  pam query users --edit")
	fmt.Println()

	gohelp.Separator()
}
