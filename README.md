# wiz

Orchestrate AI coding agents across parallel branches. You're the conductor — wiz is the baton.

```bash
wiz create feat-auth   --task "Add OAuth login"       --agent claude
wiz create fix-payments --task "Fix Stripe webhooks"   --agent claude
wiz create refactor-db  --task "Migrate to sqlc"       --agent claude

wiz spawn feat-auth
wiz spawn fix-payments
wiz spawn refactor-db
# Three terminal tabs. Three branches. Three Claude sessions. Zero confusion.
```

## Why

You want to run 3 Claude Code sessions on 3 different features. That means 3 branches, 3 directories, 3 terminals, and remembering which is which. Without wiz, you're manually cloning repos, juggling paths, and losing track of what's where.

Wiz wraps **git worktrees** so each branch gets its own isolated working directory — instantly, with zero disk overhead. You create a context, spawn a terminal tab, and an agent is working. Your prompt tells you where you are. When agents finish, `wiz finish` pushes, creates a PR, and cleans up.

## Install

**Requires Go 1.25+**

```bash
go install github.com/buck3000/wiz@latest
```

Or build from source:

```bash
git clone https://github.com/buck3000/wiz.git
cd wiz
go build -o wiz .
sudo mv wiz /usr/local/bin/   # or anywhere on your PATH
```

Verify it works:

```bash
wiz doctor
```

## Shell Setup

Add to your shell rc file so `wiz enter` works and you get prompt/title enhancements:

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

Restart your shell (or `source` the file), then run `wiz doctor` to confirm everything is wired up.

This gives you:
- A `wiz` shell function that handles `wiz enter` properly
- Automatic prompt prefix: `🧙 feat-auth*` (with dirty indicator)
- Terminal title: `🧙 feat-auth — myapp`
- iTerm2 badge support (automatic when detected)

## AI Agent Workflows

### Single agent

Create a context, spawn an agent, come back later:

```bash
wiz create feat-auth --task "Add OAuth login with Google" --agent claude
wiz spawn feat-auth
# Claude Code opens in a new tab, working on the task
# You go do something else
wiz diff feat-auth --stat   # check progress whenever you want
```

### Multiple agents in parallel

The core workflow — run several agents at once, each on its own branch:

```bash
# Set up three contexts
wiz create feat-auth    --task "Add OAuth login"            --agent claude
wiz create fix-payments --task "Fix Stripe webhook retries" --agent claude
wiz create refactor-db  --task "Migrate queries to sqlc"    --agent claude

# Launch all three
wiz spawn feat-auth
wiz spawn fix-payments
wiz spawn refactor-db

# Monitor from your main terminal
wiz watch                  # live dashboard, refreshes every 2s
wiz diff --all             # what changed across all contexts
wiz log --all              # recent commits in every context
wiz list --tasks           # what each agent is working on
```

### Orchestrated plan with dependencies

For workflows where tasks depend on each other, define a YAML plan:

```yaml
# plan.yaml
tasks:
  - name: auth-backend
    agent: claude
    prompt: "Implement OAuth2 backend with Google provider"

  - name: auth-frontend
    agent: claude
    prompt: "Build login page with Google OAuth button"

  - name: auth-tests
    agent: claude
    prompt: "Write integration tests for the OAuth flow"
    depends_on: [auth-backend, auth-frontend]
```

```bash
wiz orchestra plan.yaml
# Creates contexts, spawns agents, respects dependency order
```

## Monitoring Your Agents

Once agents are running, you have full visibility from your main terminal:

```bash
wiz watch                      # Live dashboard (auto-refreshes)
wiz list --tasks               # All contexts with task descriptions
wiz diff --all                 # Diff summary for every context
wiz diff feat-auth --stat      # What changed in one context
wiz log --all                  # Recent commits across all contexts
wiz log feat-auth              # Git log for a specific context
wiz status                     # Current context status
```

Enter any context to inspect agent work directly:

```bash
wiz enter feat-auth            # cd into the worktree
# look around, run tests, read the code
exit                           # back to where you were
```

## Reviewing and Shipping

When agents finish, review their work and ship the good stuff:

```bash
# Review what each agent did
wiz diff --all --stat          # Quick summary
wiz diff feat-auth             # Full diff for one context

# Ship it
wiz finish feat-auth           # Push, create PR, clean up

# Or ship with options
wiz finish feat-auth --title "Add OAuth" --body "Closes #42"
wiz finish feat-auth --merge   # Create PR and merge immediately

# Discard bad work
wiz delete fix-payments        # Just remove (no PR)
```

Requires [GitHub CLI](https://cli.github.com/) (`gh`) to be installed and authenticated.

## Git-style Flow

Wiz supports a git-like porcelain so the common path feels natural:

```bash
wiz checkout feat-x        # create context + enter (or just enter if it exists)
# edit files...
wiz add -A                 # stage changes (runs in context dir)
wiz commit -m "Add feature X"
wiz finish                 # push, create PR, clean up
```

`wiz checkout -` returns to the previous context (like `git checkout -`).

## Anything Git Can Do

Wiz never reduces what git can do. The `wiz git` escape hatch runs any git command inside the active context:

```bash
wiz git status
wiz git diff
wiz git log --oneline --decorate -20
wiz git rebase -i origin/main
wiz git cherry-pick abc123
wiz git reset --hard HEAD~1
wiz git stash push -m "wip"
wiz git bisect start
wiz git clean -fdx
```

Use `--ctx <name>` to target a specific context, or `--base-ok` to run in the base repo:

```bash
wiz git --ctx feat-auth log --oneline -5
wiz git --base-ok status   # run in base worktree
```

Alternatively, `wiz run <name> -- git ...` works the same way.

## Quick Start

Everything below assumes you're inside a git repo with at least one commit.

### 1. Create a context

```bash
wiz checkout feat-auth
# Creates an isolated worktree on a new branch "feat-auth" and enters it
```

Or use the explicit create command for more options:

```bash
wiz create feat-auth
# Creates an isolated worktree on a new branch "feat-auth"
```

Options:
- `--base main` — branch off a specific base branch
- `--task "Add OAuth login"` — attach a task description (used by agents and `wiz finish`)
- `--agent claude` — associate an agent for `wiz spawn`

### 2. Work in it

**Option A — Checkout (create + enter):**
```bash
wiz checkout feat-auth
# Creates if needed, then cd's into the worktree
```

**Option B — Enter an existing context:**
```bash
wiz enter feat-auth
# cd's into the worktree and sets up the environment
```

**Option C — Open a new terminal tab:**
```bash
wiz spawn feat-auth
# Opens a new iTerm2/Kitty/WezTerm/tmux tab, cd'd into the context
```

**Option D — Launch an AI agent directly:**
```bash
wiz spawn feat-auth --agent claude --prompt "Add OAuth login with Google"
# Opens a new tab and starts Claude Code with the prompt
```

### 3. Run multiple contexts in parallel

```bash
wiz create feat-auth --task "Add OAuth login"
wiz create fix-payments --task "Fix Stripe webhook retry logic"
wiz create refactor-db --task "Migrate to sqlc"

wiz spawn feat-auth --agent claude
wiz spawn fix-payments --agent claude
wiz spawn refactor-db --agent claude
# Three terminal tabs, three branches, three Claude sessions
```

### 4. Check on progress

```bash
wiz list                       # List all contexts
wiz status                     # Current context status
wiz diff feat-auth --stat      # What changed vs base branch
wiz diff --all                 # Diff summary for every context
wiz log feat-auth              # Git log for a context
wiz git log --oneline -10      # Any git command in active context
```

### 5. Run commands without entering

```bash
wiz run feat-auth -- make test
wiz run feat-auth -- git log --oneline -5
```

### 6. Finish up

```bash
wiz finish                 # Finish current context (resolved from env/CWD)
wiz finish feat-auth       # Finish by name

wiz finish --merge
# Same, but also merges the PR
```

Options:
- `--title "Add OAuth"` — custom PR title (default: context name)
- `--body "..."` — custom PR body (default: task description)

## Orchestra

Define a multi-task plan in YAML and run them all at once:

```yaml
# plan.yaml
tasks:
  - name: auth
    prompt: "Add OAuth login with Google"
    agent: claude
  - name: tests
    prompt: "Write integration tests for the auth module"
    agent: claude
    depends_on: [auth]
```

```bash
wiz orchestra plan.yaml
# Creates contexts, spawns agents, respects dependency order
```

Tasks with `depends_on` wait for their dependencies to be spawned first.

## Templates

Save and reuse context configurations:

```bash
wiz template save my-template --base main --strategy worktree --agent claude
wiz template list
wiz create feat-x --template my-template
```

## Commands

| Command | Description |
|---------|-------------|
| `wiz` | Launch interactive TUI picker |
| `wiz checkout <name>` | Switch to context (create if needed) |
| `wiz checkout -` | Return to previous context |
| `wiz create <name>` | Create a new context |
| `wiz list [--json]` | List all contexts |
| `wiz enter <name>` | Activate context in current shell |
| `wiz spawn <name> [--agent <name>] [--prompt <text>]` | Open new terminal tab in context |
| `wiz add <args...>` | Stage files in active context (`git add`) |
| `wiz commit <args...>` | Commit in active context (`git commit`) |
| `wiz git <args...>` | Run any git command in active context |
| `wiz run <name> -- <cmd...>` | Run command inside context |
| `wiz diff <name> [--stat] [--all]` | Show diff vs base branch |
| `wiz log <name> [-n N] [--all]` | Show git log for a context |
| `wiz watch [--interval <dur>]` | Live monitoring dashboard |
| `wiz finish [<name>] [--merge]` | Push, create PR, clean up |
| `wiz orchestra <file.yaml>` | Run multi-task plan |
| `wiz path <name>` | Print context filesystem path |
| `wiz rename <old> <new>` | Rename a context |
| `wiz delete <name> [--force]` | Delete a context |
| `wiz status [--porcelain] [--json]` | Show current context status |
| `wiz template save\|list\|delete` | Manage context templates |
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

## Support

Wiz is free to use with up to 10 concurrent contexts. If you find it useful, consider sponsoring the project to unlock unlimited contexts and support development:

[Sponsor on GitHub](https://github.com/sponsors/buck3000)

## Testing

```bash
go test ./... -race
```

## License

MIT
