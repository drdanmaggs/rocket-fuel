# ADR-005: Merge Safety — PreToolUse Hook Replaces Permission Prompts

## Status: Proposed

## Context

The Integrator needs to merge PRs autonomously but we can't allow unsafe merges. The original approach was a Claude Code permission prompt on `gh pr merge` — but this blocks the Integrator dead until the Visionary clicks "Yes."

## Decision

Replace the permission prompt with a **PreToolUse hook** that programmatically enforces merge safety. The Integrator is free to attempt merges at any time — the hook gates them based on CI status.

## Rules

| Check | Result | Action |
|-------|--------|--------|
| All CI checks passed | Green | Allow merge (exit 0) |
| CI still running | Pending | Block: "CI not complete yet" (exit 2) |
| Any CI check failed | Red | Block: "CI failing, do not merge" (exit 2) |
| PR is draft | Draft | Block: "PR is still a draft" (exit 2) |

## Implementation

```json
{
  "PreToolUse": [{
    "matcher": "Bash(gh pr merge*)",
    "hooks": [{
      "type": "command",
      "command": "rf watchdog check-merge-safety"
    }]
  }]
}
```

`rf watchdog check-merge-safety`:
1. Parse PR number from the intercepted command
2. Query `gh pr view <N> --json statusCheckRollup,isDraft`
3. Apply rules above
4. Exit 0 (allow) or exit 2 with reason (block)

## Trade-offs

### What we gain
- Integrator never blocked by permission prompts
- Safety enforced by code, not by prompt instructions
- Deterministic — same PR always gets the same result
- The Visionary doesn't need to babysit merges

### What we lose
- No manual approval step (the Visionary trusts CI as the gate)
- If CI is broken/misconfigured, bad merges could slip through

### Mitigation
- CI must be comprehensive (unit + integration + e2e)
- The Visionary can still review PRs manually via GitHub
- `gh pr merge --squash` preserves revert capability
- Mission Control shows PR status — Visionary has ambient awareness

## Why not an approval queue?

The original plan was a dashboard-based approval queue where the Visionary manually approves each merge. This is safer but kills momentum — the Integrator queues merges and waits. With a CI gate, green PRs merge immediately. The Visionary's approval is expressed through CI: if the tests pass, the code is safe.

## Relationship to other ADRs

- ADR-001 (Hooks): PreToolUse is one of the 22 hook types
- ADR-002 (Roles): Watchdog enforces rules, Integrator makes decisions
- ADR-003 (Mission Control): Stream Deck shows merge status for ambient awareness
