# ADR-001: Claude Code Hooks Reference

## Status: Active

## Context

Rocket Fuel uses Claude Code hooks to make the agent system event-driven rather than polling-based. This document catalogs all available hooks and how Rocket Fuel uses or plans to use them.

## All Claude Code Hook Events

| Event | When it fires | Rocket Fuel usage |
|-------|--------------|-------------------|
| `SessionStart` | Session begins or resumes | **Active** — runs `rf prime` to inject board state + integrator prompt |
| `SessionEnd` | Session terminates | **Planned** — trigger watchdog cleanup on worker/integrator death |
| `UserPromptSubmit` | User submits a prompt | Unused |
| `Stop` | Claude finishes responding | **Planned** — keep Integrator alive (force continuation when work exists) |
| `StopFailure` | Turn ends due to API error | **Planned** — watchdog logs crashes, restarts agents |
| `PreToolUse` | Before a tool executes | Unused (could gate destructive commands) |
| `PostToolUse` | After a tool succeeds | **Planned** — track worker activity timestamps for stuck detection |
| `PostToolUseFailure` | After a tool fails | Unused |
| `PermissionRequest` | Permission dialog appears | Unused (--dangerously-skip-permissions bypasses) |
| `PreCompact` | Before context compaction | **Active** — re-injects `rf prime` context |
| `PostCompact` | After compaction completes | Unused |
| `SubagentStart` | Subagent spawned | Unused |
| `SubagentStop` | Subagent finishes | Unused |
| `TeammateIdle` | Agent team teammate about to go idle | Unused |
| `TaskCompleted` | Task marked as complete | Unused |
| `Notification` | Claude sends a notification | Unused |
| `InstructionsLoaded` | CLAUDE.md or rules file loaded | Unused |
| `ConfigChange` | Config file changes during session | Unused |
| `WorktreeCreate` | Worktree created | Unused (we manage worktrees ourselves) |
| `WorktreeRemove` | Worktree removed | Unused |
| `Elicitation` | MCP server requests user input | Unused |
| `ElicitationResult` | User responds to MCP elicitation | Unused |

## Hook handler types

All events support four handler types:
- `command` — shell script (stdin/stdout/exit code)
- `http` — POST to HTTP endpoint
- `prompt` — single-turn Claude evaluation
- `agent` — subagent with tool access

## Exit code semantics

- **0**: success (stdout may contain JSON response)
- **2**: block action (stderr = reason shown to Claude)
- **Other**: non-blocking error (logged, execution continues)

## Current .claude/settings.json hooks

Created by `rf launch` via `launch.EnsureClaudeSettings()`:

```json
{
  "hooks": {
    "SessionStart": [{"command": "rf prime"}],
    "PreCompact": [{"command": "rf prime"}]
  }
}
```

## Planned hooks (epic #170)

### Stop → keep Integrator alive
```json
{"Stop": [{"command": "rf watchdog should-continue"}]}
```
Returns `decision: "block"` when work exists, forcing Claude to continue.

### SessionEnd → instant worker cleanup
```json
{"SessionEnd": [{"command": "rf watchdog session-ended"}]}
```
Triggers immediate reap instead of waiting for 3-minute poll.

### PostToolUse → track worker activity
```json
{"PostToolUse": [{"command": "rf watchdog record-activity"}]}
```
Writes timestamp for stuck detection.
