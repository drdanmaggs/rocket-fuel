# Rocket Fuel

> "Most entrepreneurial companies are missing one of two things: a Visionary or an Integrator." — Gino Wickman & Mark C. Winters, *Rocket Fuel*

## What is Rocket Fuel?

A multi-agent orchestrator that multiplies developer productivity by composing existing tools — Claude Code, GitHub, tmux-CC/iTerm2, and git worktrees — rather than building a platform.

Inspired by [gastown](https://github.com/steveyegge/gastown) (Steve Yegge's multi-agent workspace manager), but philosophically opposite: **compose, don't build.** Where gastown is 371K lines of Go with custom issue tracking (beads), a custom database (Dolt), custom daemons, and custom merge queues — Rocket Fuel is thin glue over tools you already use.

## The Model: Visionary + Integrator

From *Rocket Fuel* by Gino Wickman & Mark C. Winters. Every high-performing company has two leaders:

### The Visionary (You — the human)

The ideas person. Sets direction, thinks about product, users, architecture. The Visionary is not an AI agent — it's you. You talk directly to the Integrator. That's your main interface.

But the Visionary isn't just a delegator. Like Elon on the factory floor, the Visionary can jump into any worker tab and get hands-on with the engineering. Full access to every tab, can take over any workstream.

The Visionary is also inherently destabilising. New ideas, pivots, "what if we..." — this energy is valuable but dangerous to in-flight work. The Integrator's job is to harness it without letting it derail progress.

### The Integrator (AI Agent — your main interface)

The person who gets shit done. You talk to the Integrator. It manages the GitHub Project board, spawns workers, monitors CI, tracks progress against milestones, and keeps the machine running.

**The Integrator never says no.** When the Visionary has an idea mid-sprint, the Integrator:
1. **Indulges it** — scopes it properly, creates a well-documented GitHub issue
2. **Parks it** — files it in the Someday/Maybe column on the project board
3. **Protects the current sprint** — "Great, it's captured. Now, back to the epic."

The idea never gets lost (which would frustrate the Visionary), but it never derails current work either. Both sides happy.

The Integrator also surfaces tabs when something needs the Visionary's hands — bringing an iTerm2 tab to the foreground with the context already there.

### Workers (Ephemeral AI Agents)

Claude Code instances running in isolated git worktrees. Each worker picks up a GitHub issue and executes it using existing skills (`/tdd`, `/bug-fix`, `/epc`, `/issue-scope`, etc.). When done, they create a PR and the Integrator assigns the next issue.

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│  iTerm2 (tmux -CC = native tabs per agent)               │
├───────────────┬───────────┬───────────┬──────────────────┤
│  Integrator   │ Worker α  │ Worker β  │  Dashboard       │
│  (you talk    │ (spawned  │ (spawned  │  (status)        │
│   to this)    │  by intg) │  by intg) │                  │
└───────┬───────┴───────────┴───────────┴──────────────────┘
        │
        │  Visionary (you) ←→ Integrator (AI)
        │  "What if we added X?"
        │  "Great idea. Scoped it, filed it in Someday/Maybe.
        │   Now — worker alpha just opened a PR on #42."
        │
        ▼
┌─────────────────────────────────────────┐
│  GitHub Projects                        │
│  (Integrator's brain)                   │
│                                         │
│  Someday/Maybe → Backlog → Scoped →     │
│  In Progress → Review → Done            │
└─────────────────────────────────────────┘
```

### Why tmux -CC?

tmux control mode (`-CC`) maps tmux sessions/windows to native iTerm2 tabs. This means:

- Each agent gets a native iTerm2 tab (not a tiny terminal pane)
- Full iTerm2 features: scrollback, search, CMD+click, inline images
- **Programmatic tab switching:** `tmux select-window -t rocket-fuel:worker-alpha` brings that tab to the foreground in iTerm2. This is how the Integrator surfaces work for the Visionary.
- macOS notifications via `osascript` for when iTerm2 isn't focused

### Why GitHub Projects?

The Integrator's state is a GitHub Project board. No hidden state. Everything visible in the GitHub web UI too.

- **Columns:** Someday/Maybe → Backlog → Scoped → In Progress → Review → Done
- **Labels:** Skill routing (`workflow:tdd`, `workflow:bug-fix`, `workflow:epc`)
- **Milestones:** Time-boxing and velocity tracking
- **Automation:** Issue closed → card moves to Done

### Why git worktrees?

Each worker gets an isolated copy of the repo via `git worktree add`. This means:

- O(seconds) to spawn (shared object database, no full clone)
- Full isolation (workers can't step on each other)
- Clean branch per issue
- Auto-cleanup when work is done

### Why existing skills?

Skills (`/tdd`, `/issue-scope`, `/epc`, `/bug-fix`, `/ship`, `/pr-quality`) are battle-tested playbooks. Workers don't need custom logic — they run the right skill for the job. The Integrator routes issues to skills based on labels.

## Principles (from the book)

These principles from *Rocket Fuel* govern how the Visionary and Integrator relate to each other. They're encoded into the Integrator's behaviour.

### 1. No end runs

The Visionary talks to the Integrator, not directly to workers. If the Visionary wants a worker doing something different, they tell the Integrator, who translates it into the plan. Going around the Integrator creates chaos — conflicting instructions, the project board out of sync, workers getting whiplash.

### 2. The Integrator filters, not just parks

The Integrator doesn't just say "noted." It actively evaluates every idea:
- Does this align with the current milestone?
- What's the effort vs impact?
- Is this a shiny object or a genuine insight?
- Does this replace something on the board, or is it additive?

Some ideas go to Someday/Maybe. Some get fast-tracked to Backlog. Some get pushed back on — "we tried something similar and it didn't work because..."

### 3. Clear ownership boundary

The Visionary owns the **what** and **why**. The Integrator owns the **how** and **when**.

"We need better onboarding" = Visionary. How to break it down, which sprint, which workers, what order = Integrator. The Visionary doesn't override execution decisions.

### 4. Tie-breaker rules

- Conflict about **product direction** (what to build, why) → Visionary wins
- Conflict about **execution** (priorities, sequencing, approach) → Integrator wins

### 5. The Same Page Meeting

When the Visionary opens the Integrator tab after being away, the Integrator catches them up — milestone progress, what shipped, what's blocked, what's next. No prompting needed. The Integrator proactively summarises at natural milestones.

### 6. The Visionary going "below the line" is tracked

When the Visionary jumps into a worker tab, the Integrator should know. Otherwise the project board is out of sync with reality.

### 7. The Elon Problem (UNRESOLVED)

> **This needs more thinking.** The book says the Visionary should stay "above the line" — directing through the Integrator, not getting into the weeds. But some Visionaries (Elon is the archetype) are brilliant precisely because they get their hands dirty. They spot things on the factory floor that no amount of delegation would surface.
>
> In our context: the Visionary should be able to jump into a worker tab and start coding. But how does the Integrator handle this?
>
> Options to explore:
> - The Integrator detects the Visionary is active in a worker tab and pauses that worker's autonomous behaviour
> - The Visionary "checks out" an issue from the Integrator before going below the line
> - The Integrator simply observes and updates the board state after the Visionary finishes
> - The worker tab gets a special mode when the Visionary is present
>
> The tension: too much structure kills the Visionary's spontaneity. Too little structure means the Integrator can't track reality. Need to find the sweet spot.

## Why this matters

The traditional software development role was: understand requirements, write code, ship it. AI now does all three. The entire profession is being restructured around this reality.

### The developer role is dead. Long live the Visionary.

For decades, the developer was the bottleneck. You needed someone who could translate ideas into code. That person was expensive, scarce, and powerful — they were the only ones who could build the thing.

AI removes that bottleneck. Code generation, test writing, debugging, refactoring — these are commoditised. A Visionary with AI tools can ship what used to require a team. The skill that matters now isn't writing code. It's knowing *what to build and why*.

Developers who survive this shift won't do so by writing better code than AI. They'll do it by becoming Visionaries — developing taste, product judgment, and the ability to see what users actually need. The ones who cling to "I write the code" as their identity will find that identity increasingly worthless.

### The Integrator is AI's natural role

The Integrator role — tracking progress, managing priorities, monitoring CI, dispatching work, keeping the machine running — is structured, disciplined, relentless work. Humans are terrible at it. We get bored, lose focus, chase shiny objects, forget to follow up.

AI doesn't. An AI Integrator has perfect oversight, never gets distracted, never drops a thread, and can manage 20 workers simultaneously without breaking a sweat. This is the role AI was born to play.

### The new stack

```
Human  →  Visionary   (taste, judgment, direction, "what" and "why")
AI     →  Integrator  (execution, oversight, discipline, "how" and "when")
AI     →  Workers     (code, tests, PRs, the actual building)
```

This isn't a productivity hack. It's a fundamental restructuring of who does what in software development. The human's value is no longer their ability to type code — it's their ability to see what matters.

Rocket Fuel is this stack as a product.

### Where new people enter

This restructuring opens the door wider than it's ever been. Previously, you needed years of technical training to contribute to software. Now, anyone with vision, taste, and the ability to articulate what they want can ship real software.

The barrier to entry shifts from "can you code?" to "can you see what needs to exist?" That's a fundamentally different — and much larger — talent pool.

## What Rocket Fuel is NOT

- **Not a platform.** It's ~1000 lines of glue, not 371K lines of infrastructure.
- **Not gastown-lite.** Different philosophy entirely. No beads, no Dolt, no daemons, no custom merge queues, no federation.
- **Not a task runner.** The Visionary/Integrator dynamic creates natural tension between ideas and execution. The Integrator protects progress from the Visionary's destabilising energy while never losing an idea.
- **Not a web app.** It's a CLI + tmux. The "UI" is iTerm2.

## Prior Art

- [gastown](https://github.com/steveyegge/gastown) — Steve Yegge's multi-agent workspace manager. 371K lines of Go. Impressive but complex. Reference implementation lives at `~/gastown`.
- *Rocket Fuel* by Gino Wickman & Mark C. Winters — the Visionary/Integrator framework.

## Tech

- **Language:** Go (single binary, goroutines for worker monitoring, proven by gastown)
- **CLI framework:** Cobra
- **Dependencies:** tmux, iTerm2, Claude Code, gh CLI, git
- **Distribution:** Homebrew tap, GitHub releases
