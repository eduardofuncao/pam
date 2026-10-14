# Configuration

Squix stores its configuration at `~/.config/squix/config.yaml`.

## Saved queries

Saved queries live as plain SQL in `~/.config/squix/queries/`, one file per
**query group** at `queries/<group>.sql`. Each file is valid SQL you can open,
edit, or run in any client; squix slices it into named queries with directives:

```sql
-- @query list_orders
SELECT * FROM orders ORDER BY created_at DESC;

-- @query editable_users
-- @table users
-- @pk id
SELECT id, email FROM users;
```

A block starts at a `-- @query <name>` line. Optional `-- @table <name>` and
`-- @pk <col>[,<col>]` lines right after carry editable-table metadata (otherwise
inferred from the SQL at run time). Everything until the next `-- @query` (or
EOF) is the query body. A `-- @query` inside a body only starts a new block when
preceded by a blank line, so mid-SQL occurrences stay part of the body.

`squix edit` (no arguments) opens the whole group file in your editor; `squix
edit <name>` opens a single query.

Query ids (shown by `squix list`, usable via `squix run -s <n>`) are assigned by
position at load time and are not stored in the file. Prefer `squix run -s
<name>` — names are the stable selector.

## Shared query libraries

By default a connection owns a private group keyed by its own name, so a single
connection behaves exactly as before. To share one library across several
environments of the same service, name connections using the `{service}:{env}`
convention:

```yaml
current_connection: ecommerce:dev
connections:
  ecommerce:dev:  {db_type: postgres, conn_string: postgres://dev-host/db}
  ecommerce:prod: {db_type: postgres, conn_string: postgres://prod-host/db}
  analytics:      {db_type: sqlite,   conn_string: analytics.db}   # no colon → private group "analytics"
```

`ecommerce:dev`, `ecommerce:stg`, and `ecommerce:prod` are distinct connections
(each keeps its own `conn_string` and `schema`) but read and write
the **same** `queries/ecommerce.sql`. A query added on one is immediately visible
on the others. `analytics` (no colon) gets its own `queries/analytics.sql`.

> **Migration:** if you are upgrading from an older squix version that stored
> queries inline under each connection in `config.yaml`, they are lifted into
> `queries/<group>.sql` automatically on the first run (one-way). Back up
> `config.yaml` before the first run if you may need to downgrade. On a collision
> (two sibling envs each carrying different inline queries), the
> alphabetically-first connection wins and the others are dropped with a stderr
> warning so you can merge manually. Connection details stay in
> `config.yaml`.

## Row Limit `default_row_limit: 1000`
All queries are automatically limited to prevent fetching massive result sets. Configure via `default_row_limit` in config or use explicit `LIMIT` in your SQL queries.

Setting `default_row_limit: 0` **disables** automatic limiting — no row-limit clause is appended to your SQL. This matters on databases where the generated limit syntax is unsupported: e.g. Oracle before 12c rejects `FETCH FIRST ... ROWS ONLY` with `ORA-00933: SQL command not properly ended`.

Defaults:
- **Newly created** config files get `default_row_limit: 1000` written on first run.
- An **existing** file missing the key falls back to `0` (no limit).

## Column Width `default_column_width: 15`
The width for all columns in the table TUI is fixed to a constant size, which can be configured through `default_column_width` in the config file. There are plans to make the column widths flexible in future versions.

## Color Schemes `color_scheme: "default"`
Customize the terminal UI colors with built-in schemes:

**Available schemes:**
`default`, `dracula`, `gruvbox`, `solarized`, `nord`, `monokai`
`black-metal`, `black-metal-gorgoroth`, `vesper`, `catppuccin-mocha`, `tokyo-night`, `rose-pine`, `terracotta`

Each scheme uses a 7-color palette: Primary (titles, headers), Success (success messages), Error (errors), Normal (table data), Muted (borders, help text), Highlight (selected backgrounds), Accent (keywords, strings).

## UI Visibility `ui_visibility`

Control which UI components are displayed in the table view:

```yaml
ui_visibility:
  query_name: true          # Show query name header
  query_sql: true           # Show SQL query display
  type_display: true        # Show column type indicators
  key_icons: true           # Show primary key (⚿) and foreign key (⚭) icons
  footer_cell_content: true # Show current cell preview in footer
  footer_stats: true        # Show row/col count and position in footer
  footer_keymaps: true      # Show keybindings help in footer
```

**Tip:** Press `?` in the table view to toggle the keymaps help on/off.

## Keybindings `keybindings`

Customize keyboard shortcuts in the table view and the schema explorer. Each action accepts a single key or a list of keys:

```yaml
keybindings:
  move_up: "up"
  move_down: "down"
  move_left: "left"
  move_right: "right"
  search: "ctrl+f"
  help: "tab"
  export: ["ctrl+e", "ctrl+x"]
```

Overrides only the actions you specify — everything else keeps its default binding. See [Keybindings](keybindings.md) for the full action reference and details on modes.
