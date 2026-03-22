# ADR-001: Claude Code Hooks Reference

## Status: Active

## Context

Rocket Fuel uses Claude Code hooks to make the agent system event-driven rather than polling-based. Hooks fire at specific lifecycle events and call `rf` commands that adapt behavior based on the agent's role (Integrator vs Worker).

## Hook Input

Claude Code passes JSON to hook handlers via stdin. Common fields:

```json
{
  "session_id": "abc123",
  "cwd": "/path/to/repo",
  "hook_event_name": "SessionStart",
  "agent_type": "integrator"
}
```

The `agent_type` field is set when Claude is launched with `--agent`. This is the primary role detection mechanism. Fallback: check if `cwd` contains `.worktrees/` (Worker) or not (Integrator).

See: https://code.claude.com/docs/en/hooks

## Role-Specific Hook Matrix

All 7 hooks fire for both Integrator and Worker sessions (shared `.claude/settings.json`). Role differentiation happens inside each handler via `hookutil.DetectRole()`.

| Hook | Handler | Integrator behavior | Worker behavior |
|------|---------|-------------------|----------------|
| SessionStart | `rf prime` | Full context (board, workers, repo) | Repo context only (no board/workers) |
| PreCompact | `rf prime` | Full context (re-inject after compaction) | Repo context only |
| Stop | `rf should-continue` | Block if work exists on board | **No-op** (allow stop — Workers self-exit) |
| StopFailure | `rf handle-stop-failure` | Log error to activity feed | Log error to activity feed |
| PostToolUse | `rf record-activity` | **No-op** (not tracked for stuck detection) | Record activity timestamp |
| SessionEnd | `rf session-ended` | Log warning (unexpected death) | Reap worker, nudge Integrator |
| PreToolUse | `rf check-merge-safety` | Gate PR merges (CI green, not draft) | Gate PR merges (same) |

## Handler Types

- `command` — shell script (stdin/stdout/exit code). All Rocket Fuel hooks use this.
- `http` — POST to HTTP endpoint
- `prompt` — single-turn Claude evaluation
- `agent` — subagent with tool access

## Exit Code Semantics

- **0**: success (stdout may contain JSON response)
- **2**: block action (stderr = reason shown to Claude)
- **Other**: non-blocking error (logged, execution continues)

## Current Settings

Created by `rf launch` via `launch.EnsureClaudeSettings()`:

```json
{
  "hooks": {
    "SessionStart": [{"matcher": "", "hooks": [{"type": "command", "command": "rf prime"}]}],
    "PreCompact": [{"matcher": "", "hooks": [{"type": "command", "command": "rf prime"}]}],
    "Stop": [{"matcher": "", "hooks": [{"type": "command", "command": "rf should-continue"}]}],
    "StopFailure": [{"matcher": "", "hooks": [{"type": "command", "command": "rf handle-stop-failure"}]}],
    "PostToolUse": [{"matcher": "", "hooks": [{"type": "command", "command": "rf record-activity"}]}],
    "SessionEnd": [{"matcher": "", "hooks": [{"type": "command", "command": "rf session-ended"}]}],
    "PreToolUse": [{"matcher": "Bash(gh pr merge*)", "hooks": [{"type": "command", "command": "rf check-merge-safety"}]}]
  }
}
```

## Why Hooks Stay Project-Scoped

Hooks are written to `.claude/settings.json` per repo (not in the global plugin). The Stop hook would block Claude from stopping in ALL sessions if applied globally. See ADR-006 for the plugin architecture decision.
