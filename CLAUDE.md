# Rocket Fuel

Multi-agent orchestrator: Visionary/Integrator model over tmux-CC, GitHub Projects, Claude Code skills, and git worktrees.

## Vision

Read `docs/vision.md` for the full concept, `docs/gastown-lessons.md` for prior art analysis.

- **Visionary** = the human. Talks directly to the Integrator. Sets direction (what/why). Can jump into worker tabs to get hands-on.
- **Integrator** = AI agent, the human's main interface. Manages GitHub Project board, spawns workers, monitors CI. Owns execution (how/when). Protects progress from the Visionary's destabilising energy. Never says no — scopes ideas, parks them in Someday/Maybe, redirects to current work.
- **Workers** = ephemeral Claude Code instances in git worktrees. Run skills (`/tdd`, `/bug-fix`, `/epc`) on assigned issues.

## This project is vibe coded

The Visionary (human) does not read or write code. All technical decisions — language, libraries, architecture, patterns — are made autonomously by the agents working on this project. The human provides direction, not implementation detail.

This means:
- Make sound technical decisions without asking for approval on implementation details
- Choose the right library/pattern/approach and move forward
- Document decisions in commit messages and ADRs when significant
- The code must be correct, tested, and well-structured — the human trusts you to get it right

## Stack

- **Language:** Go
- **CLI:** Cobra
- **Runtime deps:** tmux, iTerm2 (tmux -CC), Claude Code (`claude`), gh CLI, git

## Reference

- gastown (prior art): `~/gastown` — explore for architectural patterns but do NOT copy its complexity. See `docs/gastown-lessons.md` for what to steal vs skip.
- family-meal-planner-v3: `~/Websites/family-meal-planner-v3` — reference for mature CI, testing, git hygiene patterns (translated from TypeScript to Go).

## Development

- TDD always
- Conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`
- Structural and behavioral commits separated
- CI must pass before merge — lint, test, build gated on every PR
- Pre-commit hooks: gitleaks + gofmt + golangci-lint
