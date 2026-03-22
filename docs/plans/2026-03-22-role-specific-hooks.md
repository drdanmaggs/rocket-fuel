# TDD Plan: Role-specific hook configs

## Context
All 7 hooks fire identically for Integrator and Workers. The Stop hook is the most critical bug — it prevents Workers from exiting after completing their work, forcing them to continue indefinitely. Each handler needs to detect its role and behave accordingly.

## Architecture
Shared role detection utility (`internal/hookutil`) parses the Claude Code hook input JSON for `agent_type` field (set when using `--agent`). Falls back to cwd-based detection (`.worktrees/` = Worker). All hook handlers use this utility to branch behavior per role.

## Session Constants
Test command: `go test -race ./...`
Test file pattern: colocated `*_test.go`
Test helpers: `internal/testutil/`
Acceptance test path: none

## Slice 1: Role detection utility
Type: unit | Status: done
Files: `internal/hookutil/role.go` (new), `internal/hookutil/role_test.go` (new)

- [x] DetectRole returns Integrator when agent_type is "integrator"
- [x] DetectRole returns Worker when agent_type is "worker"
- [x] DetectRole returns Worker when cwd contains ".worktrees/" (fallback)
- [x] DetectRole returns Integrator when agent_type is empty and cwd is repo root (fallback)

## Slice 2: Stop hook — no-op for Workers
Type: unit | Status: done
Files: `cmd/shouldcontinue.go`
Builds on: Slice 1

- [x] rf should-continue exits 0 (allow stop) when role is Worker
- [x] rf should-continue still blocks when role is Integrator and work exists

## Slice 3: Prime hook — simplified for Workers
Type: unit | Status: done
Files: `cmd/prime.go`, `internal/prime/prime.go`
Builds on: Slice 1

- [x] rf prime outputs only repo context for Workers (no board state, no worker status)
- [x] rf prime still outputs full context for Integrator

## Slice 4: Record-activity — explicit role check
Type: unit | Status: done
Files: `cmd/recordactivity.go`
Builds on: Slice 1

- [x] rf record-activity only records for Workers (explicit check, not implicit cwd)
- [x] rf record-activity is no-op for Integrator

## Slice 5: Session-ended — per-role behavior
Type: unit | Status: done
Files: `cmd/sessionended.go`
Builds on: Slice 1

- [x] rf session-ended reaps worker and nudges Integrator when role is Worker
- [x] rf session-ended logs warning when role is Integrator (unexpected death)

## Slice 6: Remaining hooks + documentation
Type: unit + docs | Status: done
Files: `docs/adr/001-claude-code-hooks.md`, `CLAUDE.md`
Builds on: Slice 1

- [x] rf handle-stop-failure — skipped: already reads stdin for error data, works for both roles
- [x] rf check-merge-safety — skipped: already reads stdin for tool input, works for both roles
- [x] ADR-001 updated with role-specific hook matrix and detection strategy
- [x] CLAUDE.md references hook documentation
