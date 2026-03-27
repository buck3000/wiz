# Wiz Cheatsheet

## Setup

```bash
go install github.com/buck3000/wiz@latest
eval "$(wiz init zsh)"    # or bash / fish
wiz doctor                # verify everything works
```

## Single Agent

```bash
wiz create feat-auth --task "Add OAuth login" --agent claude
wiz spawn feat-auth           # opens tab, starts Claude
wiz diff feat-auth --stat     # check progress
wiz finish feat-auth          # push + PR + clean up
```

## Multiple Agents in Parallel

```bash
# Create
wiz create feat-auth    --task "Add OAuth"        --agent claude
wiz create fix-payments --task "Fix webhooks"      --agent claude
wiz create refactor-db  --task "Migrate to sqlc"   --agent claude

# Launch
wiz spawn feat-auth
wiz spawn fix-payments
wiz spawn refactor-db

# Monitor
wiz watch                     # live dashboard
wiz diff --all                # all diffs at a glance
wiz log --all                 # all commits at a glance
wiz list --tasks              # what each agent is doing

# Ship
wiz finish feat-auth
wiz finish fix-payments
wiz finish refactor-db
```

## Orchestra (dependency graph)

```yaml
# plan.yaml
tasks:
  - name: backend
    agent: claude
    prompt: "Build the API"
  - name: frontend
    agent: claude
    prompt: "Build the UI"
    depends_on: [backend]
```

```bash
wiz orchestra plan.yaml
```

## Git-style Flow (manual work)

```bash
wiz checkout feat-x           # create + enter
wiz add -A                    # stage
wiz commit -m "message"       # commit
wiz finish                    # push + PR + clean up
```

## Context Navigation

```bash
wiz checkout feat-auth        # enter (create if needed)
wiz checkout -                # previous context (like cd -)
wiz enter feat-auth           # enter existing context
wiz spawn feat-auth           # open in new tab
```

## Inspect

```bash
wiz list                      # all contexts
wiz list --tasks              # with task descriptions
wiz status                    # current context
wiz diff feat-auth            # full diff vs base
wiz diff feat-auth --stat     # summary
wiz diff --all                # all contexts
wiz log feat-auth             # commits since base
wiz log --all                 # all contexts
wiz watch                     # live dashboard
```

## Run Commands in a Context

```bash
wiz run feat-auth -- make test
wiz git status                # git in active context
wiz git --ctx feat-auth diff  # git in specific context
wiz git --base-ok status      # git in base repo
```

## Clean Up

```bash
wiz finish feat-auth          # push + PR + delete
wiz finish --merge            # push + PR + merge + delete
wiz delete feat-auth          # just delete (no PR)
wiz delete --all --force      # nuclear option
```

## Templates

```bash
wiz template save my-tpl --base main --agent claude
wiz create feat-x --template my-tpl --task "Do the thing"
wiz template list
wiz template delete my-tpl
```

## Environment Variables (set inside a context)

| Variable | Description |
|----------|-------------|
| `WIZ_CTX` | Context name |
| `WIZ_REPO` | Repo name |
| `WIZ_DIR` | Worktree path |
| `WIZ_BRANCH` | Branch name |

## Patterns

**Spike-and-choose** — two agents, two approaches, pick the best:
```bash
wiz create cache-redis  --task "Cache with Redis" --agent claude
wiz create cache-memory --task "Cache with LRU"   --agent claude
wiz spawn cache-redis && wiz spawn cache-memory
# compare, then: wiz finish cache-redis && wiz delete cache-memory
```

**Review loop** — agent writes, you review, agent fixes:
```bash
wiz create feat-x --task "Build search" --agent claude
wiz spawn feat-x
# ... agent works ...
wiz diff feat-x
wiz enter feat-x
claude "Add pagination to the search endpoint"
# ... agent fixes ...
wiz finish
```
