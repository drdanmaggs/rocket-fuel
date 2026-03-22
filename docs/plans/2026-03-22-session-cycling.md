# TDD Plan: Session cycling on PreCompact

## Context
Workers on complex issues burn through context fast. When compaction happens, they lose context about what they've already done and may redo work. Session cycling preserves progress (via git commits) and restarts with a fresh context window that includes the issue + current branch state.

## Architecture
New `rf precompact` command replaces `rf prime` as the PreCompact hook. Detects role: Integrator gets re-primed (current behavior), Worker gets cycled (background process sends Ctrl-C to kill Claude, then restarts fresh in same tmux window). Worker state persists via git commits and worktree — no extra state file needed. Background process pattern matches existing spawnDashboardSplit.

## Session Constants
Test command: `go test -race ./...`
Test file pattern: colocated `*_test.go`
Test helpers: `internal/testutil/`
Acceptance test path: none

## Slice 1: rf precompact — Integrator re-primes
Type: unit | Status: done
Files: `cmd/precompact.go` (new), `cmd/precompact_test.go` (new)

- [x] runPrecompactWith returns prime context for Integrator role
- [x] runPrecompactWith does not cycle for Integrator role

## Slice 2: Extract worker restart command builder
Type: unit | Status: done
Files: `internal/worker/worker.go`, `internal/worker/worker_test.go`

- [x] BuildRestartCommand returns correct claude command given issue number
- [x] BuildRestartCommand fetches issue title and routes skill from labels

## Slice 3: Worker cycle — background kill + restart
Type: unit | Status: pending
Files: `cmd/precompact.go`, `internal/worker/cycle.go` (new), `internal/worker/cycle_test.go` (new)
Builds on: Slice 1 + 2

- [ ] runPrecompactWith spawns background process for Worker role (returns nil immediately)
- [ ] CycleWorker builds correct background script (Ctrl-C, sleep, restart command)
- [ ] CycleWorker extracts issue number from worktree cwd

## Slice 4: Update PreCompact hook + docs
Type: unit + docs | Status: pending
Files: `internal/launch/launch.go`, `internal/launch/launch_test.go`, `docs/adr/008-session-cycling.md` (new), `CLAUDE.md`
Builds on: Slice 3

- [ ] EnsureClaudeSettings writes rf precompact for PreCompact hook (not rf prime)
- [ ] ADR-008 documents session cycling decision (why cycle vs compact, gastown precedent)
- [ ] CLAUDE.md updated with session cycling in architecture section
