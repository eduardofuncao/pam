<div align="center">
<h1>
  <img width="auto" height="45" alt="image" src="https://github.com/user-attachments/assets/f82ceec8-9fc7-4253-9ec0-f5548c646996" />
  Squix's SQL Stash
<img width="auto" height="36" alt="image" src="https://github.com/user-attachments/assets/c128f28f-dd10-4213-9915-dedafe7ae831" />

</h1>
<img width="360" height="131" alt="image" src="https://github.com/user-attachments/assets/9428a75b-ffa4-4961-919b-e5ccf192ef26" />

### **SQL Query Stashing for Terminal Squirrels**

> **Bear Grylls:** "Out here in the wild database ecosystem, efficiency means survival. See that squirrel? That’s Squix, or _Sequillis termius_. He doesn’t panic-write queries under pressure. He prepares. He caches. He optimizes. While others are wrestling with joins in the dark, Squix already has his results. Extraordinary creature."

---

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
![go badge](https://img.shields.io/badge/Go-1.21+-00ADD8?%20logo=go&logoColor=white)

**A minimal CLI tool for managing and executing SQL queries across multiple databases. Written in Go, made beautiful with BubbleTea**

[Quick Start](#--------quick-start) • [Configuration](#--------configuration) • [Database Support](#--------database-support) • [Dbeesly](#-dbeesly) • [Features](#--------features) • [Commands](#--------all-commands) • [TUI Navigation](#--------tui-table-navigation) • [Roadmap](#--------roadmap) • [Contributing](#contributing)

> This project is currently in beta, please report unexpected behavior through the issues tab

</div>


---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/464275ac-085e-451f-b783-c991d24d3635" />
    Demo
</h2>

![squixdemo2](https://github.com/user-attachments/assets/ee9653cf-6aaa-4be9-a898-37153ab0c898)

### Highlights

- **Query Library** - Save and organize your most-used queries
- **Runs in the CLI** - Execute queries with minimal overhead
- **Multi-Database** - Works with PostgreSQL, MySQL, SQLite, Oracle, SQL Server, ClickHouse and Firebird
- **Table view TUI** - Keyboard focused navigation with vim-style bindings
- **In-Place Editing** - Update cells, delete rows and edit your SQL directly from the results table
- **Export your data** - Export your data as CSV, JSON, SQL, Markdown or HTML tables

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/30765e98-13b3-4c18-81e7-faf224b60e0b" />
    Quick Start
</h2>

### Installation
Go to [the releases page](https://github.com/eduardofuncao/squix/releases) and find the correct version for your system. Download it and make sure the file is executable and moved to a directory in your $PATH.


<details>
<summary>Go install</summary>

Use go to install `squix` directly
```bash
go install github.com/eduardofuncao/squix/cmd/squix@latest
```
this will put the binary `squix` in your $GOBIN path (usually `~/go/bin`)
</details>

<details>
<summary>Build Manually</summary>

Follow these instructions to build the project locally
```bash
git clone https://github.com/eduardofuncao/squix

go build -o squix ./cmd/squix
```
The squix binary will be available in the root project directory
</details>

<details>
<summary>Nix / NixOS (Flake)</summary>

Squix is available as a Nix flake for easy installation on NixOS and systems with
Nix.


#### Run directly without installing
```bash
nix run github:eduardofuncao/squix
```

#### Install to user profile
```bash
nix profile install github:eduardofuncao/squix
```

#### Enter development shell
```bash
nix develop github:eduardofuncao/squix
```

#### NixOS System-wide

Add to your flake-based configuration.nix or flake.nix:

```nix
{
description = "My NixOS config";

inputs = {
  nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  squix.url = "github:eduardofuncao/squix";
};

outputs = { self, nixpkgs, squix, ... }: {
  nixosConfigurations.myHostname = nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    modules = [
      {
        nixpkgs.config.allowUnfree = true;
        environment.systemPackages = [
          squix.packages.x86_64-linux.default
        ];
      }
    ];
  };
};
}
```

Then rebuild: sudo nixos-rebuild switch

#### Home Manager

Add to your home.nix or flake config:

```nix
{
inputs = {
  nixpkgs.url = "github:NixOS/nixpkgs/nix-unstable";
  squix.url = "github:eduardofuncao/squix";
};

outputs = { self, nixpkgs, squix, ... }: {
  homeConfigurations."username" = {
    pkgs = nixpkgs.legacyPackages.x86_64-linux;
    modules = [
      {
        nixpkgs.config.allowUnfree = true;
        home.packages = [
          squix.packages.x86_64-linux.default
        ];
      }
    ];
  };
};
}
```

Then apply: home-manager switch

Note: Oracle support requires `allowUnfree = true` in your Nix configuration.
</details>

### Basic Usage

```bash
# Create your first connection (PostgreSQL example)
squix init mydb postgres "postgresql://user:pass@localhost:5432/mydb"

# Add a saved query
squix add list_users "SELECT * FROM users"

# List your saved queries
squix list queries

# Run it, this opens the interactive table viewer
squix run list_users

# Or run inline SQL
squix run "SELECT * FROM products WHERE price > 100"
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
x          # Export selected data as csv, tsv, json, sql, markdown or html

# Edit data directly
u          # Update current cell (opens your $EDITOR)
D          # Delete current row

# Modify and re-run
e          # Edit the query and re-run it

# Exit
q          # Quit back to terminal
```

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle" src="https://github.com/user-attachments/assets/8f5037c9-e616-4065-adfc-cd598621c887" />
    Configuration
</h2>

Squix stores its configuration at `~/.config/squix/config.yaml`.

### Row Limit `default_row_limit: 1000`
All queries are automatically limited to prevent fetching massive result sets. Configure via `default_row_limit` in config or use explicit `LIMIT` in your SQL queries.

### Column Width `default_column_width: 15`
The width for all columns in the table TUI is fixed to a constant size, which can be configured through `default_column_width` in the config file. There are plans to make the column widths flexible in version v0.2.0.

### Color Schemes `color_scheme: "default"`
Customize the terminal UI colors with built-in schemes:

**Available schemes:**
`default`, `dracula`, `gruvbox`, `solarized`, `nord`, `monokai`
`black-metal`, `black-metal-gorgoroth`, `vesper`, `catppuccin-mocha`, `tokyo-night`, `rose-pine`, `terracotta`

Each scheme uses a 7-color palette: Primary (titles, headers), Success (success messages), Error (errors), Normal (table data), Muted (borders, help text), Highlight (selected backgrounds), Accent (keywords, strings).

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/c46a2565-a58c-472c-9393-96724d9716da" />
    Database Support
</h2>

Examples of init/create commands to start working with different database types

### PostgreSQL

```bash
squix init pg-prod postgres postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable

# or connect to a specific schema:
squix init pg-prod postgres postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable schema-name
```

### MySQL / MariaDB

```bash
squix init mysql-dev mysql 'myuser:mypassword@tcp(127.0.0.1:3306)/mydb'

squix init mariadb-docker mariadb "root:MyStrongPass123@tcp(localhost:3306)/forestgrove"
```

### SQL Server


```bash
squix init sqlserver-docker sqlserver "sqlserver://sa:MyStrongPass123@localhost:1433/master"
```

### SQLite

```bash
squix init sqlite-local sqlite file:///home/eduardo/dbeesly/sqlite/mydb.sqlite
```

### Oracle

```bash
squix init oracle-stg oracle myuser/mypassword@localhost:1521/XEPDB1

# or connect to a specific schema:
squix init oracle-stg oracle myuser/mypassword@localhost:1521/XEPDB1 schema-name
```
> Make sure you have the [Oracle Instant Client](https://www.oracle.com/database/technologies/instant-client/downloads.html) or equivalent installed in your system

### ClickHouse

```bash
squix init clickhouse-docker clickhouse "clickhouse://myuser:mypassword@localhost:9000/forestgrove"
```

### FireBird

```bash
squix init firebird-docker firebird user:masterkey@localhost:3050//var/lib/firebird/data/the_office
```

---

## 🐝 Dbeesly

To run containerized test database servers for all supported databases, use the sister project [dbeesly](https://github.com/eduardofuncao/dbeesly)

<img width="879" height="571" alt="image" src="https://github.com/user-attachments/assets/c0a131eb-ea95-4523-86ac-cd00a561a5e0" />

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/c125a9f2-d4b6-4ec3-aef4-f52e1c8f48e8" />
    Features
</h2>


### Query Management

Save, organize, and execute your SQL queries with ease. 

```bash
# Add queries with auto-incrementing IDs
squix add daily_report "SELECT * FROM sales WHERE date = CURRENT_DATE"
squix add user_count "SELECT COUNT(*) FROM users"
squix add employees "SELECT TOP 10 * FROM employees ORDER BY last_name"

# Add parameterized queries with :param|default syntax
squix add emp_by_salary "SELECT * FROM employees WHERE salary > :min_sal|30000"
squix add search_users "SELECT * FROM users WHERE name LIKE :name|P% AND status = :status|active"

# When creating queries with params and not default, squix will prompt you for the param value every time you run the query
squix add search_by_name "SELECT * FROM employees where first_name = :name"

# Run parameterized queries with named parameters (order doesn't matter!)
squix run emp_by_salary --min_sal 50000
squix run search_users --name Michael --status active
# Or use positional args (must match SQL order)
squix run search_users Michael active

# List all saved queries
squix list queries

# Search for specific queries
squix list queries emp    # Finds queries with 'emp' in name or SQL
squix list queries employees --oneline # displays each query in one line

# Run by name or ID
squix run daily_report
squix run 2

# Edit query before running (great for testing parameter values)
squix run emp_by_salary --edit
```

<img width="1188" height="714" alt="image" src="https://github.com/user-attachments/assets/016c7a61-ace4-49cc-9375-564ee6089899" />

### TUI Table Viewer

Navigate query results with Vim-style keybindings, update cells in-place, delete rows and copy data

<img width="1173" height="709" alt="image" src="https://github.com/user-attachments/assets/3959011b-532f-4374-a86d-a39217cd39f0" />

**Key Features:**
- Syntax-highlighted SQL display
- Column type indicators
- Primary key markers
- Live cell editing
- Visual selection mode

### Connection Switching

Manage multiple database connections and switch between them instantly.

```bash
# List all connections
squix list connections
squix switch production
```
Display current connection and check if it is reachable
```
squix status
```
<div align=center>
  <img width="523" height="582" alt="image" src="https://github.com/user-attachments/assets/4046f6cd-376e-45c0-bcfd-20484e34470b" />
</div>

### Database Exploration

Explore your database schema and visualize relationships between tables.

```bash
# List all tables and views in multi-column format
squix explore

# Query a table directly
squix explore employees --limit 100

# Visualize foreign key relationships
squix explain employees
squix explain employees --depth 2    # Show relationships 2 levels deep
```

<img width="860" height="139" alt="image" src="https://github.com/user-attachments/assets/4cea0f4d-d3b9-4173-8b42-6ee6b289cc7b" />

**Note:** The `squix explain` command is currently a work in progress and may change in future versions.




---

### Editor Integration

Squix uses your `$EDITOR` environment variable for editing queries and UPDATE/DELETE statements.

<div align=center>
  <img width="448" height="238" alt="image" src="https://github.com/user-attachments/assets/f416f41a-8ec3-4a35-86e7-0bba6596f75f" />
</div>

```bash
# Set your preferred editor (example in bash)
export EDITOR=vim
export EDITOR=nano
export EDITOR=code
```

You can also use the editor to edit queries before running them

```bash
# Edit existing query before running
squix run daily_report --edit

# Create and run a new query on the fly
squix run

# Re-run the last executed query
squix run --last

# Edit all queries at once
squix edit queries
```

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/4b1425ae-7918-4a3f-b37c-41c3e443929e" />
    All Commands
</h2>

### Connection Management

| Command | Description | Example |
|---------|-------------|---------|
| `create <name> <type> <conn-string> [schema]` | Create new database connection | `squix create mydb postgres "postgresql://..."` |
| `switch <name>` | Switch to a different connection | `squix switch production` |
| `status` | Show current active connection | `squix status` |
| `list connections` | List all configured connections | `squix list connections` |

### Query Operations

| Command | Description | Example |
|---------|-------------|---------|
| `add <name> [sql]` | Add a new saved query | `squix add users "SELECT * FROM users"` |
| `remove <name\|id>` | Remove a saved query | `squix remove users` or `squix remove 3` |
| `list queries` | List all saved queries | `squix list queries` |
| `list queries --oneline` | lists each query in one line | `squix list -o` |
| `list queries <searchterm>` | lists queries containing search term | `squix list employees` |
| `run <name\|id\|sql>` | Execute a query | `squix run users` or `squix run 2` |
| `run` | Create and run a new query | `squix run` |
| `run --edit` | Edit query before running | `squix run users --edit` |
| `run --last`, `-l` | Re-run last executed query | `squix run --last` |
| `run --param` | run with named params | `squix run --name Squix` |


### Database Exploration

| Command | Description | Example |
|---------|-------------|---------|
| `explore` | List all tables and views in multi-column format | `squix explore` |
| `explore <table> [-l N]` | Query a table with optional row limit | `squix explore employees --limit 100` |
| `explain <table> [-d N] [-c]` | Visualize foreign key relationships | `squix explain employees --depth 2` |

### Info

| Command | Description | Example |
|---------|-------------|---------|
| `info tables` | List all tables from current schema | `squix info tables` |
| `info views` | List all views from current schema | `squix info views` |

### Configuration

| Command | Description | Example |
|---------|-------------|---------|
| `config` | Edit main configuration file | `squix config` |
| `edit` | Edit all queries for current connection | `squix edit` |
| `edit <name\|id>` | Edit a single named query | `squix edit 3` |
| `help [command]` | Show help information | `squix help run` |

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/504a8488-69bf-43b4-860b-0659a6db3c69" />
    TUI Table Navigation
</h2>

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
| `y` | Copy selected cell(s) to clipboard |
| `Enter` | Show cell value in detail view (with JSON formatting) |
| `u` | Update current cell (opens editor) |
| `D` | Delete current row (requires WHERE clause) |
| `e` | Edit and re-run query |
| `s` | Save current query |
| `q`, `Ctrl+c`, `Esc` | Quit table view |

### Detail View Mode

Press `Enter` on any cell to open a detailed view that shows the full cell content. If the content is valid JSON, it will be automatically formatted with proper indentation.

**In Detail View:**

| Key | Action |
|-----|--------|
| `↑`, `↓`, `j`, `k` | Scroll through content |
| `e` | Edit cell content (opens editor with formatted JSON) |
| `q`, `Esc`, `Enter` | Close detail view |

When you press `e` in detail view:
- The editor opens with the full content (JSON will be formatted)
- Edit the content as needed
- Save and close to update the database
- JSON validation is performed automatically
- The table view updates with the new value

### Visual Mode

Press `v` to enter visual mode, then navigate to select a range of cells. 
Press `y` to copy the selection as plain text, or `x` to export the selected data as csv, tsv, json, sql insert statement, markdown or html

> The copied or exported data will be available in your clipboard

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/432c6b41-b2e0-4326-a3cc-7b349a987bb0" />
    Roadmap
</h2>

> This project is currently in beta, please report unexpected behavior through the issues tab

### v0.1.0 Acorn 🌰
- [x] Multi-database support (PostgreSQL, MySQL, SQLite, Oracle, SQL Server, ClickHouse)
- [x] Query library with save/edit/remove functionality
- [x] Interactive TUI with Vim navigation
- [x] In-place cell updates and row deletion
- [x] Visual selection and copy (single and multi cell)
- [x] Syntax highlighting
- [x] Query editing in external editor
- [x] Primary key detection
- [x] Column type indicators
- [x] Row limit configuration option
- [x] Info command, list all tables/views in current connection

### v0.2.0 - Walnut 🧱
- [x] Program colors configuration option
- [x] Query parameter with prompt and defaults (e.g., `WHERE first_name = :name|Squix`)
- [x] CSV/JSON export for multiple cells
- [x] Display column types correctly for join queries
- [x] `squix explore` and `squix explain`

### v0.3.0 - Hazelnut 🪴
- [ ] Shell autocomplete (bash, fish, zsh)
- [ ] Encryption on connection username/password in config file
- [ ] Dynamic column width

---

## Contributing

We welcome contributions! Get started with detailed instructions from [CONTRIBUTING.md](CONTRIBUTING.md)

Thanks a lot to all the contributors:

<a href="https://github.com/DeprecatedLuar"><img src="https://github.com/DeprecatedLuar.png" width="40" /></a>
<a href="https://github.com/caiolandgraf"><img src="https://github.com/caiolandgraf.png" width="40" /></a>
<a href="https://github.com/g4brielklein"><img src="https://github.com/g4brielklein.png" width="40" /></a>
<a href="https://github.com/eduardofuncao"><img src="https://github.com/eduardofuncao.png" width="40" /></a>


## Acknowledgments

Squix wouldn't exist without the inspiration and groundwork laid by these fantastic projects:

- **[naggie/dstask](https://github.com/naggie/dstask)** - For the elegant CLI design patterns and file-based data storage approach
- **[DeprecatedLuar/better-curl-saul](https://github.com/DeprecatedLuar/better-curl-saul)** - For demonstrating a simple and genius approach to making a CLI tool
- **[dbeaver](https://github.com/dbeaver/dbeaver)** - The OG database management tool


Built with: 
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The TUI framework
- Go standard library and various database drivers

---

## License

MIT License - see [LICENSE](LICENSE) file for details

---

<div align="center">

**Made with 🐿️ by [@eduardofuncao](https://github.com/eduardofuncao)**

<img width="320" height="224" alt="Squix mascot" src="https://github.com/user-attachments/assets/f995ce07-3742-4e98-b737-bbdbf982012e" />

Previously Pam's Database Drawer, thanks to [u/marrsd](https://www.reddit.com/user/marrsd/) for suggesting the new name!

</div>
