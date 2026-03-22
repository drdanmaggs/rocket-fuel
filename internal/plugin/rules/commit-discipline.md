# Commit Discipline (Tidy First)

**Core:** Separate structural changes from behavioral changes. Never mix in one commit.

## Two Commit Types

| Type | Examples | Prefix |
|------|----------|--------|
| **Structural** | Renames, extract method, move file, reorder | `refactor:` / `tidy:` |
| **Behavioral** | New feature, bug fix, changed logic | `feat:` / `fix:` |

**Order:** Structural first, then behavioral. Never combined.

## Commit Preconditions

**NEVER commit unless ALL of these are true:**
- All tests passing
- No compiler/linter warnings introduced
- Change is a single logical unit

## Commit Messages

State the type explicitly:
- `refactor: extract validateInput from handleSubmit`
- `feat: add category search endpoint`
- `fix: prevent duplicate household creation`

## Why This Matters

- PRs are easier to review (structure vs logic separated)
- Reverts are safer (behavioral revert won't undo refactors)
- Git blame stays meaningful
