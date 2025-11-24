package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/gohelp-luar"
)

func PrintHelp(topic string) {
	switch topic {
	case "":
		printMainHelp()
	case "tui":
		printTUIHelp()
	case "connections":
		printConnectionsHelp()
	default:
		fmt.Printf("Unknown topic: %s\n\n", topic)
		printAvailablePages()
	}
}

func printMainHelp() {
	gohelp.PrintHeader("Usage")
	fmt.Println("  pam <command> [arguments]")

	gohelp.PrintHeader("Database")
	gohelp.Item("init <name> <engine> <conn>", "Create connection")
	gohelp.Item("switch <name>", "Switch connection")
	gohelp.Item("status", "Show current connection")

	gohelp.PrintHeader("Queries")
	gohelp.Item("add <name> [sql]", "Save a query")
	gohelp.Item("remove <name>", "Delete a query")
	gohelp.Item("run <name>", "Execute saved query")
	gohelp.Item("run <name> -e", "Execute with editor")
	gohelp.Item("run '<sql>'", "Execute raw SQL")

	gohelp.PrintHeader("Browse")
	gohelp.Item("list [queries|connections]", "List items")
	gohelp.Item("explore [table]", "Browse tables/data")
	gohelp.Item("conf", "Edit config in $EDITOR")

	printAvailablePages()
	gohelp.Separator()
}

func printAvailablePages() {
	gohelp.PrintHeader("Help Pages")
	gohelp.Item("help tui", "TUI navigation and command prompt")
	gohelp.Item("help connections", "Connection string formats")
	fmt.Println()
}

func printTUIHelp() {
	gohelp.PrintHeader("Navigation")
	gohelp.Item("hjkl / Arrows", "Move cursor")
	gohelp.Item("g / G", "First / last row")
	gohelp.Item("0 / $", "First / last column")
	gohelp.Item("Ctrl+U / Ctrl+D", "Page up / down")
	gohelp.Item("q", "Quit")

	gohelp.PrintHeader("Actions")
	gohelp.Item("v", "Toggle visual selection")
	gohelp.Item("y", "Yank cell or selection")
	gohelp.Item("e", "Edit cell (opens $EDITOR)")
	gohelp.Item("d", "Clear cell to NULL (with confirm)")

	gohelp.PrintHeader("Command Prompt")
	fmt.Println("  Press ; to open the command prompt")
	fmt.Println()
	gohelp.Item("Esc", "Cancel prompt")
	gohelp.Item("Enter", "Execute command")

	gohelp.PrintHeader("SQL Expansion")
	fmt.Println("  Commands auto-expand using current table:")
	fmt.Println()
	gohelp.Item("run SELECT *", "→ SELECT * FROM <table>")
	gohelp.Item("run SELECT * WHERE id=1", "→ SELECT * FROM <table> WHERE id=1")
	gohelp.Item("run DELETE WHERE id=1", "→ DELETE FROM <table> WHERE id=1")
	gohelp.Item("run UPDATE SET x=1 WHERE y=2", "→ UPDATE <table> SET x=1 WHERE y=2")
	fmt.Println()

	gohelp.Separator()
}

func printConnectionsHelp() {
	gohelp.PrintHeader("SQLite")
	fmt.Println("  pam init mydb sqlite3 ./data.db")
	fmt.Println("  pam init mydb sqlite3 /absolute/path/to/db.sqlite")

	gohelp.PrintHeader("PostgreSQL")
	fmt.Println("  pam init mydb postgres \"host=localhost port=5432 dbname=mydb sslmode=disable\"")
	fmt.Println("  pam init mydb postgres \"postgres://user:pass@localhost/dbname?sslmode=disable\"")

	gohelp.PrintHeader("MySQL")
	fmt.Println("  pam init mydb mysql \"user:pass@tcp(localhost:3306)/dbname\"")
	fmt.Println("  pam init mydb mysql \"user:pass@unix(/var/run/mysqld/mysqld.sock)/dbname\"")

	gohelp.PrintHeader("Oracle")
	fmt.Println("  pam init mydb oracle \"user/pass@localhost:1521/service\"")
	fmt.Println("  pam init mydb oracle 'user/pass@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=localhost)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=orcl)))'")
	fmt.Println()

	gohelp.Separator()
}
