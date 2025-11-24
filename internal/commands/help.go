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
	fmt.Println()

	gohelp.PrintHeader("Query Execution")
	gohelp.Item("run <query-name>", "Execute a saved query")
	gohelp.Item("run <query-name> -e", "Execute saved query with editor")
	gohelp.Item("run -e", "Open empty editor for one-shot query")
	gohelp.Item("run '<raw-sql>'", "Execute raw SQL directly (quotes required)")
	gohelp.Item("query <query-name>", "Alias for run")
	fmt.Println()

	gohelp.PrintHeader("Listing & Inspection")
	gohelp.Item("list", "List saved queries (default)")
	gohelp.Item("list queries", "List all saved queries for current connection")
	gohelp.Item("list connections", "List all database connections")
	gohelp.Item("list tables", "List all tables in current database")
	gohelp.Item("ls", "Alias for list")
	fmt.Println()

	gohelp.PrintHeader("Configuration")
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
	fmt.Println("  pam add users 'SELECT * FROM users'")
	fmt.Println("  pam run users")
	fmt.Println("  pam run users -e")
	fmt.Println("  pam run 'SELECT * FROM users WHERE id = 5'")
	fmt.Println("  pam run -e")
	fmt.Println()

	gohelp.Separator()
}
