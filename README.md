# wiz

Magical git branch contexts. Work on multiple branches simultaneously across terminal windows with zero confusion.

```
wiz create feat-auth
wiz create bugfix-login
wiz spawn feat-auth       # opens new terminal tab
wiz spawn bugfix-login    # opens another tab
# Each tab shows: 🧙 feat-auth — myapp
# Run Claude Code in parallel on different branches
```

## Install

```bash
go install github.com/buck3000/wiz@latest
```

Or build from source:

```bash
git clone https://github.com/buck3000/wiz.git
cd wiz
go build -o wiz .
```

## Shell Setup

Add to your shell rc file:

**zsh** (`~/.zshrc`):
```zsh
eval "$(wiz init zsh)"
```

**bash** (`~/.bashrc`):
```bash
eval "$(wiz init bash)"
```

**fish** (`~/.config/fish/config.fish`):
```fish
wiz init fish | source
```

This gives you:
- A `wiz` shell function that handles `wiz enter` properly
- Automatic prompt prefix: `🧙 feat-auth*` (with dirty indicator)
- Terminal title: `🧙 feat-auth — myapp`
- iTerm2 badge support (automatic when detected)

## Workflows

### Basic: Create and enter a context

```bash
wiz create feat-auth
wiz enter feat-auth
# You're now in an isolated working directory on the feat-auth branch
```

### Concurrent Claude Code sessions

```bash
wiz create feat-auth
wiz create bugfix-login --base main
wiz spawn feat-auth       # New terminal tab → cd into feat-auth context
wiz spawn bugfix-login    # Another tab → cd into bugfix-login context
# Run `claude` in each tab independently
```

### Quick status across all contexts

```bash
wiz status          # Show current context status
wiz list            # List all contexts
wiz list --json     # Machine-readable output
```

### Run a command in a context without entering it

```bash
wiz run feat-auth -- git log --oneline -5
wiz run feat-auth -- make test
```

### Clean up

```bash
wiz delete feat-auth
wiz delete bugfix-login --force  # Skip dirty check
```

## Commands

| Command | Description |
|---------|-------------|
| `wiz` | Launch interactive TUI picker |
| `wiz create <name> [--base <branch>] [--strategy auto\|worktree\|clone]` | Create a new context |
| `wiz list [--json]` | List all contexts |
| `wiz enter <name>` | Activate context in current shell |
| `wiz spawn <name>` | Open new terminal tab in context |
| `wiz run <name> -- <cmd...>` | Run command inside context |
| `wiz path <name>` | Print context filesystem path |
| `wiz rename <old> <new>` | Rename a context |
| `wiz delete <name> [--force]` | Delete a context |
| `wiz status [--porcelain]` | Show current context status |
| `wiz init <bash\|zsh\|fish>` | Print shell integration script |
| `wiz doctor` | Check environment and show active enhancements |

## How It Works

Under the hood, `wiz` uses **git worktrees** to create isolated working directories that share the same object store. This means:

- Contexts are instant to create (no cloning)
- All contexts share git objects (disk efficient)
- Each context has its own working tree, index, and HEAD
- Multiple terminals can operate on different branches without conflicts

State is stored in `<repo>/.git/wiz/`:
- `state.json` — context registry
- `wiz.lock` — file lock for concurrent safety
- `trees/` — worktree directories

A **clone strategy** is available as a fallback (`--strategy clone`) for edge cases where worktrees aren't suitable. It uses `git clone --shared` for object sharing.

## Terminal Enhancements

`wiz doctor` shows which enhancements are active:

```
 ✓ Git: git version 2.43.0
 ✓ Terminal: iTerm2 (features: title, badge, tab-color)
 ✓ Shell integration: Active (WIZ_CTX set)
 ✓ Active context: feat-auth
```

| Terminal | Title | Badge | Tab Color |
|----------|-------|-------|-----------|
| iTerm2 | Yes | Yes | Yes |
| Kitty | Yes | — | — |
| WezTerm | Yes | — | — |
| tmux | Yes | — | — |
| VS Code | Yes | — | — |

## Environment Variables

When inside a context, these are exported:

| Variable | Description |
|----------|-------------|
| `WIZ_CTX` | Context name |
| `WIZ_REPO` | Repository name |
| `WIZ_DIR` | Context directory path |
| `WIZ_BRANCH` | Git branch name |
| `WIZ_PROMPT` | Formatted prompt string (set by hook) |

## Testing

```bash
go test ./... -race
```
