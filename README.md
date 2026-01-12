<div align="center">

<h1>
  <img src="https://github.com/user-attachments/assets/ba9b84d3-860b-4225-bf34-34572d4833e0" alt="Pam logo" height="45" style="vertical-align: middle;"/> 
  Pam's Database Drawer
</h1>
<img width="320" height="224" alt="Pam mascot" src="https://github.com/user-attachments/assets/f995ce07-3742-4e98-b737-bbdbf982012e" />



### *"Pam, the receptionist, has been doing a fantastic job."*

> **Michael Scott:** "You know what's amazing? Pam. Pam is amazing. She's got this drawer - not just any drawer - a database drawer. Full of SQL queries. I didn't even know we needed that, but apparently everyone does because they keep asking her for them. 'Pam, I need the users query.' 'Pam, where's that sales report?' And she just opens the drawer and boom. There it is. I think it's the most popular drawer in the entire office. Maybe even in Scranton. Possibly Pennsylvania."

---

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
![go badge](https://img.shields.io/badge/Go-1.21+-00ADD8?%20logo=go&logoColor=white)

**A minimal, pretty fast CLI tool for managing and executing SQL queries across multiple databases. Written in Go, made beautify with CharmBracelet's BubbleTea**


[Quick Start](#-quick-start) • [Configuration](#%EF%B8%8F-configuration) • [Database Support](#%EF%B8%8F-database-support) • [SQL Dialects](#%EF%B8%8F-sql-dialect-differences) • [Features](#-features) • [Commands](#-all-commands) • [TUI Navigation](#%EF%B8%8F-tui-table-navigation) • [Troubleshooting](#%EF%B8%8F-troubleshooting) • [Roadmap](#%EF%B8%8F-roadmap) • [Contributing](#-contributing) • [Testing](#-testing)

</div>

---

## 🎬 Demo

![demo](https://github.com/user-attachments/assets/c20ee5e9-ce01-41e4-ac12-e5206da49cdc)

### Highlights

- **Pretty Fast** - Execute queries with minimal overhead
- **Table view TUI** - Keyboard focused navigation with vim-style bindings
- **Query Library** - Save and organize your most-used queries
- **Multi-Database** - Works with PostgreSQL, MySQL, SQLite, Oracle, SQL Server, DuckDB, and ClickHouse
- **In-Place Editing** - Update cells, delete rows and edit your SQL directly from the results table
- **Smart Copy** - Yank cells or ranges with visual mode

---

## 🚀 Quick Start

### Installation
Find the pre-built binaries for your computer's architecture or install it with go:

```bash
go install github.com/eduardofuncao/pam/cmd/pam@latest
```

### Basic Usage

```bash
# Create your first connection (PostgreSQL example)
pam init mydb postgres "postgresql://user:pass@localhost:5432/mydb"

# Add a saved query
pam add list_users "SELECT * FROM users"

# Run it - this opens the interactive table viewer
pam run list_users

# Or run inline SQL
pam run "SELECT * FROM products WHERE price > 100"
```

### Navigating the Table

Once your query results appear, you can navigate and interact with the data: 

```bash
# Use vim-style navigation or arrow-keys
j/k        # Move down/up
h/l        # Move left/right
g/G        # Jump to first/last row

# Copy data
y          # Yank (copy) current cell
v          # Enter visual mode to select multiple cells and copy with y

# Edit data directly
u          # Update current cell (opens your $EDITOR)
D          # Delete current row

# Modify and re-run
e          # Edit the query and re-run it

# Exit
q          # Quit back to terminal
```

---

## ⚙️ Configuration

Pam stores its configuration at `~/.config/pam/config.yaml`.

### Row Limit
All queries are automatically limited to prevent fetching massive result sets. Configure via `default_row_limit` in config or use explicit `LIMIT` in your SQL queries.

---

## 🗄️ Database Support
Examples of init/create commands to start working with different database types

### PostgreSQL

```bash
pam init pg-prod postgres postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable

# or connect to a specific schema:
pam init pg-prod postgres postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable schema-name
```

### MySQL / MariaDB

```bash
pam init mysql-dev mysql 'myuser:mypassword@tcp(127.0.0.1:3306)/mydb'

pam init mariadb-docker mariadb "root:MyStrongPass123@tcp(localhost:3306)/dundermifflin"
```

### SQL Server


```bash
pam init sqlserver-docker sqlserver "sqlserver://sa:MyStrongPass123@localhost:1433/master"
```

### SQLite

```bash
pam init sqlite-local sqlite file:/home/eduardo/code/dbeesly/sqlite/mydb.sqlite
```

### DuckDB

```bash
pam init duckdb-local duckdb /home/user/code/dbeesly/duckdb/data/dundermifflin.duckdb
```


### Oracle

```bash
pam init oracle-stg oracle myuser/mypassword@localhost:1521/XEPDB1

# or connect to a specific schema:
pam init oracle-stg oracle myuser/mypassword@localhost:1521/XEPDB1 schema-name
```

### ClickHouse

```bash
pam init clickhouse-docker clickhouse "clickhouse://myuser:mypassword@localhost:9000/dundermifflin"
```


---

## ✨ Features

### Query Management

Save, organize, and execute your SQL queries with ease. 

ADD GIF HERE

```bash
# Add queries with auto-incrementing IDs
pam add daily_report "SELECT * FROM sales WHERE date = CURRENT_DATE"
pam add user_count "SELECT COUNT(*) FROM users"
pam add employees "SELECT TOP 10 * FROM employees ORDER BY last_name"

# List all saved queries
pam list queries

# Search for specific queries
pam list queries emp    # Finds queries with 'emp' in name or SQL
pam list queries employees

# Run by name or ID
pam run daily_report
pam run 2
```

### Interactive Editing

Open queries in your favorite editor before execution.

ADD GIF HERE

```bash
# Edit existing query before running
pam run daily_report --edit

# Create and run a new query on the fly
pam run --new

# Edit all queries at once
pam edit queries
```

### TUI Table Viewer

Navigate query results with Vim-style keybindings, update cells in-place, and copy data effortlessly.

ADD GIF HERE

**Key Features:**
- Syntax-highlighted SQL display
- Column type indicators
- Primary key markers
- Live cell editing
- Visual selection mode

### 🔄 Connection Switching

Manage multiple database connections and switch between them instantly.

```bash
# List all connections
pam list connections

# Switch active connection
pam switch mysql_prod

# Check current connection
pam status
```

### In-Place Updates & Deletes

Modify data directly from the results table - no need to write UPDATE statements manually! 

ADD GIF HERE

---

### 📝 Editor Integration

Pam uses your `$EDITOR` environment variable for editing queries and UPDATE/DELETE statements.

```bash
# Set your preferred editor
export EDITOR=vim
export EDITOR=nano
export EDITOR=code
```

---

## 📖 All Commands

### Connection Management

| Command | Description | Example |
|---------|-------------|---------|
| `create <name> <type> <conn-string> [schema]` | Create new database connection | `pam create mydb postgres "postgresql://..."` |
| `switch <name>` | Switch to a different connection | `pam switch production` |
| `status` | Show current active connection | `pam status` |
| `list connections` | List all configured connections | `pam list connections` |

### Query Operations

| Command | Description | Example |
|---------|-------------|---------|
| `add <name> [sql]` | Add a new saved query | `pam add users "SELECT * FROM users"` |
| `remove <name\|id>` | Remove a saved query | `pam remove users` or `pam remove 3` |
| `list queries` | List all saved queries | `pam list queries` |
| `run <name\|id\|sql>` | Execute a query | `pam run users` or `pam run 2` |
| `run --edit` | Edit query before running | `pam run users --edit` |
| `run --new` | Create and run new query | `pam run --new` |
| `run` | Re-run last query | `pam run` |

### Configuration

| Command | Description | Example |
|---------|-------------|---------|
| `edit config` | Edit main configuration file | `pam edit config` |
| `edit queries` | Edit all queries for current connection | `pam edit queries` |
| `help [command]` | Show help information | `pam help run` |

---

## ⌨️ TUI Table Navigation

When viewing query results in the TUI, you have full Vim-style navigation and editing capabilities. 

### Basic Navigation

| Key | Action |
|-----|--------|
| `h`, `←` | Move left |
| `j`, `↓` | Move down |
| `k`, `↑` | Move up |
| `l`, `→` | Move right |
| `g` | Jump to first row |
| `G` | Jump to last row |
| `0`, `_`, `Home` | Jump to first column |
| `$`, `End` | Jump to last column |
| `Ctrl+u`, `PgUp` | Page up |
| `Ctrl+d`, `PgDown` | Page down |

### Data Operations

| Key | Action |
|-----|--------|
| `v` | Enter visual selection mode |
| `y`, `Enter` | Copy selected cell(s) to clipboard |
| `u` | Update current cell (opens editor) |
| `D` | Delete current row (requires WHERE clause) |
| `e` | Edit and re-run query |
| `q`, `Ctrl+c`, `Esc` | Quit table view |

### Visual Mode

Press `v` to enter visual mode, then navigate to select a range of cells.  Press `y` or `Enter` to copy the selection.

ADD GIF HERE

---
## 🗺️ Roadmap

### v0.1 Kelly
- [x] Multi-database support (PostgreSQL, MySQL, SQLite, Oracle, SQL Server, DuckDB, ClickHouse)
- [x] Query library with save/edit/remove functionality
- [x] Interactive TUI with Vim navigation
- [x] In-place cell updates and row deletion
- [x] Visual selection and copy (single and multi cell)
- [x] Syntax highlighting
- [x] Query editing in external editor
- [x] Primary key detection
- [x] Column type indicators

### v0.2 - Jim
- [ ] Row limit configuration option
- [ ] Program colors configuration option
- [ ] Query parameter/placeholder support (e.g., `WHERE id = $1`)
- [ ] Query execution history with persistence
- [ ] CSV/JSON export for multiple cells
- [ ] Database schema handling
- [ ] Display column types correctly for join queries

### v0.3 - Dwight
- [ ] Shell autocomplete (bash, fish, zsh)
- [ ] `pam info table <table>` - Show table metadata (columns, types, constraints)
- [ ] `pam info tables` - List all tables in current connection
- [ ] `pam info connection` - Show connection/database overview

---

## 🤝 Contributing

We welcome contributions! Get started with detailed instructions from [CONTRIBUTING.md](CONTRIBUTING.md)

## 🙏 Acknowledgments

Pam wouldn't exist without the inspiration and groundwork laid by these fantastic projects:

- **[naggie/dstask](https://github.com/naggie/dstask)** - For the elegant CLI/TUI design patterns and file-based configuration approach
- **[DeprecatedLuar/better-curl-saul](https://github.com/DeprecatedLuar/better-curl-saul)** - For demonstrating a simple and genius approach to making a CLI program

Built with: 
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The TUI framework
- Go standard library and various database drivers

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

<div align="center">

**Made with ❤️ by [@eduardofuncao](https://github.com/eduardofuncao)**

> *"I don't think it would be the worst thing if it didn't work out...  Wait, can I say that?"* - Pam Beesly (definitely NOT about Pam's Database Drawer)

</div>
