# Features

## Query Management

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
squix list

# Search for specific queries
squix list emp    # Finds queries with 'emp' in name or SQL
squix list employees --oneline # displays each query in one line

# Run by name or ID
squix run daily_report
squix run 2

# Edit query before running (great for testing parameter values)
squix run emp_by_salary --edit
```

> **Shared libraries:** Connections named `{service}:{env}` (e.g. `ecommerce:dev`, `ecommerce:prod`) share one query library — a query added on one is visible on the others. See [Shared query libraries](configuration.md#shared-query-libraries).

<img width="1188" height="714" alt="image" src="https://github.com/user-attachments/assets/016c7a61-ace4-49cc-9375-564ee6089899" />

## TUI Table Viewer

Navigate query results with Vim-style keybindings, update cells in-place, delete rows and copy data

<img width="1173" height="709" alt="image" src="https://github.com/user-attachments/assets/3959011b-532f-4374-a86d-a39217cd39f0" />

**Key Features:**
- Syntax-highlighted SQL display
- Column type indicators
- Primary key markers
- Live cell editing
- Visual selection mode
- Help overlay (`H`) with context-aware keybind reference
- Configurable keybindings via `config.yaml`

## Connection Switching

Manage multiple database connections and switch between them instantly.

```bash
# List all connections
squix list connections
squix use production
```
Display current connection and check if it is reachable
```
squix status
```
<div align=center>
  <img width="523" height="582" alt="image" src="https://github.com/user-attachments/assets/4046f6cd-376e-45c0-bcfd-20484e34470b" />
</div>

## Database Exploration

Explore your database schema and visualize relationships between tables.

```bash
# Open the interactive schema explorer (foldable tables/views sections)
# Keybinds: enter = SELECT *, c = columns, r = relationships, Tab = fold,
#           / = search items, H = help
squix explore

# Scriptable listing: print table/view names one per line
squix explore --oneline
squix explore --oneline --views   # views only

# Query a table directly
squix explore employees --limit 100

# View a local CSV file in the results view (read-only)
squix explore sales.csv

# Visualize foreign key relationships
squix explain employees
squix explain employees --depth 2    # Show relationships 2 levels deep
```

<img width="860" height="139" alt="image" src="https://github.com/user-attachments/assets/4cea0f4d-d3b9-4173-8b42-6ee6b289cc7b" />

**Note:** The `squix explain` command is currently a work in progress and may change in future versions.

---

## Editor Integration

Squix uses your `$EDITOR` environment variable for editing queries and UPDATE/DELETE statements, and for opening cell values from the detail view (press `o`).

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

# Edit a specific query
squix edit recent_users
```

## Interactive Shell

Run queries in an interactive REPL with persistent connection, history, and multi-line support.

```bash
squix shell          # or: squix repl
```

**Example session:**
```bash
squix@mydb> select * from users limit 5;
squix@mydb> list_users --status active
squix@mydb> --last
squix@mydb> list user
squix@mydb> status
```

**Meta-commands:**

| Command | Description |
|---------|-------------|
| `exit`, `quit`, `\q` | Exit the shell |
| `help`, `\h` | Show help |
| `status` | Show connection info |
| `list`, `ls`, `\l` | List queries or connections |

Multi-line: type SQL without trailing `;` to continue. End with `;` or press Enter on blank line to execute.

## Build variants

Squix ships in three variants so you can trade driver coverage for binary size and portability. DuckDB is the only driver needing CGO; the slimmer variants build without it.

| Variant | Excluded drivers | Included drivers |
|---------|------------------|------------------|
| full | none | postgres, mysql, sqlite, sqlserver, clickhouse, oracle, firebird, duckdb, snowflake |
| lite | duckdb, snowflake, oracle | postgres, mysql, sqlite, sqlserver, clickhouse, firebird |
| minimal | duckdb, snowflake, oracle, clickhouse, firebird, sqlserver | postgres, mysql, sqlite |

Check what your binary supports:

```bash
squix --drivers
```

Excluded drivers report which build tag removes them, e.g. `not included (excluded via -tags noduckdb)`. Rebuild without the tag, or install the full build, to enable them.

Build a variant from source:

```bash
make build          # full
make build-lite     # lite (-tags "noduckdb,nosnowflake,noracle")
make build-minimal  # minimal (-tags "noduckdb,nosnowflake,noracle,noclickhouse,nofirebird,nosqlserver")
```

Or via Nix flakes: `squix.packages.<system>.default` / `.lite` / `.minimal`. Release pages carry all three per platform (`squix-<os>-<arch>`, `squix-lite-<os>-<arch>`, `squix-minimal-<os>-<arch>`); the Homebrew formula and nixpkgs package track the full build.
