# ADR-009: Split general-purpose skills into their own repo

**Status:** Accepted
**Date:** 2026-07-25
**Supersedes:** part of [ADR-006](006-hybrid-plugin-architecture.md)

## Context

ADR-006 established the hybrid plugin model: agent definitions, skills and rules live in
`internal/plugin/`, embedded via `go:embed` and shipped alongside the orchestrator. At the
time this looked like cohesion — one install, one version, everything together.

In practice the repo accumulated 26 skills, 17 agents and 8 rules. An audit found:

**The coupling ran the wrong way.** Only `board-setup` and `worktree-reset` depended on
Rocket Fuel; only `integrator.md` and `worker.md` among the agents. The remaining 24 skills
and 15 agents — `/tdd`, `/ship`, `/code-reviewer`, `/test-fixer` and the rest — had no
knowledge that an orchestrator existed. They were a general-purpose toolkit that happened
to be riding in the orchestrator's delivery van. Shipping a one-line fix to `/ship`
required bumping the version of an unfinished Go project.

**The `go:embed` justification was no longer live.** ADR-006 assumed skills would ship
inside the binary. Installation moved to Claude Code's native marketplace, and
`ExtractPlugin` has had no production callers since. Co-location bought nothing.

**Private content was public.** 6.7 MB of the repo's 8.7 MB was a YouTube ideation skill
containing channel-specific strategy and notes from paid mastermind sessions, published to
a public repo. Anyone cloning the orchestrator got it.

**`rules/` never worked at all.** `rules/` is not a Claude Code plugin primitive. The eight
files shipped to the install cache and were never loaded. The rules that actually load come
from `~/.claude/rules/`, where four of the five overlapping files were byte-identical
duplicates — and `testing.md` had already silently drifted between the two copies, which is
exactly the failure mode duplication invites.

## Decision

Three repos, each with one job.

| Repo | Visibility | Contents |
|---|---|---|
| `drdanmaggs/claude-skills` | public | 24 general-purpose skills, 15 agents, the TDD phase-gate hook |
| `drdanmaggs/youtube-skill` | private | YouTube ideation skill, images intact |
| `drdanmaggs/rocket-fuel` | public | Go orchestrator + `board-setup`, `worktree-reset`, `integrator`, `worker` |

`rules/` is deleted outright rather than moved — the correct home is `~/.claude/rules/`,
which already has it.

The equestrian images were purged from this repo's git history with `git filter-repo`. The
repo had 0 stars and 0 forks, and the content entered in a single commit, making the
rewrite unusually cheap.

Workers now dispatch to `/claude-skills:tdd` rather than `/tdd`. **Both plugins are
required** for Rocket Fuel to function.

## Consequences

**Good:**
- The toolkit ships on its own cadence, unblocked by an unfinished orchestrator.
- Rocket Fuel's repo is ~1 MB instead of 8.7 MB, and is actually about orchestration.
- Private content is private.
- No more silent rule drift — one copy, in the place that loads it.

**Bad:**
- Two marketplaces to `claude plugin marketplace update` instead of one.
- Skill invocations are longer: `/claude-skills:tdd`, not `/tdd`.
- A cross-repo dependency now exists: `worker.md` and `internal/worker/worker.go` name
  skills they cannot see. Nothing enforces that the names stay valid.

**Neutral:**
- `ExtractPlugin` survives but is now demonstrably dead. Tracked for deletion.
- Nested-subdirectory extraction lost its only test fixture; no remaining plugin content
  nests. Noted inline in `extract_test.go` rather than covered with synthetic files.
