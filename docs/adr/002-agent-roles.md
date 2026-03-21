# ADR-002: Agent Roles and Responsibilities

## Status: Active

## Context

Rocket Fuel implements the Visionary/Integrator model from *Rocket Fuel* by Gino Wickman & Mark C. Winters, extended with a Watchdog daemon and Worker agents. Each role has clear boundaries.

## Roles

### Visionary (Human)

**What they own:** Product direction. What to build and why.

**What they do:**
- Set direction, priorities, and vision
- Scope vague issues (product judgment the AI cannot provide)
- Approve merges (via the approval queue)
- Jump into worker tabs when they want hands-on
- Call meetings with the Integrator to brainstorm

**What they don't do:**
- Write code (the project is vibe-coded)
- Manage the board (Integrator does that)
- Monitor workers (Watchdog does that)

**Interface:** Meeting Room session (rf-meeting). Can also read any tab.

---

### Integrator (Claude Code session, runs continuously)

**What they own:** Execution. How to build and when.

**What they do:**
- Manage the GitHub Project board (move cards, close epics, update status)
- Dispatch workers on scoped issues
- Review completed work (PRs, CI status)
- Surface items needing the Visionary's input
- Make judgment calls on priorities and sequencing
- Respond to Watchdog nudges (worker completed, worker stuck)

**What they don't do:**
- Scope vague issues (that's the Visionary's product judgment)
- Dispatch workers on unscoped issues (surface to Visionary instead)
- Generate product ideas unprompted
- Stop working (kept alive by Stop hook / Watchdog nudges)

**Interface:** rf-integrator tmux session. Receives nudges from Watchdog.

**Decision tree before dispatch:**

| Condition | Action |
|-----------|--------|
| Clear scope + acceptance criteria | Dispatch worker |
| Vague / one-liner / brainstorm | Surface to Visionary |
| All sub-issues closed | Close the epic |
| Needs product decision | Surface to Visionary |
| Already has a PR | Check PR status |

---

### Watchdog (Go daemon, background process)

**What they own:** System health. Keeping everything alive and running.

**What they do:**
- Keep the Integrator alive (restart on crash, Stop hook forces continuation)
- Detect stuck workers (no git activity for 20 min)
- Reap completed workers (window gone → cleanup worktree)
- Nudge the Integrator with events (worker done, worker stuck, new issues)
- Track worker activity via PostToolUse hook

**What they don't do:**
- Dispatch workers (Integrator decides)
- Manage the board (Integrator decides)
- Make any judgment calls (purely mechanical)

**Interface:** rf-watchdog tmux session (formerly rf-mission-control).

---

### Workers (Claude Code sessions, ephemeral)

**What they own:** A single issue. Execute it autonomously.

**What they do:**
- Read the issue (via `gh issue view`)
- Execute the assigned skill (/tdd, /epc, /bug-fix)
- Write tests first (TDD is the default)
- Create a PR
- Exit when done

**What they don't do:**
- Scope issues (if the issue is vague, that's a dispatch problem)
- Talk to the Visionary directly (go through the Integrator)
- Modify other workers' code
- Stop and ask for permission (GUPP)
- Take shortcuts (removing broken features instead of fixing them)

**Interface:** rf-integrator session, individual window per worker (#N: title).

---

### Ground Control (Stream Deck plugin, separate repo)

**What they own:** Physical dashboard and tactile interface.

**What they do:**
- Display worker status on physical buttons (idle/working/PR ready/stuck)
- Provide quick actions (launch, land, dispatch, call meeting)
- Flash amber when the Integrator needs the Visionary's attention
- Show board pipeline state at a glance

**What they don't do:**
- Make decisions (pure display + input routing)

**Interface:** Elgato Stream Deck → WebSocket → `rf streamdeck serve`
**Repo:** github.com/drdanmaggs/ground-control

---

## Communication

The GitHub Project board is the shared state. No direct agent-to-agent communication needed.

```
Visionary ←→ Meeting Room → creates issues → Board
                                                ↓
Watchdog → nudges → Integrator → reads board → dispatches Workers
    ↑                                              ↓
    └──── detects completion/stuck ←── Workers create PRs
```

## tmux Sessions

| Session | Process | Visible |
|---------|---------|---------|
| rf-integrator | Claude Code (Integrator) + Worker windows | Yes — Visionary's main view |
| rf-watchdog | `rf watchdog --loop` | Background (viewable in tmux dashboard) |
| rf-meeting | Claude Code (on demand) | When meeting is active |
