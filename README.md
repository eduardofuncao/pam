# Pam's Database Drawer
<img width="320" height="224" alt="image" src="https://github.com/user-attachments/assets/f995ce07-3742-4e98-b737-bbdbf982012e" />

## Demo
![demo](https://github.com/user-attachments/assets/c20ee5e9-ce01-41e4-ac12-e5206da49cdc)

## Commands

| Command | Description |
|---------|-------------|
| `init <name> <engine> <conn>` | Create connection |
| `switch <name>` | Switch connection |
| `status` | Show current connection |
| `add <name> [sql]` | Save a query |
| `remove <name>` | Delete a query |
| `run <name>` | Execute saved query |
| `list [queries\|connections]` | List items |
| `explore [table]` | Browse tables/data |
| `conf` | Edit config in $EDITOR |

## TUI Keys

| Key | Action |
|-----|--------|
| `hjkl` / Arrows | Move cursor |
| `g` / `G` | First / last row |
| `0` / `$` | First / last column |
| `v` | Visual selection |
| `y` | Yank |
| `e` | Edit cell |
| `;` | Command prompt |
| `q` | Quit |

Run `pam help connections` for connection string formats.

---

Thanks to these awesome projects for the inspiration:
- [naggie/dstask](https://github.com/naggie/dstask)
- [DeprecatedLuar/better-curl-saul](https://github.com/DeprecatedLuar/better-curl-saul)
