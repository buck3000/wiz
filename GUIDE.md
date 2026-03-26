# Wiz: 0-to-60 Power User Guide

Wiz wraps git worktrees into named **contexts** — isolated working directories with their own branch, index, and HEAD. You create contexts, point AI agents at them, monitor progress, and ship the results. Three Claude Code sessions, three feature branches, three terminal tabs — all managed from one place.

This guide covers everything, from first install to orchestrating multi-agent workflows.

---

## 1. Setup (2 minutes)

### Install

```bash
# From source
go install github.com/buck3000/wiz@latest

# Or build locally
git clone https://github.com/buck3000/wiz.git
cd wiz && go build -o wiz . && mv wiz ~/.local/bin/
```

### Shell Integration (required)

Add to your shell rc file:

```bash
# zsh (~/.zshrc)
eval "$(wiz init zsh)"

# bash (~/.bashrc)
eval "$(wiz init bash)"

# fish (~/.config/fish/config.fish)
wiz init fish | source
```

This does three things:
1. Wraps `wiz enter` and `wiz checkout` so they can `cd` your shell
2. Adds a prompt hook that shows your active context: `🧙 feat-auth*`
3. Sets your terminal title to `🧙 context-name — repo-name`

### Verify

```bash
wiz doctor
```

You want green checks on Git and Shell integration. Terminal detection is a bonus (iTerm2 gets badges and tab colors for free).

---

## 2. Core Workflow: checkout / work / finish

This is the git-like flow most people should start with:

```bash
cd my-project

# Create and enter a context (creates worktree + branch automatically)
wiz checkout feat-auth

# You're now in an isolated copy of your repo.
# Edit files, run tests — everything is scoped to this branch.
echo "hello" > new-file.txt
wiz add -A
wiz commit -m "Add new file"

# Push, create PR, clean up — one command
wiz finish
```

`wiz checkout` is idempotent: if the context already exists, it just enters it. If it doesn't exist, it creates it first. This mirrors `git checkout -b` behavior.

### Switch between contexts

```bash
wiz checkout feat-auth       # Enter feat-auth
wiz checkout fix-payments    # Switch to fix-payments
wiz checkout -               # Go back to feat-auth (like cd -)
```

`checkout -` reads from `.git/wiz/last_context` — it always remembers where you just were.

---

## 3. How Contexts Work Under the Hood

When you run `wiz create feat-auth`, here's what actually happens:

1. **Worktree creation**: `git worktree add --detach .git/wiz/trees/feat-auth HEAD`
2. **Branch creation**: `git checkout -b feat-auth` inside the worktree
3. **State persistence**: Context metadata written to `.git/wiz/state.json`
4. **File locking**: All state mutations go through `flock` on `.git/wiz/wiz.lock`

The worktree shares the same `.git/objects` store as your main repo — no disk duplication. Creation is instant regardless of repo size.

### State file

Every context is tracked in `.git/wiz/state.json`:

```json
{
  "version": 1,
  "contexts": [{
    "name": "feat-auth",
    "branch": "feat-auth",
    "path": "/Users/you/project/.git/wiz/trees/feat-auth",
    "strategy": "worktree",
    "baseBranch": "main",
    "task": "Add OAuth login",
    "agent": "claude",
    "createdAt": "2026-03-15T10:30:00Z"
  }]
}
```

This file is atomically written (write to temp, rename) and protected by file lock for concurrent access.

---

## 4. Context Resolution: How Wiz Knows Where You Are

Many commands need to know which context is active. Resolution follows this priority:

1. **`--ctx` flag** — explicit override: `wiz add --ctx feat-auth -A`
2. **`WIZ_CTX` environment variable** — set by `wiz enter` / shell integration
3. **Current working directory** — if your CWD is inside a context's path
4. **Error** — "not in a wiz context"

This means you can always target any context from anywhere with `--ctx`, or just be inside one and it works automatically.

---

## 5. The Git Escape Hatch

`wiz add` and `wiz commit` cover the basics, but sometimes you need full git. That's `wiz git`:

```bash
# Run any git command in the active context
wiz git status
wiz git log --oneline -10
wiz git stash
wiz git rebase -i origin/main

# Target a specific context from anywhere
wiz git --ctx feat-auth diff --stat
wiz git --ctx fix-payments cherry-pick abc123

# Run git in the base repo (normally blocked for safety)
wiz git --base-ok fetch --all
```

Every flag after the wiz-specific ones (`--ctx`, `--base-ok`) is passed through to git exactly as-is. Stdin/stdout/stderr are passed through directly — interactive commands like `git rebase -i` work.

### Safety: base worktree protection

`wiz add` and `wiz commit` refuse to run in the base worktree (your main checkout). This prevents accidentally committing to `main` when you think you're in a context. Use `wiz git --base-ok` to explicitly override this.

---

## 6. Spawning Agents in Terminal Tabs

This is where wiz shines — launching AI agents across parallel branches from a single terminal.

### The basics

```bash
# Open feat-auth in a new terminal tab
wiz spawn feat-auth

# Open it with Claude Code running
wiz spawn feat-auth --agent claude

# Open it with Claude Code and a specific prompt
wiz spawn feat-auth --agent claude --prompt "Add OAuth login with Google"
```

### End-to-end: 3 agents, 3 features

Here's the full workflow for running multiple agents in parallel:

```bash
# 1. Create contexts with tasks and agent assignments
wiz create feat-auth    --task "Add OAuth login with Google provider" --agent claude --base main
wiz create fix-payments --task "Fix Stripe webhook retry logic"       --agent claude --base main
wiz create refactor-db  --task "Migrate raw SQL queries to sqlc"      --agent claude --base main

# 2. Launch all three agents
wiz spawn feat-auth
wiz spawn fix-payments
wiz spawn refactor-db
# Three new terminal tabs open, each with Claude Code working on its task

# 3. Monitor from your main terminal while agents work
wiz watch                      # live dashboard — see all contexts and their status
wiz list --tasks               # check what each agent is working on
wiz diff --all                 # quick diff summary across all contexts
wiz log --all                  # see what commits agents have made

# 4. Dive into a specific context if you need details
wiz diff feat-auth --stat      # what files changed?
wiz diff feat-auth             # full diff
wiz log feat-auth              # commit history
wiz enter feat-auth            # cd into the worktree to look around
exit                           # back to your main terminal

# 5. When agents finish, review and ship
wiz diff --all --stat          # final summary of all changes
wiz finish feat-auth           # push + create PR + clean up
wiz finish fix-payments
wiz finish refactor-db
```

### Terminal detection

Wiz auto-detects your terminal:
- **iTerm2** — full support: custom tab title, badge, tab color
- **Kitty** — title support
- **WezTerm** — title support
- **tmux** — new pane with title
- **Generic** — basic fallback

Each spawned tab gets the full context environment (`WIZ_CTX`, `WIZ_DIR`, etc.) and shell integration active.

### Built-in agents

Three agents are recognized by name:
- `claude` — Claude Code
- `gemini` — Google Gemini CLI
- `codex` — OpenAI Codex CLI

You can define custom agents in `.wiz/config.yaml`:

```yaml
agents:
  my-agent:
    command: /usr/local/bin/my-agent
    args: ["--config", "~/.my-agent.yaml"]
```

---

## 7. Running Commands Across Contexts

`wiz run` executes any command inside a context's directory:

```bash
# Run tests in a specific context
wiz run feat-auth -- make test

# Run an arbitrary command
wiz run fix-payments -- git log --oneline -5

# Run an agent directly (without a terminal tab)
wiz run feat-auth --agent claude --prompt "Fix the failing tests"
```

The `--` separator is required for raw commands. Everything after it is passed through exactly. The command runs with context environment variables set.

---

## 8. Inspecting Your Contexts

### List all contexts

```bash
wiz list              # Human-readable
wiz list --tasks      # Include task descriptions and agents
wiz list --json       # Machine-readable JSON
```

The current context is marked with `▸`.

### Status of active context

```bash
wiz status              # Full human-readable output
wiz status --porcelain  # One-line: ctx repo branch state ahead behind
wiz status --json       # Full JSON with staged/unstaged/untracked counts
```

The porcelain format is what the shell prompt hook uses internally — it's designed to be fast and cacheable.

### Diffs and logs

```bash
# Diff against the base branch (3-dot diff)
wiz diff feat-auth          # Full diff
wiz diff feat-auth --stat   # Summary only
wiz diff --all              # All contexts at a glance

# Git log since branching from base
wiz log feat-auth           # Last 10 commits
wiz log feat-auth -n 5      # Last 5
wiz log --all               # All contexts
```

These use `git diff base...branch` and `git log base..branch` — you see exactly what's changed since the context branched off.

---

## 9. Finishing: Push, PR, Clean Up

```bash
wiz finish                           # Current context
wiz finish feat-auth                 # Specific context
wiz finish --title "Add OAuth"       # Custom PR title
wiz finish --body "Closes #42"       # Custom PR body
wiz finish --merge                   # Create PR and merge immediately
```

Under the hood, `finish` does:
1. `git push -u origin <branch>`
2. `gh pr create --title <title> --head <branch> --base <base>`
3. (optional) `gh pr merge <url> --merge --delete-branch`
4. Destroys the worktree and removes from state

Requires GitHub CLI (`gh`) installed and authenticated.

---

## 10. Orchestration: Multi-Agent Plans

For coordinating multiple agents with dependencies, use a YAML plan:

```yaml
# plan.yaml
tasks:
  - name: auth-backend
    agent: claude
    prompt: "Implement OAuth2 backend with Google provider"
    base: main

  - name: auth-frontend
    agent: claude
    prompt: "Build login page with Google OAuth button"
    base: main

  - name: auth-tests
    agent: claude
    prompt: "Write integration tests for the OAuth flow"
    depends_on: [auth-backend, auth-frontend]

  - name: auth-docs
    agent: claude
    prompt: "Document the OAuth setup process"
    depends_on: [auth-backend]
```

```bash
wiz orchestra plan.yaml
```

This creates all four contexts, then spawns agents respecting dependency order. `auth-backend` and `auth-frontend` spawn immediately (no deps). `auth-tests` waits for both to spawn. `auth-docs` waits for `auth-backend`.

### A larger real-world example

A full-stack feature with backend, frontend, tests, and documentation:

```yaml
# feature-plan.yaml
tasks:
  # Backend work — no dependencies, starts immediately
  - name: api-endpoints
    agent: claude
    prompt: |
      Add REST endpoints for user notifications:
      - GET /api/notifications (list, paginated)
      - POST /api/notifications/:id/read (mark as read)
      - POST /api/notifications/read-all
      Use the existing service object pattern in app/services/.
    base: main

  - name: notification-model
    agent: claude
    prompt: |
      Create the Notification model with:
      - belongs_to :user
      - belongs_to :source, polymorphic: true
      - scopes: unread, recent
      - migration with proper indexes
    base: main

  # Frontend — depends on API being defined
  - name: notification-ui
    agent: claude
    prompt: |
      Build a notification dropdown component:
      - Bell icon with unread count badge
      - Dropdown list with infinite scroll
      - Mark as read on click
      - "Mark all as read" button
    depends_on: [api-endpoints]

  # Tests — depend on both backend pieces
  - name: notification-tests
    agent: claude
    prompt: |
      Write tests for the notification system:
      - Model specs for Notification
      - Request specs for the API endpoints
      - Service specs for NotificationService
    depends_on: [api-endpoints, notification-model]

  # Docs — depend on everything
  - name: notification-docs
    agent: claude
    prompt: "Document the notification system: API endpoints, model, frontend component"
    depends_on: [api-endpoints, notification-model, notification-ui]
```

```bash
# Launch the full plan
wiz orchestra feature-plan.yaml

# Monitor progress
wiz watch                      # live dashboard
wiz list --tasks               # see what each agent is working on
wiz diff --all --stat          # see what's changed

# When everything finishes, review and ship
wiz diff notification-model    # check the migration
wiz diff api-endpoints         # check the API
wiz finish api-endpoints
wiz finish notification-model
wiz finish notification-ui
wiz finish notification-tests
wiz finish notification-docs
```

### Plan fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | yes | Context name |
| `agent` | yes | Agent to run |
| `prompt` | no | Task prompt for the agent |
| `base` | no | Base branch (default: HEAD) |
| `branch` | no | Branch name (default: same as name) |
| `strategy` | no | `worktree` or `clone` |
| `depends_on` | no | List of task names to wait for |

---

## 11. Reviewing Agent Work

After agents finish their tasks, you need to review what they did before shipping. Here's how to do that efficiently.

### Quick triage across all contexts

```bash
wiz diff --all --stat          # One-line summary of changes per context
wiz log --all                  # Recent commits across all contexts
wiz list --tasks               # Reminder of what each agent was asked to do
```

This gives you the bird's-eye view: which contexts have changes, how large they are, and what commits were made.

### Deep-dive into a specific context

```bash
# Read the diff
wiz diff feat-auth             # Full diff against base branch
wiz diff feat-auth --stat      # File-level summary

# Read the commit history
wiz log feat-auth              # What the agent committed and when
wiz log feat-auth -n 20        # More history if needed

# Enter the context and look around
wiz enter feat-auth
ls -la                         # see the file layout
make test                      # run the test suite
exit                           # back to main terminal

# Or run commands without entering
wiz run feat-auth -- make test
wiz run feat-auth -- go vet ./...
```

### Decide: ship, fix, or discard

For each context, you have three choices:

**Ship it** — the agent did good work:
```bash
wiz finish feat-auth                          # push + PR + clean up
wiz finish feat-auth --title "Add OAuth"      # with a custom PR title
```

**Fix it** — mostly good, needs tweaks:
```bash
wiz enter feat-auth        # go into the context
# make your edits
wiz add -A
wiz commit -m "Fix edge case in OAuth callback"
wiz finish                 # then ship
```

**Discard it** — not usable:
```bash
wiz delete feat-auth       # remove the context entirely
wiz delete feat-auth --force  # skip confirmation
```

---

## 12. Patterns for AI-Assisted Development

### Spike-and-choose

When you're not sure of the best approach, have two agents try different strategies and pick the winner:

```bash
# Two agents, two approaches to the same problem
wiz create cache-redis   --task "Add caching layer using Redis"        --agent claude
wiz create cache-memory  --task "Add caching layer using in-memory LRU" --agent claude

wiz spawn cache-redis
wiz spawn cache-memory

# Compare the results
wiz diff cache-redis --stat
wiz diff cache-memory --stat
wiz diff cache-redis          # read the full implementation
wiz diff cache-memory         # compare approaches

# Ship the one you prefer, discard the other
wiz finish cache-redis
wiz delete cache-memory
```

### Pipeline: sequential dependent tasks

When later work depends on earlier work, use `depends_on` to build a pipeline:

```yaml
# pipeline.yaml
tasks:
  - name: auth-backend
    agent: claude
    prompt: "Implement OAuth2 backend with Google provider"

  - name: auth-frontend
    agent: claude
    prompt: "Build the login page — the backend OAuth endpoints are already implemented"
    depends_on: [auth-backend]

  - name: auth-e2e-tests
    agent: claude
    prompt: "Write end-to-end tests for the full OAuth login flow"
    depends_on: [auth-frontend]
```

```bash
wiz orchestra pipeline.yaml
# auth-backend starts immediately
# auth-frontend starts after auth-backend spawns
# auth-e2e-tests starts after auth-frontend spawns
```

### Review loop: agent writes, you review, agent fixes

Use contexts as a tight feedback loop with an agent:

```bash
# Round 1: agent writes the initial implementation
wiz create feat-search --task "Add full-text search to the API" --agent claude
wiz spawn feat-search
# ... agent works ...

# Review the work
wiz diff feat-search --stat
wiz diff feat-search

# Round 2: enter the context and give the agent feedback
wiz enter feat-search
claude "The search implementation looks good but it's missing pagination. Add cursor-based pagination to the search endpoint."
# ... agent fixes ...

# Round 3: check again
wiz diff feat-search --stat
# Looks good — ship it
wiz finish feat-search
```

### Batch cleanup: finish or discard everything

After a big orchestration run, clean up all contexts at once:

```bash
# See what's still open
wiz list --tasks

# Finish the ones you want to keep
wiz finish auth-backend
wiz finish auth-frontend

# Discard the rest
wiz delete auth-tests
wiz delete auth-docs
```

---

## 13. Templates: Reusable Context Configs

Save common configurations as templates:

```bash
# Save a template
wiz template save claude-feature --base main --agent claude --strategy worktree

# Use it
wiz create feat-auth --template claude-feature --task "Add OAuth"

# List templates
wiz template list
wiz template list --json

# Delete
wiz template delete claude-feature
```

Templates store base branch, agent, and strategy — the things you'd otherwise type every time.

---

## 14. Interactive Mode

Run `wiz` with no arguments for an interactive context picker:

```bash
wiz
```

This launches a TUI where you can browse contexts and choose actions (enter, spawn, delete, create).

For a live dashboard that auto-refreshes:

```bash
wiz watch                   # Default: refresh every 2s
wiz watch --interval 5s     # Custom interval
```

The dashboard shows all contexts with their branch, status, and path — useful for monitoring parallel agent sessions.

---

## 15. Environment Variables Reference

### Set by Wiz (inside a context)

| Variable | Example | Description |
|----------|---------|-------------|
| `WIZ_CTX` | `feat-auth` | Active context name |
| `WIZ_REPO` | `my-project` | Repository name |
| `WIZ_DIR` | `/path/to/worktree` | Context directory path |
| `WIZ_BRANCH` | `feat-auth` | Git branch name |
| `WIZ_PROMPT` | `🧙 feat-auth*` | Formatted prompt (set by shell hook) |

### User-configurable

| Variable | Effect |
|----------|--------|
| `WIZ_NO_PROMPT` | Disable prompt customization |
| `WIZ_CTX` | Override active context for resolution |

---

## 16. Scripting & Automation

Wiz's JSON outputs make it scriptable:

```bash
# Get all context paths
wiz list --json | jq -r '.[].path'

# Run tests in every context
wiz list --json | jq -r '.[].name' | while read ctx; do
  echo "Testing $ctx..."
  wiz run "$ctx" -- make test
done

# Find dirty contexts
wiz list --json | jq -r '.[].name' | while read ctx; do
  status=$(wiz git --ctx "$ctx" status --porcelain)
  [ -n "$status" ] && echo "$ctx is dirty"
done

# Get path for shell use
cd "$(wiz path feat-auth)"
```

---

## 17. Provisioning Strategies

### Worktree (default, recommended)

```bash
wiz create feat-auth --strategy worktree
```

Uses `git worktree add`. Instant creation, shares object store, zero disk overhead. This is what you want 99% of the time.

### Clone

```bash
wiz create feat-auth --strategy clone
```

Uses `git clone --shared`. Creates a separate clone that still shares the object store. Useful as a fallback if worktrees cause issues.

### Auto

```bash
wiz create feat-auth --strategy auto
```

Tries worktree first, falls back to clone. This is the default.

---

## 18. Tips & Patterns

### Parallel feature development

```bash
wiz checkout feat-auth
# ... work on auth in this tab ...

# Open a new tab for payments (without leaving this one)
wiz spawn fix-payments --agent claude --prompt "Fix the Stripe webhook retry logic"

# Check on it later
wiz diff fix-payments --stat
wiz log fix-payments
```

### Quick context hopping

```bash
wiz checkout auth     # work on auth
wiz checkout payments # switch to payments
wiz checkout -        # back to auth
wiz checkout -        # back to payments
```

### Inspect everything at once

```bash
wiz diff --all        # What changed in every context
wiz log --all         # Recent commits across all contexts
wiz list --tasks      # What each context is working on
```

### Clean up finished work

```bash
wiz finish feat-auth              # Push + PR + delete
wiz delete old-experiment         # Just delete (no PR)
wiz delete --all --force          # Nuclear option
```

### Use with Claude Code directly

```bash
# Create context, spawn Claude with a task
wiz create refactor-api --task "Refactor the API layer to use service objects" --agent claude
wiz spawn refactor-api

# Or one-shot: create + enter + you're in the context directory
wiz checkout refactor-api
claude "Refactor the API layer to use service objects"
```

---

## Command Reference

| Command | Description |
|---------|-------------|
| `wiz` | Interactive context picker (TUI) |
| `wiz checkout <name>` | Create (if needed) and enter context |
| `wiz checkout -` | Return to previous context |
| `wiz create <name>` | Create a new context |
| `wiz enter <name>` | Activate context in current shell |
| `wiz spawn <name>` | Open context in new terminal tab |
| `wiz add [args...]` | Stage files (git pass-through) |
| `wiz commit [args...]` | Commit changes (git pass-through) |
| `wiz git [args...]` | Full git escape hatch |
| `wiz run <name> -- <cmd>` | Run command in context directory |
| `wiz finish [name]` | Push, create PR, delete context |
| `wiz list` | List all contexts |
| `wiz status` | Show active context status |
| `wiz diff <name>` | Diff context vs base branch |
| `wiz log <name>` | Show commits since base |
| `wiz watch` | Live monitoring dashboard |
| `wiz delete <name>` | Delete a context |
| `wiz rename <old> <new>` | Rename a context |
| `wiz path <name>` | Print context filesystem path |
| `wiz orchestra <file>` | Run multi-task YAML plan |
| `wiz template save\|list\|delete` | Manage templates |
| `wiz doctor` | Check environment |
| `wiz init <shell>` | Print shell integration script |

---

## Licensing

- **Free tier**: 10 concurrent contexts. All features included.
- **Pro tier**: Unlimited contexts. Available via [GitHub Sponsors](https://github.com/sponsors/buck3000).
