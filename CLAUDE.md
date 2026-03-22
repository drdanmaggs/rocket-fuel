# Rocket Fuel

Multi-agent orchestrator: Visionary/Integrator model over tmux-CC, GitHub Projects, Claude Code skills, and git worktrees.

## Vision

Read `docs/vision.md` for the full concept, `docs/gastown-lessons.md` for prior art analysis.

- **Visionary** = the human. Owns product direction (what/why). Scopes vague issues. Approves merges. Calls meetings.
- **Integrator** = AI agent (Claude Code, runs continuously). Owns execution (how/when). Manages board, dispatches workers, reviews PRs. Never scopes issues — surfaces vague ones to the Visionary. Can call meetings when it needs the Visionary's input.
- **Watchdog** = Go daemon (background). Keeps agents alive, detects stuck workers, reaps completed ones. No decisions — purely mechanical. Event-driven via Claude Code hooks.
- **Workers** = ephemeral Claude Code instances in git worktrees. Run skills (`/tdd`, `/bug-fix`, `/epc`) on assigned issues. Fully autonomous (GUPP).
- **Mission Control** = Stream Deck plugin (separate repo: `drdanmaggs/mission-control`). Physical dashboard for the Visionary.

See `docs/adr/002-agent-roles.md` for full role definitions and boundaries.

## This project is vibe coded

The Visionary (human) does not read or write code. All technical decisions — language, libraries, architecture, patterns — are made autonomously by the agents working on this project. The human provides direction, not implementation detail.

This means:
- Make sound technical decisions without asking for approval on implementation details
- Choose the right library/pattern/approach and move forward
- Document decisions in commit messages and ADRs when significant
- The code must be correct, tested, and well-structured — the human trusts you to get it right

## Architecture: Hybrid Plugin Model

See `docs/adr/006-hybrid-plugin-architecture.md` for full rationale.

| Component | Location | Purpose |
|-----------|----------|---------|
| Agent definitions | `internal/plugin/agents/` (17 agents) | Integrator, Worker, TDD subagents, code-reviewer subagents, etc. |
| Skills | `internal/plugin/skills/` (26 skills) | /tdd, /ship, /code-reviewer, /issue-scope, board-setup, etc. |
| Rules | `internal/plugin/rules/` (8 rules) | testing, commit-discipline, code-quality, etc. |
| Hooks | `.claude/settings.json` per repo | Project-scoped lifecycle hooks (installed by `rf launch`). See `docs/adr/001-claude-code-hooks.md` |
| Hook handlers | `cmd/*.go` | Role-aware via `hookutil.DetectRole()` — different behavior for Integrator vs Worker |
| Orchestration | `cmd/`, `internal/` | tmux, worktrees, watchdog, dashboard |

Plugin files are embedded via `go:embed` and extracted on every `rf launch`. Always-overwrite — fork to customize.

## Stack

- **Language:** Go
- **CLI:** Cobra
- **Runtime deps:** tmux, iTerm2 (tmux -CC), Claude Code (`claude`), gh CLI, git

## Reference

- gastown (prior art): `~/.claude/gastown` — explore for architectural patterns but do NOT copy its complexity. See `docs/gastown-lessons.md` for what to steal vs skip.
- family-meal-planner-v3: `~/family-meal-planner-v3` — reference for mature CI, testing, git hygiene patterns (translated from TypeScript to Go).
- **mission-control**: `~/mission-control` (repo: `drdanmaggs/mission-control`) — Stream Deck plugin for Rocket Fuel. Physical mission control dashboard. Separate repo (Node.js, Elgato SDK). Rocket Fuel exposes state via `rf streamdeck serve`, mission-control renders it on Stream Deck buttons.

## Development

- TDD always
- Conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`
- Structural and behavioral commits separated
- CI must pass before merge — lint, test, build gated on every PR
- Pre-commit hooks: gitleaks + gofmt + golangci-lint

## Testing Anti-Patterns (BANNED)

These were mistakes we made early and corrected. Do not repeat them.

### Never fake system tools this project depends on

Rocket Fuel lives and dies by tmux, git, and the GitHub API. Testing against recorders or stubs that don't prove commands actually work is dangerous — you ship code that "passes tests" but breaks in reality.

**Banned:**
```go
// Fake recorder that proves nothing
fake := NewFakeTmux()
fake.Record("new-session", "-s", "test")
assert(fake.HasCommand("new-session")) // So what? Did tmux actually create the session?
```

**Required:** Test real tools, isolate via configuration:
```go
// Real tmux with isolated socket (can't interfere with user's tmux)
tm := testutil.NewRealTmux(t)  // uses -L rf-test-<PID> socket
tm.NewSession(t, "test")       // real tmux session
assert(tm.HasSession(t, "test")) // actually checks tmux
```

```go
// Real git in temp dir (automatic cleanup)
repoDir := testutil.InitTestRepo(t)  // real git init + commit in t.TempDir()
_, wtDir := testutil.InitTestRepoWithWorktree(t, "worker-1")  // real worktree
```

**The rule:** Isolate via sockets, temp dirs, and httptest servers — not by replacing the tool with a fake that doesn't exercise the real code path.

### Never run Cobra Execute() in parallel tests

Cobra's `rootCmd` is a package-level singleton. `SetArgs`/`SetOut`/`Execute` mutate it. Parallel tests cause data races.

```go
// BANNED — data race on rootCmd
func TestFoo(t *testing.T) {
    t.Parallel()
    rootCmd.SetArgs([]string{"version"})
    rootCmd.Execute()
}
```

```go
// CORRECT — sequential, reset after
func TestFoo(t *testing.T) {
    // Not parallel — mutates rootCmd.
    rootCmd.SetOut(buf)
    rootCmd.SetArgs([]string{"version"})
    rootCmd.Execute()
    rootCmd.SetOut(nil)
    rootCmd.SetArgs(nil)
}
```

### RecordingTmux is for orchestration logic only

`RecordingTmux` exists for testing "what commands WOULD be issued" in orchestration logic (e.g., does the Integrator issue the right sequence of tmux commands). It does NOT replace integration tests against real tmux.

## Testing Strategy

### What to test where

| Layer | Type | Tag | What |
|-------|------|-----|------|
| Pure logic | Unit | (none) | Routing, prompt building, dashboard rendering, queue ops |
| tmux operations | Integration | `//go:build integration` | Session/window/pane creation with real tmux |
| git operations | Unit | (none) | Real git in t.TempDir() — no tag needed |
| GitHub API | Unit | (none) | GHRunner mocks — inject test doubles |
| Claude Code | Manual | — | Can't automate Claude in CI |

### Test commands

```bash
go test -race ./...                              # unit tests only
go test -race -tags=integration ./...            # unit + integration
make test                                         # unit (CI fast path)
make test-integration                            # integration (CI, installs tmux)
```

### Integration test patterns

All integration tests MUST:
- Use `//go:build integration` tag
- Use isolated tmux socket (`testutil.SetupTmuxSocket()` in TestMain)
- Use `t.Cleanup` for session teardown
- NOT use `t.Parallel()` (tmux socket is shared per process)
- Use retry loops for CI (SendKeys needs time to process on slow runners)

### Mock runners must implement full interface

When adding a method to `tmux.Runner`, update ALL mock runners:
- `internal/session/session_test.go`
- `internal/status/status_test.go`
- `internal/worker/reap_test.go`
- `internal/notify/notify_test.go`

### Lessons learned

- `filepath.Walk` silently fails when hidden-dir skipping interacts with traversal — use `ReadDir` for explicit control
- `SelectWindow` has side effects (switches active tab) — use `HasWindow` for existence checks
- Mock runners hide real-world side effects — always pair unit tests with integration tests for tmux operations
- CI runners are slower — use retry loops for timing-dependent assertions (e.g., SendKeys → capture-pane)
- Always run with `-race` flag
