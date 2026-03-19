# Rocket Fuel

Multi-agent orchestrator: Visionary/Integrator model over tmux-CC, GitHub Projects, Claude Code skills, and git worktrees.

## Vision

Read `docs/vision.md` for the full concept. Key points:

- **Visionary** = the human's tab (+ AI agent). Sets direction, can jump into any worker tab.
- **Integrator** = autonomous AI agent. Manages GitHub Project board, spawns workers, monitors CI. Never interacts with the Visionary directly — surfaces tabs when attention is needed.
- **Workers** = ephemeral Claude Code instances in git worktrees. Run skills (`/tdd`, `/bug-fix`, `/epc`) on assigned issues.

## Stack

- **Language:** Go
- **CLI:** Cobra
- **Runtime deps:** tmux, iTerm2 (tmux -CC), Claude Code (`claude`), gh CLI, git

## Reference

- gastown (prior art): `~/gastown` — explore for architectural patterns but do NOT copy its complexity. Rocket Fuel composes existing tools; gastown builds a platform.

## Development

- TDD always
- Conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`
- Structural and behavioral commits separated (see global rules)
