<div align="center">
<h1>
  <img width="auto" height="45" alt="image" src="https://github.com/user-attachments/assets/f82ceec8-9fc7-4253-9ec0-f5548c646996" />
  Squix's SQL Stash
<img width="auto" height="36" alt="image" src="https://github.com/user-attachments/assets/c128f28f-dd10-4213-9915-dedafe7ae831" />

</h1>
<img width="360" height="131" alt="image" src="https://github.com/user-attachments/assets/9428a75b-ffa4-4961-919b-e5ccf192ef26" />

### **SQL Query Stashing for Terminal Squirrels**

---

[![MIT License](https://img.shields.io/badge/license-MIT-white.svg)](LICENSE)
![go badge](https://img.shields.io/badge/g=Go-1.25+-00ADD8?%20logo=go&logoColor=white&label=go)
[![Matrix](https://img.shields.io/matrix/squix:matrix.org?server_fqdn=matrix.org&label=chat&color=green)](https://matrix.to/#/#squix-sql:matrix.org)
[![Sponsor](https://img.shields.io/badge/sponsor-%E2%9D%A4-pink.svg)](https://github.com/sponsors/eduardofuncao)

**A CLI tool for managing and executing SQL queries across multiple databases. Written in Go, made beautiful with BubbleTea**

[Quick Start](#--------quick-start) • [Configuration](docs/configuration.md) • [Commands](docs/commands.md) • [Keybindings](docs/keybindings.md) • [Features](docs/features.md) • [Completion](docs/completion.md) • [Databases](docs/databases.md) • [Roadmap](#--------roadmap) • [Contributing](CONTRIBUTING.md)

> This project is currently in beta, please report unexpected behavior through the issues tab

</div>


---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/464275ac-085e-451f-b783-c991d24d3635" />
    Demo
</h2>

![squixdemo2](https://github.com/user-attachments/assets/ee9653cf-6aaa-4be9-a898-37153ab0c898)

> Try out the [live demo](https://squix.live.eduardofuncao.com) (desktop only, no install required!)

### Highlights

- **Query Library** - Save and organize your most-used queries, with parameterized support
- **Multi-Database** - Works with PostgreSQL, MySQL, SQLite, Oracle, SQL Server, ClickHouse, Firebird and DuckDB
- **Table view TUI** - Keyboard focused navigation and search with configurable vim-style bindings
- **In-Place Editing** - Update cells, delete rows and edit your SQL directly from the results table
- **Export your data** - Export your data as CSV, JSON, SQL, Markdown or HTML tables
- **Connection Switching** - Manage multiple databases and switch whenever you want
- **Database Exploration** - Interactive schema explorer with search and foldable sections, plus foreign-key relationship trees
- **Interactive Shell** - REPL with history, multi-line, and meta-commands

See [Features](docs/features.md) for details and examples

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/30765e98-13b3-4c18-81e7-faf224b60e0b" />
    Quick Start
</h2>

### Installation
Go to [the releases page](https://github.com/eduardofuncao/squix/releases) and find the correct version for your system. Download it and make sure the file is executable and moved to a directory in your $PATH. Alternatively, use one of the methods below:


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

See [Build variants](docs/features.md#build-variants) for the full/lite/minimal builds and which drivers each one includes.
</details>

<details>
<summary>Nixpkgs</summary>

Squix is available in [nixpkgs](https://search.nixos.org/packages?query=squix) (v0.4.0-beta+).

```bash
# Try it
nix-shell -p squix

# Or with flakes
nix run nixpkgs#squix
```
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
</details>

<details>
<summary>Homebrew</summary>

```bash
brew tap eduardofuncao/squix
brew install squix
```

</details>

<details>
<summary>Arch (AUR) - Unofficial</summary>
  
There's also an unofficial AUR package for squix available at: [squix-bin](https://aur.archlinux.org/packages/squix-bin)
  
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

No database handy? Spin up a sample one to explore right away:

```bash
# Create an Office-themed SQLite database in the current directory
squix example
squix init --name example --type sqlite --conn example.db
squix explore
squix run "select * from employees"
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

# Search around
/          # Search cell's contents (n/N to cycle through results)
f          # Search column names (,/; to cycle through results)

# Help
H          # Show keyboard shortcuts overlay
?          # Toggle footer keybinds

# Exit
q          # Quit back to terminal
```

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/c46a2565-a58c-472c-9393-96724d9716da" />
    Database Support
</h2>

PostgreSQL, MySQL, MariaDB, SQL Server, SQLite, Oracle, ClickHouse, Firebird, DuckDB

See connection init examples in [Database Support](docs/databases.md)

---

<h2>
    <img width="auto" height="24" alt="image" src="https://github.com/user-attachments/assets/be97c01c-9140-4503-92df-e1b6f2d7c31a" />
    Shell Completion
</h2>

Dynamic tab completion for bash, zsh, and fish that includes your saved queries and connections.

See [Shell Completion](docs/completion.md)

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle" src="https://github.com/user-attachments/assets/8f5037c9-e616-4065-adfc-cd598621c887" />
    Configuration
</h2>

Row limits, column widths, color schemes (`dracula`, `gruvbox`, `catppuccin-mocha`, etc.) and UI visibility options can be set through the config file at `~/.config/squix/config.yaml`.

Connections named `{service}:{env}` (e.g. `ecommerce:dev`, `ecommerce:prod`) share one saved-query library — add a query once, run it on any sibling environment. See [Shared query libraries](docs/configuration.md#shared-query-libraries).

See [Configuration](docs/configuration.md)

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/4b1425ae-7918-4a3f-b37c-41c3e443929e" />
    All Commands
</h2>

See [Commands](docs/commands.md) for the full command reference and database init examples

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/504a8488-69bf-43b4-860b-0659a6db3c69" />
    TUI Table Navigation
</h2>

See [Keybindings](docs/keybindings.md) for all navigation, editing, search, and visual mode keybindings on the results table view

---


<h2>
  <img width="auto" height="24" alt="image" src="https://github.com/user-attachments/assets/ebbabd15-87c4-47e5-b881-968b03e1d85d" />
    For Robots
</h2>

Squix ships a `SKILL.md` file in the repo root, a simple reference for AI
coding agents (Claude Code, Copilot, etc.) to use squix non-interactively. It
covers safe commands, format flags, parameterized queries, and which commands
to avoid (TUI/editor). Point your agent at it if you want it to run SQL queries
as part of your workflow.

---

<h2>
    <img width="auto" height="24" alt="image" style="vertical-align:middle;" src="https://github.com/user-attachments/assets/432c6b41-b2e0-4326-a3cc-7b349a987bb0" />
    Roadmap
</h2>

> This project is currently in beta, please report unexpected behavior through the issues tab

### v0.5.0 - Scrat 🧊
- [x] Configurable keybinds
- [x] Migrate to bubbletea v2
- [x] Return more info on exec statements (insert, update, delete, etc.)
- [x] Homebrew custom tap
- [x] nixpkgs entry
- [x] Handle windows env vars seamlessly

### v1.0.0 - Scrat 🧊
- [x] Move saved configs from top level yaml file to per-connection sql files
- [x] Display counters for selected lines and search results
- [x] Allow different connections to share query collection (eg. dev, stg, prod)
- [x] Rework database exploration commands
- [ ] Option to encrypt sensitive data in the config file

---
 
<h2>
    <img width="auto" height="24" alt="image" src="https://github.com/user-attachments/assets/cd243b3c-c39f-4e3f-9003-14ef6d713169" />
    Sponsor
</h2>

If Squix saves you time, consider supporting its development through [GitHub
Sponsors](https://github.com/sponsors/eduardofuncao). I work on Squix in my
spare time and your support helps me dedicate more hours to it, keep the [live
demo](https://squix.live.eduardofuncao.com) running, build a proper
docs/landing site, and cover GitHub Actions minutes for cross-platform builds,
instead of relying on local builds. 

Any amount is greatly appreciated ❤

---

<h2>
    <img width="auto" height="24" alt="image" src="https://github.com/user-attachments/assets/20aad60b-adc7-4a4d-971b-cb37f8a0cbbf" />
    Contributing & acknowledgements
</h2>

We welcome contributions! Get started with detailed instructions from [CONTRIBUTING.md](CONTRIBUTING.md)

Thanks a lot to all the contributors:

<a href="https://github.com/DeprecatedLuar"><img src="https://github.com/DeprecatedLuar.png" width="40" /></a>
<a href="https://github.com/caiolandgraf"><img src="https://github.com/caiolandgraf.png" width="40" /></a>
<a href="https://github.com/g4brielklein"><img src="https://github.com/g4brielklein.png" width="40" /></a>
<a href="https://github.com/eduardofuncao"><img src="https://github.com/eduardofuncao.png" width="40" /></a>
<a href="https://github.com/udirona"><img src="https://github.com/udirona.png" width="40" /></a>
<a href="https://github.com/Leosallin"><img src="https://github.com/Leosallin.png" width="40" /></a>
<a href="https://github.com/chasinglogic"><img src="https://github.com/chasinglogic.png" width="40" /></a>
<a href="https://github.com/akdev1l"><img src="https://github.com/akdev1l.png" width="40" /></a>
<a href="https://github.com/KalebJS"><img src="https://github.com/KalebJS.png" width="40" /></a>
<a href="https://github.com/Liferenko"><img src="https://github.com/Liferenko.png" width="40" /></a>

Squix wouldn't exist without the inspiration and groundwork laid by these fantastic projects:

- **[naggie/dstask](https://github.com/naggie/dstask)** - For the elegant CLI design patterns and file-based data storage approach
- **[DeprecatedLuar/better-curl-saul](https://github.com/DeprecatedLuar/better-curl-saul)** - For demonstrating a simple and genius approach to making a CLI tool
- **[dbeaver](https://github.com/dbeaver/dbeaver)** - The OG database management tool

Built with: 
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The TUI framework
- Go standard library and various database drivers

<div align="center">

**Made with 🐿️ by [@eduardofuncao](https://github.com/eduardofuncao)**

<img width="320" height="224" alt="Squix mascot" src="https://github.com/user-attachments/assets/f995ce07-3742-4e98-b737-bbdbf982012e" />

Previously Pam's Database Drawer, thanks to [u/marrsd](https://www.reddit.com/user/marrsd/) for suggesting the new name!

</div>
