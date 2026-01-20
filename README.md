<div align="center">

<h1>
    <img src="https://github.com/user-attachments/assets/ba9b84d3-860b-4225-bf34-34572d4833e0" alt="Pam logo" height="45" style="vertical-align: middle;"/> 
  Pam's Database Drawer
  <img width="auto" height="45" alt="bitmap" src="https://github.com/user-attachments/assets/c4dd1637-3e8d-45e8-8196-0d8b48324265" />
</h1>
<img width="363" height="120" alt="image" src="https://github.com/user-attachments/assets/4495a407-4897-4b22-8b5e-6ac8a9340ca5" />



### *"Pam, the receptionist, has been doing a fantastic job."*

> **Michael Scott:** "You know what's amazing? Pam. Pam is amazing. She's got this drawer - not just any drawer - a database drawer. Full of SQL queries. I didn't even know we needed that, but apparently everyone does because they keep asking her for them. 'Pam, I need the users query.' 'Pam, where's that sales report?' And she just opens the drawer and boom. There it is. I think it's the most popular drawer in the entire office. Maybe even in Scranton. Possibly Pennsylvania."

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


![pam-demo](https://github.com/user-attachments/assets/b62bec1d-2255-4d02-9b7f-1c99afbeb664)

### Highlights

- **Query Library** - Save and organize your most-used queries
- **Runs in the CLI** - Execute queries with minimal overhead
- **Multi-Database** - Works with PostgreSQL, MySQL, SQLite, Oracle, SQL Server and ClickHouse
- **Table view TUI** - Keyboard focused navigation with vim-style bindings
- **In-Place Editing** - Update cells, delete rows and edit your SQL directly from the results table

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/30765e98-13b3-4c18-81e7-faf224b60e0b" />
    Quick Start
</h2>

### Installation
Go to [the releases page](https://github.com/eduardofuncao/pam/releases) and find the correct version for your system. Download it and make sure the file is executable and moved to a directory in your $PATH.


<details>
<summary>Go install</summary>

Use go to install `pam` directly
```bash
go install github.com/eduardofuncao/pam/cmd/pam@latest
```
this will put the binary `pam` in your $GOBIN path (usually `~/go/bin`)
</details>

<details>
<summary>Build Manually</summary>

Follow these instructions to build the project locally
```bash
git clone https://github.com/eduardofuncao/pam

go build -o pam ./cmd/pam
```
The pam binary will be available in the root project directory
</details>

<details>
<summary>Nix / NixOS (Flake)</summary>

Pam is available as a Nix flake for easy installation on NixOS and systems with
Nix.


#### Run directly without installing
```bash
nix run github:eduardofuncao/pam
```

#### Install to user profile
```bash
nix profile install github:eduardofuncao/pam
```

#### Enter development shell
```bash
nix develop github:eduardofuncao/pam
```

#### NixOS System-wide

Add to your flake-based configuration.nix or flake.nix:

```nix
{
description = "My NixOS config";

inputs = {
  nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  pam.url = "github:eduardofuncao/pam";
};

outputs = { self, nixpkgs, pam, ... }: {
  nixosConfigurations.myHostname = nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    modules = [
      {
        nixpkgs.config.allowUnfree = true;
        environment.systemPackages = [
          pam.packages.x86_64-linux.default
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
  pam.url = "github:eduardofuncao/pam";
};

outputs = { self, nixpkgs, pam, ... }: {
  homeConfigurations."username" = {
    pkgs = nixpkgs.legacyPackages.x86_64-linux;
    modules = [
      {
        nixpkgs.config.allowUnfree = true;
        home.packages = [
          pam.packages.x86_64-linux.default
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
pam init mydb postgres "postgresql://user:pass@localhost:5432/mydb"

# Add a saved query
pam add list_users "SELECT * FROM users"

# List your saved queries
pam list queries

# Run it, this opens the interactive table viewer
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

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle" src="https://github.com/user-attachments/assets/8f5037c9-e616-4065-adfc-cd598621c887" />
    Configuration
</h2>

Pam stores its configuration at `~/.config/pam/config.yaml`.

### Row Limit `default_row_limit: 1000`
All queries are automatically limited to prevent fetching massive result sets. Configure via `default_row_limit` in config or use explicit `LIMIT` in your SQL queries.

### Column Width `default_column_width: 15`
The width for all columns in the table TUI is fixed to a constant size, which can be configured through `default_column_width` in the config file. There are plans to make the column widths flexible in version v0.2.0.

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/c46a2565-a58c-472c-9393-96724d9716da" />
    Database Support
</h2>

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
pam init sqlite-local sqlite file:///home/eduardo/dbeesly/sqlite/mydb.sqlite
```

### Oracle

```bash
pam init oracle-stg oracle myuser/mypassword@localhost:1521/XEPDB1

# or connect to a specific schema:
pam init oracle-stg oracle myuser/mypassword@localhost:1521/XEPDB1 schema-name
```
> Make sure you have the [Oracle Instant Client](https://www.oracle.com/database/technologies/instant-client/downloads.html) or equivalent installed in your system

### ClickHouse

```bash
pam init clickhouse-docker clickhouse "clickhouse://myuser:mypassword@localhost:9000/dundermifflin"
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
pam add daily_report "SELECT * FROM sales WHERE date = CURRENT_DATE"
pam add user_count "SELECT COUNT(*) FROM users"
pam add employees "SELECT TOP 10 * FROM employees ORDER BY last_name"

# List all saved queries
pam list queries

# Search for specific queries
pam list queries emp    # Finds queries with 'emp' in name or SQL
pam list queries employees --oneline # displays each query in one line

# Run by name or ID
pam run daily_report
pam run 2
```

<img width="1166" height="687" alt="image" src="https://github.com/user-attachments/assets/6f05c2dc-aa48-49ca-ab68-fdf3cfcc4eae" />

### TUI Table Viewer

Navigate query results with Vim-style keybindings, update cells in-place, dekete rows and copy data

<img width="1155" height="689" alt="image" src="https://github.com/user-attachments/assets/839bb77d-b358-43d0-98cd-0dc8102a9ac0" />

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
pam list connections

# Switch active connection
pam switch mysql_prod

# Display current connection and check if it is reacheable
pam status
```
<div align=center>
  <img width="425" height="503" alt="image" src="https://github.com/user-attachments/assets/e291de99-3c03-4e2a-b559-dcbbb89dc232" />
</div>

---

### Editor Integration

Pam uses your `$EDITOR` environment variable for editing queries and UPDATE/DELETE statements.

<div align=center>
  <img width="448" height="238" alt="image" src="https://github.com/user-attachments/assets/f416f41a-8ec3-4a35-86e7-0bba6596f75f" />
</div>

```bash
# Set your preferred editor
export EDITOR=vim
export EDITOR=nano
export EDITOR=code
```

You can also use the editor to edit queries before running them

```bash
# Edit existing query before running
pam run daily_report --edit

# Create and run a new query on the fly
pam run

# Re-run the last executed query
pam run --last

# Edit all queries at once
pam edit queries
```

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/4b1425ae-7918-4a3f-b37c-41c3e443929e" />
    All Commands
</h2>

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
| `list queries --oneline` | lists each query in one line | `pam list -o` |
| `list queries <searchterm>` | lists queries containing search term | `pam list employees` |
| `run <name\|id\|sql>` | Execute a query | `pam run users` or `pam run 2` |
| `run` | Create and run a new query | `pam run` |
| `run --edit` | Edit query before running | `pam run users --edit` |
| `run --last`, `-l` | Re-run last executed query | `pam run --last` |


### Info

| Command | Description | Example |
|---------|-------------|---------|
| `info tables` | List all tables from current schema | `pam info tables` |
| `info views` | List all views from current schema | `pam info views` |

### Configuration

| Command | Description | Example |
|---------|-------------|---------|
| `edit config` | Edit main configuration file | `pam edit config` |
| `edit queries` | Edit all queries for current connection | `pam edit queries` |
| `help [command]` | Show help information | `pam help run` |

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

Press `v` to enter visual mode, then navigate to select a range of cells. Press `y` to copy the selection.

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/432c6b41-b2e0-4326-a3cc-7b349a987bb0" />
    Roadmap
</h2>

> This project is currently in beta, please report unexpected behavior through the issues tab

### v0.1.0 Ryan
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

### v0.2.0 - Kelly
- [ ] Dynamic column width
- [ ] Program colors configuration option
- [ ] Query parameter/placeholder support (e.g., `WHERE id = $1`)
- [ ] Query execution history with persistence
- [ ] CSV/JSON export for multiple cells
- [ ] Display column types correctly for join queries

### v0.3.0 - Jim
- [ ] Shell autocomplete (bash, fish, zsh)
- [ ] `pam info table <table>` - Show table metadata (columns, types, constraints)
- [ ] `pam info connection` - Show connection/database overview


---

## Contributing

We welcome contributions! Get started with detailed instructions from [CONTRIBUTING.md](CONTRIBUTING.md)

Thanks a lot to all the contributors:

<a href="https://github.com/DeprecatedLuar"><img src="https://github.com/DeprecatedLuar.png" width="40" /></a>
<a href="https://github.com/caiolandgraf"><img src="https://github.com/caiolandgraf.png" width="40" /></a>
<a href="https://github.com/eduardofuncao"><img src="https://github.com/eduardofuncao.png" width="40" /></a>


## Acknowledgments

Pam wouldn't exist without the inspiration and groundwork laid by these fantastic projects:

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

**Made with 👚 by [@eduardofuncao](https://github.com/eduardofuncao)**

> *"I don't think it would be the worst thing if it didn't work out...  Wait, can I say that?"* - Pam Beesly (definitely NOT about Pam's Database Drawer)

<img width="320" height="224" alt="Pam mascot" src="https://github.com/user-attachments/assets/f995ce07-3742-4e98-b737-bbdbf982012e" />


</div>
