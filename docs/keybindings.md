# TUI Table Navigation

When viewing query results in the TUI, you have full Vim-style navigation and editing capabilities.

## Basic Navigation

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

## Data Operations

| Key | Action |
|-----|--------|
| `v` | Enter visual selection mode |
| `y` | Copy selected cell(s) to clipboard |
| `x` | Export selected cell(s) as csv, tsv, json, sql insert statement, markdown or html to clipboard |
| `Enter` | Show cell value in detail view (with JSON formatting) |
| `u` | Update current cell |
| `D` | Delete current row |
| `e` | Edit and re-run query |
| `s` | Save current query |
| `H` | Show keyboard shortcuts overlay |
| `?` | Toggle keybindings help in footer |
| `q`, `Ctrl+c`, `Esc` | Quit table view |

## Search

| Key | Action |
|-----|--------|
| `/` | Search cell content |
| `n` | Jump to next cell match |
| `N` | Jump to previous cell match |
| `f` | Search column names |
| `;` | Jump to next column match |
| `,` | Jump to previous column match |

## Detail View Mode

Press `Enter` on any cell to open a detailed view that shows the full cell content. If the content is valid JSON, it will be automatically formatted with proper indentation.

**In Detail View:**

| Key | Action |
|-----|--------|
| `↑`, `↓`, `j`, `k` | Scroll through content |
| `u` | Update cell content (opens editor with formatted JSON) |
| `o` | Open value in `$EDITOR` (read-only) |
| `y` | Copy value to clipboard |
| `q`, `Esc`, `Enter` | Close detail view |

## Help Overlay

Press `H` to open a categorized keyboard shortcuts reference. The overlay is context-aware — it adjusts the "Actions" section based on the current view (tables list, query without primary key, etc.). Press `H`, `q`, or `Esc` to close.

## Schema Explorer

`squix explore` opens an interactive schema explorer with foldable Tables and Views sections (expanded by default).

| Key | Action |
|-----|--------|
| `j`, `↓`, `k`, `↑` | Move selection |
| `g` / `G` | Jump to top / bottom |
| `Ctrl+u`, `Ctrl+d` | Page up / down |
| `enter` | Run `SELECT *` on selected table/view |
| `c` | Show columns metadata in the results view |
| `r` | Show relationships tree (scrollable panel) |
| `Tab` | Fold/unfold current section |
| `/` | Search items |
| `n` / `N` | Next/previous match |
| `H` | Show keyboard shortcuts overlay |
| `?` | Toggle keybindings help in footer |
| `q`, `ctrl+c` | Quit |

**In the relationships panel:**

| Key | Action |
|-----|--------|
| `j`, `↓`, `k`, `↑` | Scroll (`move_down` / `move_up`) |
| `q`, `ctrl+c` | Close panel (`quit`) |

## Configurable Keybindings

All keybindings can be customized in `~/.config/squix/config.yaml` under the `keybindings` key. Each action accepts a single key or a list of keys.

```yaml
keybindings:
  quit: ["q", "ctrl+c"]
  move_up: k
  move_down: j
  search: "/"
  help: H
```

Keybindings are resolved per **mode** — different views can bind the same key to different actions:

| Mode | Context |
|------|---------|
| `normal` | Main table view |
| `detail` | Cell detail view (opened with `Enter`) |
| `help` | Help overlay |
| `explorer` | Schema explorer list view |

Unknown action names print a warning at startup. Conflicting keys within the same mode also print a warning.

### Action Reference

| Action | Default | Description |
|--------|---------|-------------|
| `move_up` | `k`, `up` | Move selection up |
| `move_down` | `j`, `down` | Move selection down |
| `move_left` | `h`, `left` | Move selection left |
| `move_right` | `l`, `right` | Move selection right |
| `jump_first_row` | `g` | Jump to first row |
| `jump_last_row` | `G` | Jump to last row |
| `jump_first_col` | `0`, `home`, `_` | Jump to first column |
| `jump_last_col` | `$`, `end` | Jump to last column |
| `page_up` | `ctrl+u`, `pgup` | Page up |
| `page_down` | `ctrl+d`, `pgdown` | Page down |
| `visual_mode` | `v` | Visual (characterwise) selection |
| `visual_line_mode` | `V` | Visual line selection |
| `yank` | `y` | Copy selection to clipboard |
| `export` | `x` | Export selected cells |
| `export_all` | `X` | Export all rows |
| `enter` | `enter` | Open detail view / select table |
| `update` | `u` | Update current cell |
| `delete_row` | `D` | Delete current row |
| `edit_sql` | `e` | Edit and re-run SQL |
| `save_query` | `s` | Save current query |
| `search` | `/` | Search cell content |
| `search_col` | `f` | Search column names |
| `next_match` | `n` | Next search match |
| `prev_match` | `N` | Previous search match |
| `prev_col_match` | `,` | Previous column match |
| `next_col_match` | `;` | Next column match |
| `help` | `H` | Show help overlay |
| `toggle_footer` | `?` | Toggle footer keybinds display |
| `quit` | `q`, `ctrl+c` | Quit table view |
| `detail_close` | `q`, `esc`, `enter` | Close detail view |
| `detail_edit` | `u` | Update cell from detail view |
| `detail_yank` | `y` | Copy from detail view |
| `detail_scroll_up` | `k`, `up` | Scroll detail view up |
| `detail_scroll_down` | `j`, `down` | Scroll detail view down |
| `detail_open` | `o` | Open value in editor (detail view) |
| `help_close` | `H`, `q`, `esc` | Close help overlay |
| `explorer_select` | `enter` | SELECT * from selected item (explorer) |
| `explorer_columns` | `c` | Show columns metadata (explorer) |
| `explorer_relations` | `r` | Show relationships tree (explorer) |
| `explorer_toggle_fold` | `tab` | Fold/unfold current section (explorer) |
