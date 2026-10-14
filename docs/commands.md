# Commands

## Connection Management

| Command | Description | Example |
|---------|-------------|---------|
| `init <name> <type> <conn-string> [schema]` | Create new database connection | `squix create mydb postgres "postgresql://..."` |
| `use/switch <name>` | Switch to a different connection | `squix use production` |
| `status` | Show current active connection | `squix status` |
| `list connections` | List all configured connections | `squix list connections` |

## Query Operations

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
| `shell` | Interactive query REPL (alias: `repl`) | `squix shell` |


## Database Exploration

| Command | Description | Example |
|---------|-------------|---------|
| `explore` | Open interactive schema explorer (tables + views) | `squix explore` |
| `explore --oneline [--tables] [--views]` | Print table/view names one per line | `squix explore --oneline` |
| `explore <table> [-l N]` | Query a table with optional row limit | `squix explore employees --limit 100` |
| `explore <file.csv>` | View a local CSV file in the results view (read-only) | `squix explore sales.csv` |
| `explain <table> [-d N] [-v]` | Visualize foreign key relationships | `squix explain employees --depth 2` |
| `example [path] [--force]` | Create a sample office-themed SQLite database to try squix | `squix example` |

## Configuration

| Command | Description | Example |
|---------|-------------|---------|
| `config` | Edit main configuration file | `squix config` |
| `edit` | Edit all queries for current connection's library (shared across `{service}:{env}` siblings) | `squix edit` |
| `edit <name\|id>` | Edit a single named query | `squix edit 3` |
| `remove --connection <name>` | Remove a db connection | `squix remove --conection dev4`` |
| `help [command]` | Show help information | `squix help run` |

