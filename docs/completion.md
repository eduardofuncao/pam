# Shell Completion

Squix provides intelligent tab completion for bash, zsh, and fish. Completions are **dynamic** - they automatically include your saved queries and connections.

## Installation

<details>
<summary>Bash</summary>

```bash
# Temporary (current session)
source <(squix completion bash)

# Permanent (add to ~/.bashrc)
echo 'eval "$(squix completion bash)"' >> ~/.bashrc
```
</details>

<details>
<summary>Zsh</summary>

```bash
# Temporary (current session)
autoload -U compinit && compinit
source <(squix completion zsh)

# Permanent (add to ~/.zshrc)
echo 'autoload -U compinit && compinit' >> ~/.zshrc
echo 'eval "$(squix completion zsh)"' >> ~/.zshrc
```
</details>

<details>
<summary>Fish</summary>

```bash
# Fish loads completions automatically from ~/.config/fish/completions/
squix completion fish > ~/.config/fish/completions/squix.fish
# Restart your shell or run: exec fish
```
</details>

## Usage

After installation, press TAB to autocomplete:

```bash
squix [TAB]              # List all commands
squix run [TAB]          # List queries from current connection
squix switch [TAB]       # List connection names
squix list [TAB]         # List: queries
squix edit [TAB]         # List queries to edit
```

**Note:** Query completion draws from your current connection's query library — shared across `{service}:{env}` siblings (e.g. `ecommerce:dev`, `ecommerce:prod` share one library). Use `squix switch <connection>` to change connections.

> Currently unreleased, build from source to make this available
