# TDD Plan: Full Integrator Session (v0.2)

## Context
v0.1 shipped the CLI primitives (up/down/work/reap/status/project). But `rocket-fuel up` creates empty tmux windows — no Claude session, no automation. For basic functionality, the Integrator needs to: (1) launch as a Claude Code session the Visionary can talk to, (2) autonomously dispatch workers from the board, and (3) keep the board state in sync.

## Architecture
One Integrator Claude session handles both conversation and dispatch (shared context = smarter decisions). A Go heartbeat (`rocket-fuel heartbeat`) runs in a hidden window, periodically calling dispatch + reap. The Scoped column is the handoff point: the Integrator (AI) triages ideas onto the board; the heartbeat (Go) picks up Scoped items and spawns workers.

Inspired by gastown: `rocket-fuel prime` is our `gt prime` — a universal context injector. The daemon/heartbeat is deliberately dumb (keeps agents alive, nudges them). All intelligence lives in the Claude session.

## Session Constants
Test command: `go test -race ./...`
Test file pattern: colocated `*_test.go` in each package
Test helpers: `internal/testutil/` (git.go, tmux.go, exec.go)

## Slice 1: `rocket-fuel prime` — context injector
Type: unit | Status: pending
Files: `internal/prime/prime.go`, `internal/prime/prime_test.go`, `cmd/prime.go`

The Integrator's "eyes." Gathers board state + active workers + repo info and outputs it as structured markdown. Like gastown's `gt prime`, this is the universal way agents get context.

- [ ] returns markdown with board state section when project is linked
- [ ] returns markdown with worker status section (active/done workers)
- [ ] returns markdown with repo context (current branch, repo dir)
- [ ] handles missing project config gracefully (no board linked yet)
- [ ] handles empty board (no items in any column)
- [ ] handles no workers (no .worktrees directory)
- [ ] includes integrator.md prompt content in output
- [ ] cmd/prime prints to stdout (piped to claude or read by human)

## Slice 2: Board state transitions — move cards between columns
Type: unit | Status: pending
Files: `internal/project/move.go`, `internal/project/move_test.go`
Builds on: existing project package

Workers spawn and complete, but the board doesn't update. This slice adds the ability to move items between columns using `gh project item-edit`.

- [ ] MoveItem builds correct `gh project item-edit` command with item ID, field ID, and value
- [ ] MoveItem returns error when gh command fails
- [ ] GetStatusFieldID fetches the Status field ID from the project (needed by gh API)
- [ ] GetStatusOptionID fetches the option ID for a status value (e.g., "In Progress")
- [ ] field/option IDs are cached after first fetch (avoid repeated API calls)

## Slice 3: `rocket-fuel up` launches Claude in Integrator tab
Type: unit | Status: pending
Files: `cmd/up.go`, `internal/session/session.go`, `internal/session/session_test.go`
Builds on: Slice 1

Currently `up` creates windows but doesn't launch anything. This slice makes it launch Claude Code in the integrator window with the prime context as the initial prompt.

- [ ] Setup accepts an optional LaunchConfig with commands to send to each window
- [ ] on new session, sends `claude --prompt-file <path>` to the integrator window
- [ ] writes prime output to a temp file for --prompt-file (avoids shell escaping issues with long prompts)
- [ ] on existing session (reattach), does NOT relaunch Claude (session already running)
- [ ] --dry-run skips both Claude launch and tmux attach

## Slice 4: `rocket-fuel dispatch` — automated pickup from Scoped
Type: unit | Status: pending
Files: `internal/dispatch/dispatch.go`, `internal/dispatch/dispatch_test.go`, `cmd/dispatch.go`
Builds on: Slice 2

The mechanical part: check board for Scoped items, check capacity, spawn a worker, move the card. Single invocation — no loop.

- [ ] dispatches first Scoped item when capacity is available
- [ ] moves item from Scoped to In Progress on successful spawn
- [ ] skips dispatch when no Scoped items exist (returns "nothing to dispatch")
- [ ] skips dispatch when at max capacity (returns "at capacity: N/N workers")
- [ ] max capacity is configurable (default 3, stored in .rocket-fuel/config.json)
- [ ] returns dispatch result with issue number, worker name, action taken
- [ ] extends `rocket-fuel reap` to move reaped items to Done when PR exists

## Slice 5: `rocket-fuel heartbeat` — periodic dispatch + reap
Type: unit | Status: pending
Files: `internal/heartbeat/heartbeat.go`, `internal/heartbeat/heartbeat_test.go`, `cmd/heartbeat.go`
Builds on: Slice 4

The dumb, reliable background loop. Runs dispatch + reap on a ticker. Like gastown's daemon but much simpler — no decision-making, just periodic execution.

- [ ] runs one cycle: dispatch then reap, returns combined results
- [ ] --loop flag runs continuously with configurable --interval (default 3m)
- [ ] logs each cycle's results to stdout (visible in tmux scrollback)
- [ ] handles errors gracefully (logs and continues, doesn't crash the loop)
- [ ] --dry-run shows what would be dispatched/reaped without acting

## Slice 6: Wire it together in `rocket-fuel up`
Type: integration | Status: pending
Files: `cmd/up.go`, `internal/session/session.go`
Builds on: Slices 3, 5

The final assembly. `rocket-fuel up` creates three windows: Integrator (Claude), heartbeat (hidden), dashboard (status). The system is fully operational after `up`.

- [ ] session.Windows includes "heartbeat" as a window
- [ ] heartbeat window launches `rocket-fuel heartbeat --loop` via SendKeys
- [ ] dashboard window launches `rocket-fuel status` (or watch mode if added)
- [ ] integrator window is selected by default (user sees it first)
- [ ] on reattach (session exists), no windows are relaunched
- [ ] --dry-run shows what would be launched in each window
