# Rocket Fuel

> "Most entrepreneurial companies are missing one of two things: a Visionary or an Integrator." — Gino Wickman & Mark C. Winters, *Rocket Fuel*

## What is Rocket Fuel?

A multi-agent orchestrator that multiplies developer productivity by composing existing tools — Claude Code, GitHub, tmux-CC/iTerm2, and git worktrees — rather than building a platform.

Inspired by [gastown](https://github.com/steveyegge/gastown) (Steve Yegge's multi-agent workspace manager), but philosophically opposite: **compose, don't build.** Where gastown is 371K lines of Go with custom issue tracking (beads), a custom database (Dolt), custom daemons, and custom merge queues — Rocket Fuel is thin glue over tools you already use.

## The Model: Visionary + Integrator

From *Rocket Fuel* by Gino Wickman & Mark C. Winters. Every high-performing company has two leaders:

### The Visionary (You)

The ideas person. Sets direction, thinks about product, users, architecture. Lives in their main Claude Code tab — collaborating with a Visionary agent on scoping, strategy, and creative problem-solving.

But the Visionary isn't just a delegator. Like Elon on the factory floor, the Visionary can jump into any worker tab and get hands-on with the engineering. The Visionary has full access to every tab and can take over any workstream.

### The Integrator (Autonomous AI Agent)

The person who gets shit done. Runs autonomously — you don't interact with it. It manages the GitHub Project board, spawns workers, monitors CI, tracks progress against milestones, and keeps the machine running.

The Integrator's only communication to the Visionary is **surfacing a tab** — bringing an iTerm2 tab to the foreground when something needs the Visionary's attention. No chat interface. No status reports to read. Just "this needs you" → tab appears → context is there.

### Workers (Ephemeral AI Agents)

Claude Code instances running in isolated git worktrees. Each worker picks up a GitHub issue and executes it using existing skills (`/tdd`, `/bug-fix`, `/epc`, `/issue-scope`, etc.). When done, they create a PR and the Integrator assigns the next issue.

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│  iTerm2 (tmux -CC = native tabs per agent)               │
├───────────────┬──────────────┬───────────┬───────────────┤
│  Visionary    │  Integrator  │ Worker α  │  Worker β     │
│  (you live    │  (autonomous │ (spawned  │  (surfaced    │
│   here)       │   — no       │  by intg) │   for you     │
│               │   interaction│           │   when needed)│
└───────────────┴──────┬───────┴───────────┴───────────────┘
                       │
            ┌──────────┴──────────┐
            │  GitHub Projects    │
            │  (Integrator's      │
            │   brain — single    │
            │   source of truth)  │
            └─────────────────────┘
```

### Why tmux -CC?

tmux control mode (`-CC`) maps tmux sessions/windows to native iTerm2 tabs. This means:

- Each agent gets a native iTerm2 tab (not a tiny terminal pane)
- Full iTerm2 features: scrollback, search, CMD+click, inline images
- **Programmatic tab switching:** `tmux select-window -t rocket-fuel:worker-alpha` brings that tab to the foreground in iTerm2. This is how the Integrator surfaces work for the Visionary.
- macOS notifications via `osascript` for when iTerm2 isn't focused

### Why GitHub Projects?

The Integrator's state is a GitHub Project board. No hidden state. Everything visible in the GitHub web UI too.

- **Columns:** Backlog → Scoped → In Progress → Review → Done
- **Labels:** Skill routing (`workflow:tdd`, `workflow:bug-fix`, `workflow:epc`)
- **Milestones:** Time-boxing and velocity tracking
- **Automation:** Issue closed → card moves to Done

### Why git worktrees?

Each worker gets an isolated copy of the repo via `git worktree add`. This means:

- O(seconds) to spawn (shared object database, no full clone)
- Full isolation (workers can't step on each other)
- Clean branch per issue
- Auto-cleanup when work is done

### Why existing skills?

Skills (`/tdd`, `/issue-scope`, `/epc`, `/bug-fix`, `/ship`, `/pr-quality`) are battle-tested playbooks. Workers don't need custom logic — they run the right skill for the job. The Integrator routes issues to skills based on labels.

## What Rocket Fuel is NOT

- **Not a platform.** It's ~1000 lines of glue, not 371K lines of infrastructure.
- **Not gastown-lite.** Different philosophy entirely. No beads, no Dolt, no daemons, no custom merge queues, no federation.
- **Not a task runner.** The Visionary/Integrator dynamic creates natural tension between ideas and execution.
- **Not a web app.** It's a CLI + tmux. The "UI" is iTerm2.

## Prior Art

- [gastown](https://github.com/steveyegge/gastown) — Steve Yegge's multi-agent workspace manager. 371K lines of Go. Impressive but complex. Reference implementation lives at `~/gastown`.
- *Rocket Fuel* by Gino Wickman & Mark C. Winters — the Visionary/Integrator framework.

## Tech

- **Language:** Go (single binary, goroutines for worker monitoring, proven by gastown)
- **CLI framework:** Cobra
- **Dependencies:** tmux, iTerm2, Claude Code, gh CLI, git
- **Distribution:** Homebrew tap, GitHub releases
