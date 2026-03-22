# ADR-008: Session Cycling for Workers

## Status: Active

## Context

Workers on complex issues (exploration, TDD cycles, code review) burn through Claude's context window. When compaction happens, the conversation is lossy-compressed — the Worker loses track of what it's already done and may redo work or lose plan progress.

Gastown solved this with session cycling: instead of compacting, kill the Claude session and restart fresh with re-injected context. The key insight is that git commits preserve code state, and the plan file (docs/plans/) persists on disk. Only the conversation context is ephemeral.

## Decision

Workers get **session cycling** instead of compaction. The `rf precompact` command handles the PreCompact hook for both roles:

- **Integrator**: Re-primes context (board, workers, repo) — same as before
- **Worker**: Spawns a background process that kills Claude (Ctrl-C via tmux) and restarts fresh in the same window

### Cycling Flow

1. PreCompact hook fires → `rf precompact` runs
2. `hookutil.DetectRole()` identifies Worker from `agent_type` or `.worktrees/` in cwd
3. `worker.CycleWorker()` extracts issue number from worktree path
4. Background process spawns (detached, same pattern as `spawnDashboardSplit`)
5. Hook returns immediately (Claude continues briefly)
6. Background: sleep 2 → Ctrl-C kills Claude → sleep 1 → sends new `claude --agent worker` command
7. SessionStart hook fires → `rf prime` re-injects repo context
8. Worker continues from git state

### What State Survives

- **Git commits** — all code changes committed during TDD cycles
- **Worktree** — same branch, same working directory
- **Plan file** — `docs/plans/*.md` persists on disk with checked/unchecked items
- **Issue context** — re-fetched via `gh issue view` on restart

### What State Is Lost

- **Conversation history** — the current chat context (that's the point)
- **In-flight reasoning** — any mid-thought analysis (mitigated by frequent TDD commits)

## Alternatives Considered

### Allow compaction (status quo)
Lossy compression preserves some context but degrades quality. Workers lose track of progress and redo work.

### Save progress summary to disk
Write a `.rocket-fuel/worker-state.md` before cycling. Adds complexity — git commits + plan file already provide sufficient state.

### Block compaction (exit code 2)
Return exit code 2 from PreCompact to prevent compaction. Buys time but eventually hits the hard context limit with no recovery.

## Consequences

- Workers get full context on every cycle instead of degraded compacted context
- Background process pattern is proven (same as dashboard split)
- Requires tmux for the kill/restart mechanism (already a hard dependency)
- Session cycling is **normal operation**, not failure — borrowed from gastown
- PreCompact hook now calls `rf precompact` instead of `rf prime`

## References

- gastown: `gt handoff --cycle` in `internal/cmd/handoff.go`
- Claude Code hooks: https://code.claude.com/docs/en/hooks.md
