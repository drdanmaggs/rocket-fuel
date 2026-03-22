# ADR-007: Dedicated Rocket Fuel Board Per Repo

## Status: Active

## Context

When `rf launch` auto-discovers an existing GitHub Project board in a repo, the Integrator doesn't understand the board's structure — different column names, conventions, and context it wasn't designed for. This was observed when testing on Productivity Coach: the Integrator found the board but couldn't effectively use it, and autonomously closed an epic without consulting the Visionary.

## Decision

Rocket Fuel creates its own dedicated GitHub Project board per repo with a standard column structure. The Integrator owns this board completely.

### Standard Columns

| Column | Purpose |
|--------|---------|
| **Backlog** | New issues awaiting scope review |
| **Ready** | Scoped issues ready for worker assignment |
| **Scoped** | Currently assigned to a worker |
| **In Progress** | Worker is actively working |
| **In Review** | PR open, waiting for review/merge |
| **Done** | Completed and closed |

### Board Creation

A `board-setup` skill in the Claude Code plugin guides the Integrator through board creation using `gh` CLI. The board starts clean — no auto-importing of existing issues. The Integrator populates it on demand as work is scoped and dispatched.

### Existing Boards

Existing repo boards are left untouched. The Rocket Fuel board coexists alongside them. The board title clearly identifies it as Rocket Fuel's workspace.

## Relationship to #68

Issue #68 (GitHub Project board setup automation) covers GitHub Actions for board automation (PR merged → move to Done, auto-labeling). This ADR covers the board structure and creation strategy. The two complement each other — #68 automates the cloud-side, this ADR defines the board the automation operates on.

## Consequences

- Integrator knows the exact column structure (it defined it)
- No confusion with existing board conventions
- Clean board — focused on active Rocket Fuel work
- Two boards per repo (existing + Rocket Fuel) — acceptable trade-off for clarity
- Board creation is a one-time setup step per repo
