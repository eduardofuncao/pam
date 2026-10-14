---
name: squix
description: Run SQL queries across databases (Postgres, MySQL, SQLite, etc.) via the squix CLI. Use for running named queries, exploring schema, or parameterized queries. Always use -f for non-interactive output.
---

# Squix Skill for AI Coding Agents

## What is Squix

Squix is a CLI tool for managing and executing SQL queries across multiple
databases (PostgreSQL, MySQL, SQLite, Oracle, SQL Server, ClickHouse, Firebird).
It stores named queries as `.sql` files in `~/.config/squix/queries/`.

**Shared query libraries:** Connections named `{service}:{env}` (e.g. `ecommerce:dev`, `ecommerce:prod`) share one query library keyed by service — a query added on one is visible on siblings. `conn_string` stays per-connection. See [docs/configuration.md](docs/configuration.md#shared-query-libraries).

## Critical Rules

1. **Always use `-f` with `squix run` for SELECT queries** — without it, SELECT results open an interactive TUI; non-SELECT queries print a status message either way
2. **No inline connection flag** — must `squix switch <name>` before queries
3. **Check connection first** — run `squix status` to verify reachability

## Non-Interactive Output

```bash
squix run <query> -f json        # JSON array of objects
squix run <query> -f csv         # CSV with header row
squix run <query> -f tsv         # Tab-separated values
squix run <query> -f markdown    # Markdown table
squix run <query> -f html        # HTML table with styling
squix run <query> -f sql         # INSERT statements
squix run "SELECT 1" -f json     # Inline SQL works too
squix run --last -f csv          # Re-run last query
```

Output goes to stdout, errors to stderr. Exit code 1 on failure.
Pipe cleanly: `squix run list_users -f json > users.json`

Empty results: "No results found" on stderr, nothing on stdout.

## Setup Workflow

```bash
squix init mydb "postgresql://user:pass@localhost:5432/mydb"  # auto-infers DB type
squix switch mydb                                              # or: squix use mydb
squix add list_users "SELECT * FROM users"                     # SQL must be inline
squix run list_users -f json                                   # always use -f
```

## Explore Schema

```bash
squix explore                # Interactive schema explorer TUI (tables + views)
squix explore --oneline      # Non-interactive: print table/view names one per line
squix explore --oneline --tables   # tables only (--views for views only)
squix explain orders -d 2    # Show FK relationships as ASCII tree
```

Explorer TUI keys: `enter`=SELECT *, `c`=columns, `r`=relationships,
`Tab`=fold section, `/`=search. Defaults can be overridden by adding
`explorer_*` entries under `keybindings:` in config.yaml — see
docs/keybindings.md.

## List Queries

```bash
squix list queries            # Format: ◆ <id>/<name> (<table>)
squix list queries --oneline  # Compact: <id> <name> <table>
squix list connections        # List all configured connections
```

## Parameterized Queries

Named params use `:name` syntax. Optional defaults with `:name|default`:

```bash
squix add user_by_id "SELECT * FROM users WHERE id = :user_id"
squix run user_by_id --user_id 42 -f csv    # named
squix run user_by_id 42 -f csv              # positional
```

## Status & Cleanup

```bash
squix status                 # Shows: ● postgres/mydb (schema: public) — reachable
squix remove list_users      # Delete query (no confirmation needed)
echo y | squix remove -c mydb  # Delete connection (needs confirmation)
```

## Commands to Avoid (Interactive)

- `squix run <query>` without `-f` on a SELECT — opens TUI
- `squix shell` — interactive REPL
- `squix edit` — opens `$EDITOR`
- `squix explore` — interactive schema explorer TUI (use `--oneline`)
- `squix explore <table>` — opens TUI
- `squix add <name>` without SQL — opens `$EDITOR`
- `squix run` with no args — opens `$EDITOR`
