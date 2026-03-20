# Rocket Fuel

A multi-agent orchestrator that multiplies developer productivity by composing existing tools — Claude Code, GitHub Projects, tmux-CC/iTerm2, and git worktrees.

Based on the Visionary/Integrator model from *Rocket Fuel* by Gino Wickman & Mark C. Winters.

## The idea

You're the **Visionary** — ideas, direction, product thinking. You talk to the **Integrator**, an AI agent that manages execution: spawning workers, tracking progress, protecting the current sprint from your destabilising brilliance.

Workers are ephemeral Claude Code instances running in isolated git worktrees, each picking up a GitHub issue and delivering a PR using your existing skills (`/tdd`, `/bug-fix`, `/epc`).

The Integrator never says no. It scopes your ideas, parks them in Someday/Maybe, and redirects back to the current epic.

Read [`docs/vision.md`](docs/vision.md) for the full concept.

## Prerequisites

- [tmux](https://github.com/tmux/tmux) — `brew install tmux`
- [iTerm2](https://iterm2.com/) — for tmux -CC control mode (native tabs per agent)
- [Claude Code](https://claude.com/claude-code) — `npm install -g @anthropic-ai/claude-code`
- [GitHub CLI](https://cli.github.com/) — `brew install gh`
- [Go](https://go.dev/) — `brew install go` (for building from source)

## Install

```bash
# From source
git clone https://github.com/drdanmaggs/rocket-fuel.git
cd rocket-fuel
make install

# Verify
rocket-fuel version
```

## Quick start

### 1. Set up a GitHub Project

Create a [GitHub Project](https://docs.github.com/en/issues/planning-and-tracking-with-projects) for your repo with these columns:

**Someday/Maybe** → **Backlog** → **Scoped** → **In Progress** → **Review** → **Done**

### 2. Link your project

```bash
cd your-repo
rocket-fuel project link https://github.com/users/yourname/projects/1
```

### 3. Start the session

```bash
rocket-fuel up
```

This creates a tmux session with **Integrator** and **Dashboard** tabs, then attaches with tmux -CC. iTerm2 renders each as a native tab.

### 4. Spawn a worker

Label an issue with a workflow (e.g., `workflow:tdd`) and spawn a worker:

```bash
rocket-fuel work 42
# or
rocket-fuel work https://github.com/yourname/repo/issues/42
```

This creates a git worktree, opens a new iTerm2 tab, and launches Claude Code with the issue context and appropriate skill.

### 5. Monitor

```bash
# One-shot status
rocket-fuel status

# Live dashboard (in the dashboard tab)
watch -n 30 rocket-fuel status

# View the project board
rocket-fuel project board
```

### 6. Surface a tab

The Integrator brings tabs to the foreground when the Visionary's attention is needed:

```bash
rocket-fuel surface worker-42 "CI failing — needs your eyes"
```

This switches the active iTerm2 tab and sends a macOS notification.

### 7. Clean up

```bash
# Remove completed workers (worktrees + windows)
rocket-fuel reap

# Tear down the entire session
rocket-fuel down
```

## Skill routing

Issues are routed to skills based on labels:

| Label | Skill | Approach |
|-------|-------|----------|
| `workflow:tdd` | `/tdd` | RED → GREEN → REFACTOR |
| `workflow:bug-fix` | `/bug-fix` | Failing test first |
| `workflow:epc` | `/epc` | Explore → Plan → Code |
| `workflow:issue-scope` | `/issue-scope` | Break down into sub-issues |
| *(no label)* | `/epc` | Default |

## Development

```bash
make setup          # Configure git hooks
make build          # Build binary
make test           # Run tests (with race detector)
make test-integration  # Run integration tests (needs tmux)
make lint           # golangci-lint
make fmt            # Format code
make all            # fmt + lint + test + build
```

## Prior art

Inspired by [gastown](https://github.com/steveyegge/gastown) (Steve Yegge's multi-agent workspace manager) but philosophically opposite — compose existing tools rather than building a platform. See [`docs/gastown-lessons.md`](docs/gastown-lessons.md).
