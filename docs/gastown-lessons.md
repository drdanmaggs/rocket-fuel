# Lessons from Gastown

Reference implementation: `~/gastown` (Steve Yegge's multi-agent workspace manager, ~371K lines of Go).

## What we steal (concepts, not code)

### Git worktrees for worker isolation
Gastown's "polecats" use `git worktree add` — fast spawn (shared object DB), full isolation, clean branch per task, auto-cleanup. We do the same.

### Persistent identity for agents
Polecats have names and work history that survive session restarts. Our Integrator should maintain continuity across sessions — the GitHub Project board IS its persistent state.

### GUPP: "If there is work on your hook, you MUST run it"
Simple autonomous behaviour principle. Workers don't wait for instructions — if assigned an issue, they execute. The Integrator doesn't wait for the Visionary — if there are scoped issues and available capacity, it spawns workers.

### Context priming at session start
Gastown injects role + current state when an agent session starts (`gt prime`). Our agents need the same — the Integrator prompt includes current project board state, the worker prompt includes issue context + skill routing.

### Capacity control
Gastown's scheduler prevents spawning too many agents simultaneously (API rate limits, machine resources). We'll need something similar when running many concurrent workers.

## What we skip (and why)

### Beads + Dolt (custom issue tracking + database)
**Skip.** GitHub Issues + GitHub Projects already does this. We don't need a custom issue tracker — we need to use the one that already has a UI, API, notifications, and integrations.

### Custom daemon (14-step heartbeat loop)
**Skip.** The Integrator IS the daemon. It runs as a Claude Code session with `/loop` or similar polling. No separate Go process needed.

### Custom merge queue (batch-then-bisect)
**Skip.** GitHub has native merge queues. Our workers create PRs; GitHub handles merging.

### Mail / Nudge / Message routing
**Skip.** We use `tmux select-window` (tab surfacing) + GitHub issue comments. No custom inter-agent messaging protocol.

### Federation (cross-workspace, DoltHub sync)
**Skip.** Premature. This is a personal productivity tool, not a distributed system.

### Plugin system (dogs, wisps, scheduled plugins)
**Skip.** Claude Code skills ARE the plugins. No need for another abstraction layer.

### Custom TUI / Dashboard framework
**Skip.** iTerm2 via tmux-CC IS the UI. The dashboard is `watch` + `gh` commands.

### Witness / Refinery / Deacon (specialised agents)
**Skip.** Gastown has 6+ agent types with specific roles. We have two: the Integrator and Workers. The Integrator handles monitoring, health checks, and dispatching — no need for separate agents.

## The philosophical difference

Gastown builds a platform. Custom everything, deep infrastructure, 371K lines.

Rocket Fuel composes existing tools. GitHub, tmux, Claude Code, git. ~1000 lines of glue.

Both are valid approaches. Gastown scales to 20-30 agents across multiple repos with federation. Rocket Fuel is a personal force multiplier that leverages tools you already know.
