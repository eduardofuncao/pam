<div align="center">

# 🗃️ Pam's Database Drawer

<img width="320" height="224" alt="Pam mascot" src="https://github.com/user-attachments/assets/f995ce07-3742-4e98-b737-bbdbf982012e" />

### *"Pam, the secretary, has been doing a fantastic job."*

> **Michael Scott:** "You know what's amazing? Pam. Pam is amazing. She's got this drawer - not just any drawer - a database drawer. Full of SQL queries. I didn't even know we needed that, but apparently everyone does because they keep asking her for them. 'Pam, I need the users query.' 'Pam, where's that sales report?' And she just opens the drawer and boom. There it is. I think it's the most popular drawer in the entire office. Maybe even in Scranton. Possibly Pennsylvania."

---

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
![go badge](https://img.shields.io/badge/Go-1.21+-00ADD8?%20logo=go&logoColor=white)

**A minimal, pretty fast CLI tool for managing and executing SQL queries across multiple databases**
**Written in Go, powered by CharmBracelet's Bubblegum**


[Quick Start](#-quick-start) • [Database Support](#%EF%B8%8F-database-support) • [Features](#-features) • [Commands](#-all-commands) • [TUI Navigation](#%EF%B8%8F-tui-table-navigation) • [Roadmap](#%EF%B8%8F-roadmap)

</div>

---

## 🎬 Demo

![demo](https://github.com/user-attachments/assets/c20ee5e9-ce01-41e4-ac12-e5206da49cdc)

### Highlights

- **Lightning Fast** - Execute queries instantly with minimal overhead
- **Beautiful TUI** - Keyboard focused navigation with vim-style bindings
- **Query Library** - Save and organize your most-used queries
- **Multi-Database** - Works with PostgreSQL, MySQL, SQLite, and Oracle
- **In-Place Editing** - Update cells, delete rows and edit your SQL directly from the results table
- **Smart Copy** - Yank cells or ranges with visual mode
- **Inline Execution** - Run ad-hoc queries without saving

---

## 🚀 Quick Start

### Installation

```bash
go install github.com/eduardofuncao/pam/cmd/pam@latest
```

### Basic Usage

```bash
# Create your first connection
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
# Use vim-style navigation
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

## 🗄️ Database Support

### PostgreSQL

```bash
# Standard connection
pam init pg_prod postgres "postgresql://user:password@localhost:5432/dbname"

# With SSL
pam init pg_secure postgres "postgresql://user:pass@host:5432/db? sslmode=require"
```

### MySQL / MariaDB

```bash
# Basic connection
pam init mysql_dev mysql "user:password@tcp(localhost:3306)/dbname"

# With parameters
pam init mysql_prod mysql "user:pass@tcp(host:3306)/db?charset=utf8mb4&parseTime=True"
```

### SQLite

```bash
# Local file
pam init local_db sqlite "/path/to/database.db"

# Relative path
pam init app_db sqlite "./data/app.db"
```

### Oracle

```bash
# Standard connection
pam init oracle_prod oracle "user/password@localhost:1521/ORCL"

# TNS connection
pam init oracle_tns oracle "user/pass@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=host)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=svc)))"
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

# List all saved queries
pam list queries

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

## 📖 All Commands

### Connection Management

| Command | Description | Example |
|---------|-------------|---------|
| `create <name> <type> <conn-string>` | Create new database connection | `pam create mydb postgres "postgresql://..."` |
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
- [x] Multi-database support (PostgreSQL, MySQL, SQLite, Oracle)
- [x] Query library with save/remove functionality
- [x] Interactive TUI with Vim navigation
- [x] In-place cell updates
- [x] Visual selection and copy
- [x] Syntax highlighting
- [x] Query editing in external editor
- [x] Primary key detection
- [x] Column type indicators
- [ ] SQL Server driver support

### v0.2 - Jim
- [ ] Row limit configuration option
- [ ] Query parameter/placeholder support (e.g., `WHERE id = $1`)
- [ ] Query execution history with persistence
- [ ] CSV/JSON export for multiple cells
- [ ] Database schema handling

### v0.3 - Dwight
- [ ] Shell autocomplete (bash, fish, zsh)
- [ ] Search/filter within table view
- [ ] `pam info table <table>` - Show table metadata (columns, types, constraints)
- [ ] `pam info tables` - List all tables in current connection
- [ ] `pam info connection` - Show connection/database overview

---

## 🙏 Acknowledgments

Pam wouldn't exist without the inspiration and groundwork laid by these fantastic projects:

- **[naggie/dstask](https://github.com/naggie/dstask)** - For the elegant CLI/TUI design patterns and file-based configuration approach
- **[DeprecatedLuar/better-curl-saul](https://github.com/DeprecatedLuar/better-curl-saul)** - For demonstrating an easier, reapetable way to make http requests in the terminal

Built with: 
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The TUI framework
- Go standard library and various database drivers

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

<div align="center">

**Made with ❤️ by [@eduardofuncao](https://github.com/eduardofuncao)**

> *"I don't think it would be the worst thing if they didn't work out...  Wait, can I say that?"* - Pam Beesly (definitely NOT about Pam's Database Drawer)

</div>
