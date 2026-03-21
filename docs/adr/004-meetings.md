# ADR-004: Meetings — Visionary/Integrator Communication

## Status: Active

## Context

The Integrator runs continuously and shouldn't be interrupted for conversations. The Visionary needs to brainstorm, scope ideas, and make product decisions. Meetings are the structured way these two interact.

## The meeting model

A meeting is a separate Claude Code session (rf-meeting) where the Visionary and an AI partner brainstorm, scope issues, and make product decisions. The output is always GitHub issues — which the Integrator picks up naturally from the board.

## Who calls meetings

### Visionary calls a meeting
- **Trigger**: Visionary has an idea, wants to scope work, needs to brainstorm
- **How**: Press the Meeting button on Stream Deck, or switch to meeting tab and run `rf meet`
- **What happens**: Fresh Claude session starts with `/issue-scope` and project context
- **Output**: GitHub issues created/updated → land on the board → Integrator dispatches

### Integrator calls a meeting
- **Trigger**: Vague issue found, product decision needed, unclear priority
- **How**: Integrator surfaces a request via the notification system
- **Notification cascade**:
  1. Stream Deck Meeting button pulses amber (if connected)
  2. macOS notification via `osascript`: "Integrator needs your input on #1146"
  3. tmux bell on the meeting tab (iTerm2 shows attention indicator)
  4. Dashboard pane shows pending meeting request
- **What happens**: Visionary sees the notification, opens meeting tab, addresses the Integrator's question
- **Output**: Clarified issue, updated scope, decision made → Integrator proceeds

## Key principle: the board is the handoff

Meetings produce GitHub issues. The Integrator consumes GitHub issues. No direct communication between the meeting session and the Integrator session. The board IS the shared state.

```
Meeting → creates/updates issues → Board → Integrator reads → dispatches
```

## Notification fallback chain

The Visionary might not always have the Stream Deck. Notifications cascade:

| Priority | Channel | When available |
|----------|---------|----------------|
| 1 | Stream Deck button pulse | At desk with Stream Deck |
| 2 | macOS notification center | Any Mac, even with iTerm2 in background |
| 3 | tmux bell / tab attention | iTerm2 open, any tab visible |
| 4 | Dashboard pane message | Integrator tab visible |

The system should try all available channels. At least one will reach the Visionary.

## Implementation

### rf meet command
- Creates or reattaches to the rf-meeting tmux session
- Launches Claude Code with project context + `/issue-scope` ready
- Session is ephemeral — ends when the Visionary is done

### rf surface (existing command, extended)
- Currently switches tmux windows and sends macOS notification
- Extend to: also trigger Stream Deck flash via `rf streamdeck serve`
- Extend to: also ring tmux bell on the meeting tab

### Integrator prompt
- When the Integrator encounters a vague issue or needs product direction:
  - Call `rf surface meeting "Need your input on #1146"`
  - This triggers the notification cascade
  - Add the request to the dashboard's pending items
  - Continue working on other things (don't block)
