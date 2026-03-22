# Supabase Migration Workflow

This workflow applies whenever you need to change the database schema. Schema changes must be tracked in git.

## Golden Rule

The migration file in `supabase/migrations/` is the **source of truth**. Always create the file first, then apply it.

## Workflow

### Development Branch (Manual)

**Step 0 — Check state before doing anything**

Run this first, every time, no exceptions:
```bash
supabase migration list --linked
```

This shows LOCAL vs REMOTE side-by-side. Any migration in LOCAL only (not REMOTE) is unapplied. Any migration in REMOTE only (not LOCAL) is an auto-generated Supabase snapshot. Note the highest REMOTE version — your new migration timestamp must be greater than it.

```
LOCAL            │   REMOTE          │   TIME (UTC)
─────────────────┼───────────────────┼───────────────────
                 │ 20260308230819    │ ← highest REMOTE — your timestamp must beat this
20260308200000   │                   │ ← ONLY LOCAL = ordering conflict, fix before proceeding
```

If you see a LOCAL-only row below a REMOTE row, you have a timestamp conflict. Fix it by renaming the file to a timestamp greater than the highest REMOTE version before doing anything else.

**Step 1 — Choose a safe timestamp**

New migration timestamp must satisfy: `timestamp > max(highest_remote_version, today_at_midnight_UTC)`

**Convention: always use tomorrow's date at `000000`.**

Example: today is `20260309` → use `20260310000000`. This gives a 24-hour buffer against same-day auto-generated snapshots that Supabase creates at merge time.

- Never use sequential suffixes on the same date (`000001`, `000002`) — use unique timestamps per migration
- Two migrations needed at once? Space by one second: `20260310000000` and `20260310000001`

**Steps 2–6**

2. Create migration file with safe timestamp and write idempotent SQL (`IF NOT EXISTS`, `CREATE OR REPLACE`, etc.)
3. Dry run: `supabase db push --dry-run`
4. Apply to dev branch: `supabase db push`
5. Verify via Supabase MCP
6. Update `supabase/seed.sql` if needed
7. Commit migration and create PR

### Production Deployment (Automatic)
**⚠️ NEVER manually run `supabase db push` against production/main branch.**

Production migrations are applied **automatically when PR is merged to main**. The CI/CD pipeline handles deployment - no manual intervention required or permitted.

## Fixing Timestamp Conflicts

If a migration is stuck (MIGRATIONS_FAILED on production, or LOCAL-only row below a REMOTE row):

1. **Rename the file** to a timestamp greater than the highest REMOTE version (Supabase's documented fix for timestamp ordering conflicts)
2. Ensure SQL is idempotent (`IF NOT EXISTS`) so it's safe on environments where it partially ran
3. Delete the old filename, create the new one — do not modify the old file in place

**Do not use `supabase migration repair` for this.** That command modifies the tracking table without running SQL — it's for history divergence, not ordering conflicts. Using it here would either hide the bug or leave the schema broken.

## MCP Boundaries
- ✅ Read: Query data, inspect schema, verify changes
- ❌ Never use `apply_migration` — it creates timestamp mismatches. Use `supabase db push`.
