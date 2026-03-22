---
name: tdd
description: >
  Test-driven development with context-isolated RED/GREEN/REFACTOR phases.
  Use when implementing features, fixing bugs, or adding functionality.
  Auto-triggers on: "implement", "add feature", "build", "fix bug",
  "something broken", "not working", "add endpoint", "create component",
  "TDD", "test-driven". Supports both feature development and bug fixing
  with separate subagent contexts to prevent test contamination.
---

# Test-Driven Development

Context-isolated RED/GREEN/REFACTOR. Each phase runs in a separate subagent to prevent test contamination — the single most impactful improvement for AI-assisted TDD.

**Architecture:** Orchestrator (you) stays lean. Subagents read their own context from disk. You compose short "phase briefs" pointing to files, not paste file contents. See [phase-prompts.md](references/phase-prompts.md) for templates.

**Lifecycle:** This skill spans the ENTIRE session. After planning completes (Stage 0), continue with Session Constants → The Loop. **Never use `EnterPlanMode`** — it causes context amnesia. Use the Plan subagent instead.

## Mode Detection

| User Signal | Mode |
|-------------|------|
| "implement", "add feature", "build", "create" | **Feature** |
| "fix", "broken", "not working", "error", "bug" | **Bug-Fix** |
| Ambiguous | Ask user |

## Exceptions — When NOT to TDD

Skip TDD for: **UI styling**, **exploratory spikes**, **static content**, **one-off scripts**.

See [when-to-skip-tdd.md](references/when-to-skip-tdd.md) for detailed criteria, examples, and decision framework.

---

## Stage 0: Setup (One-Time Per Feature)

### Branch Creation (First Thing)

Before any exploration or planning, pull latest main and create a feature branch:

```bash
git checkout main && git pull origin main && git checkout -b <type>/<short-slug>
```

- **Name it yourself** — don't ask the user. Derive from the feature/bug description.
- `feat/add-category-search`, `fix/duplicate-household-creation`, `refactor/extract-validation-logic`
- Use the same `type` prefix as the commits (`feat/`, `fix/`, `refactor/`)
- Keep it short (3-5 words max, kebab-case)

This ensures all exploration, planning, and coding happens against the latest codebase.

### Planning Decision Tree

**0-check:** Look for existing plan file in `docs/plans/*.md`
- Found → Validate with Explore agent, proceed to 0-review
- Not found → Continue to 0-plan

**0-plan:** Apply planning heuristic
- Question: "Could a junior dev with limited intelligence implement this?"
- Outcome A (**COMPLEX**) → Auto-generate plan (Plan subagent) with user Q&A first
- Outcome B (**INTERMEDIATE**) → Auto-generate plan (Plan subagent)
- Outcome C (**SIMPLE**) → Offer skip planning (proceed directly to 0-review)

**NEVER use `EnterPlanMode`.** It causes context amnesia — the orchestrator forgets the TDD skill after exiting. All planning goes through the Plan subagent (stays in the same conversation context).

**0-review:** Quality gate (FRESH EYES)
- **Skip heuristic:** Is this plan stupid simple and obvious?
  - Pure utility (no side effects, no integration), formatting/styling only, trivial scope → **SKIP** ("Obviously simple")
  - Anything else → **RUN** (get fresh eyes, don't bet on blind spots)
- If RUN: Spawn `tdd-plan-reviewer` (Opus) subagent
  - Reads plan file + codebase (fresh context, no prior conversation)
  - Evaluates: testability gaps, YAGNI violations, missing slices, architectural fit
  - Returns categorized findings (Critical/Warning/Note)
- Orchestrator presents findings, asks clarifying questions if needed
- User approves plan or requests changes
- If changes needed → update plan, optionally re-review
- Plan approved → Proceed to **Stage 0f** (Acceptance Test)

**0f — Acceptance Test (GOOS Outer Loop)**
After plan is approved, before inner loop begins. Use Sequential Thinking: does this feature have user-facing behaviour (new page, form, flow, navigation, toast, dialog)?
- YES → spawn tdd-test-writer in `Mode: acceptance` (see [phase-prompts.md](references/phase-prompts.md#stage-0f-acceptance-test-goos-outer-loop))
- NO → skip 0f, set `Acceptance test path: none` in session constants

**Gate:** test written + fails for expected reason → commit `test: add acceptance test for {feature_name}`

Then proceed to **Session Constants** (Stage 0a) → inner loop.

**See [planning-workflow.md](references/planning-workflow.md) for:**
- Complete heuristic evaluation criteria
- Detailed flow for each path (auto-plan / manual)
- Stage 0-auto Plan subagent prompt template
- Stage 0-review plan-reviewer subagent prompt template
- Stage 0a-0d manual planning flow

### Session Constants

Discovered during Stage 0 (Explore agent) and reused in every phase brief. When discovering constants, also find: E2E test command (`playwright` in package.json scripts), E2E test directory (look for `tests/e2e/` or `e2e/`), and auth fixture path (look for `auth-fixture.ts` in E2E directories).

| Constant | Example | Notes |
|----------|---------|-------|
| Test command | `pnpm vitest run --reporter=verbose --bail 1` | `--bail 1` stops on first failure — faster feedback in GREEN and REFACTOR gates |
| Test file pattern | colocated `*.test.ts` or `tests/__tests__/` | |
| Test helpers | `tests/helpers/isolated-test-household.ts` | |
| Standards file | `.claude/skills/tdd/references/test-standards.md` | |
| Bug context | PR #123 or main..fix/bug or diagnosis summary | (empty for features) |
| E2E test command | `pnpm playwright test` | Discovered in Stage 0a |
| E2E test directory | `tests/e2e/` | Discovered in Stage 0a |
| Auth fixture | `tests/e2e/fixtures/auth-fixture.ts` | Discovered in Stage 0a |
| Acceptance test path | `tests/e2e/feature.spec.ts` | Set after Stage 0f. "none" if non-user-facing. |

### Task Tracking

After plan is saved, create one `TaskCreate` per slice:

```
subject:     "Slice N: {slice_description}"        # e.g. "Slice 1: Core validation logic"
activeForm:  "Testing Slice N: {slice_description}" # e.g. "Testing Slice 1: Core validation logic"
description: Copy the test bullets from the plan slice (provides context when viewing task details)
```

This gives a live progress view throughout the session — the user can see which slices are pending, in progress, and done at a glance.

---

## The Loop: TIDY → RED → GREEN → REFACTOR

**This is the core of TDD. Keep it tight.**

For each iteration, you (the orchestrator) do exactly this:

### 1. Read Plan → Find Next Test

Read the active plan file in `docs/plans/`. Find the **first unchecked** `- [ ]` item (simple regex search: `- \[ \]`). Note the slice, test type, and description.

**Plan contract:** The first unchecked item is ALWAYS the next test. No ambiguity, no interpretation needed. Plan files are processed strictly top-to-bottom.

**If this is the first unchecked test of a new slice** → `TaskUpdate` the slice's task to `status: in_progress`.

### 2. TIDY FIRST — Prep refactoring (always)

**Always spawn the tdd-refactorer subagent for prep analysis.** It will read the code the next test will touch and decide whether prep tidying is needed.

Why this works:
- Agent evaluates actual code structure, not orchestrator's guess
- Fast decisions (~5 sec with Haiku) when code is ready
- Consistent with POST-GREEN refactor approach
- Removes heuristic from orchestrator

Compose phase brief (see [phase-prompts.md](references/phase-prompts.md#tidy-first-phase-tdd-refactorer-prep)) with next test description and files to be modified.

**Refactorer will either:**
- Skip: "Code is ready — clean structure, easy to extend"
- Prep tidy: Make structural changes to create space for new behavior

**GATE:** ALL tests still PASS after tidying (when tidying is done).

**If prep tidy was done, commit separately:**

```
refactor: [prep tidy — what was restructured and why]
```

**Kent Beck's insight:** Tidying BEFORE adding behavior (prep) is often cleaner than forcing new code into messy structure.

### 3. RED — Spawn tdd-test-writer

**Before spawning:** Use Sequential Thinking to verify:
1. Which specific test from the plan is next?
2. How many tests should tdd-test-writer write? (Answer: EXACTLY ONE)
3. Why one? (TDD requires tight feedback: 1 test → 1 implementation → 1 refactor. Batching breaks the cycle.)

Compose a lean phase brief using the template from [phase-prompts.md](references/phase-prompts.md#red-phase-tdd-test-writer). Pass file paths, not file contents.

Spawn `tdd-test-writer` subagent.

**GATE 1 — Test Count:** Verify tdd-test-writer added EXACTLY ONE test.
- One test added → proceed to GATE 2
- Multiple tests → **HARD FAILURE**, respawn with stronger constraint

**GATE 2 — Test Failure:** Test must FAIL.
- Expected assertion failure → proceed to GREEN
- Syntax/import error → fix test (still RED, same subagent or respawn)
- Test passes → test doesn't cover new behavior, respawn with adjusted brief
- Previous tests must still pass (no regressions)

### 4. GREEN — Spawn tdd-implementer

Compose phase brief (see [phase-prompts.md](references/phase-prompts.md#green-phase-tdd-implementer)) with:
- Feature name, test command
- Test file path and one-line failure summary from RED
- Source file paths to read

**Model selection (smart routing):**
- **Default (Haiku):** Single-file implementations, localized changes → 4-5x faster
- **Override to Sonnet:** Multi-file changes, cross-file reasoning, complex integrations → use `model: "sonnet"` parameter in Task tool

Use Sequential Thinking if scope is unclear. When in doubt, start with Haiku — if it struggles, respawn with Sonnet.

Spawn `tdd-implementer` subagent.

**GATE:** ALL tests in scope must PASS.
- **Unit tests** → subagent runs test file only.
- **Integration tests** → subagent runs the affected feature's test file(s) only — not the full suite.
- **Full regression suite** → CI on PR (isolated environment, authoritative gate).
- All pass → proceed
- New test fails → revise implementation (same subagent or respawn)
- Previous test fails → implementation broke something, fix without modifying tests
- Complex failures → spawn `general-purpose` agent to diagnose

**Trust subagent results.** The orchestrator reads the subagent's output and checks the gate. Do NOT re-run tests the subagent already ran.

**After GREEN passes:** Mark the test `- [x]` in the plan file. **Commit the behavioral change:**

```
test: [test description]
feat: [what was implemented to pass it]
```

### 5. REFACTOR — Spawn tdd-refactorer (always)

**Always spawn the tdd-refactorer subagent.** It will analyze the code and decide whether refactoring is needed or skip.

Why this works:
- Agent has better context by reading the actual code
- With Haiku model, "no refactoring needed" decisions are fast (~5 sec) and cheap
- Removes heuristic judgment from orchestrator
- Consistent process every cycle

Compose phase brief (see [phase-prompts.md](references/phase-prompts.md#refactor-phase-tdd-refactorer)) with test + implementation file paths.

**Refactorer will either:**
- Skip: "No refactoring needed — change is clean/small/well-structured"
- Refactor: Make structural improvements, verify tests still pass

**GATE:** ALL tests still PASS after refactoring (when refactoring is done).

**If refactoring was done, commit separately** (Tidy First — never mix structural and behavioral):

```
refactor: [what was tidied]
```

### 6. Next Test or Complete

Read the plan file:
- **More unchecked tests in current slice?** → return to step 1
- **Slice complete?** → **always** spawn tdd-refactorer on the full slice (cumulative mess builds even when individual cycles are small), then update slice status in plan (`Status: done`), `TaskUpdate` the slice task to `status: completed`, return to step 1
- **All tests checked?** → go to Completion

### Warm Loop (Token Optimization)

**Trigger:** After 3+ consecutive clean cycles (RED → GREEN → REFACTOR with all gates passing first try)

**The pattern is proven. Trust it. Minimize orchestrator tokens.**

#### What to Stop Doing

❌ **Explanatory commentary** about the process
- "Now I'll spawn tdd-test-writer to write the failing test"
- "The test failed as expected, which is good"
- "Let me spawn tdd-implementer to make it pass"

❌ **Reminders** about gates or constraints
- "Remember, the test must fail"
- "We need exactly one test"
- "Make sure all tests still pass"

❌ **Verbose summaries** after each phase
- "The tdd-test-writer successfully created a failing test for..."
- "The tdd-implementer has successfully made the test pass by..."

❌ **Redundant reads**
- Don't re-read plan file if you already know next test
- Don't re-read test file if gate just passed
- Trust subagent results

#### What to Keep Doing

✅ **Minimal status updates** (1 line each)
```
TIDY FIRST: Skip (code clean, ready to extend)
RED: Spawning tdd-test-writer for "should reject duplicate names"
RED GATE: ✓ One test, fails as expected
GREEN: Spawning tdd-implementer
GREEN GATE: ✓ Test passes, no regressions
REFACTOR: Skip (change small, code clean)
Committed: feat: reject duplicate category names
```

✅ **Gate verification** (but no explanation)
```
GATE 1: One test added ✓
GATE 2: Test fails (expected) ✓
```

✅ **Error handling** (when gates fail, switch back to verbose)
```
GATE 1: FAIL - Multiple tests added
[Return to normal verbosity, explain issue, respawn with stronger constraint]
```

#### Token Savings Example

**Cold loop (first 3 cycles):**
~300 tokens orchestrator commentary per cycle

**Warm loop (after 3+ cycles):**
~50 tokens orchestrator commentary per cycle

**Savings:** 250 tokens × 10 tests = 2,500 tokens saved in a typical feature

#### When to Exit Warm Loop

Return to normal verbosity when:
- Any gate fails (need to debug)
- New slice starts (reset context)
- User asks a question (provide context)
- Complex refactoring needed (explain trade-offs)

After resolving issue, return to warm loop if 3+ clean cycles resume.

---

## Completion

Commits have been happening throughout the loop (behavioral after GREEN, structural after REFACTOR). At completion:

1. Mark plan file complete (all slices `Status: done`)
2. Verify `git log --oneline` shows clean commit history — each commit is either behavioral or structural, never mixed

### Outer Loop: Acceptance Test Gate

If `Acceptance test path: none` → skip this section.

Otherwise, run: `{e2e_test_command} {acceptance_test_path}`

**PASSES** → proceed to completion banner.

**FAILS** → gap cycle (max 2):
1. Use Sequential Thinking to identify the specific failing assertion
2. Create a "Gap" slice: describe the missing behaviour precisely
3. Run one RED-GREEN-REFACTOR cycle for the gap (unit/integration mode, not acceptance mode)
4. Re-run acceptance test

After 2 gap cycles if still failing, surface to user with: the failure message, possible causes (env issue vs logic gap vs incorrect test assumption), and ask: fix the test / skip and ship / investigate further.

### Bug-Fix Mode → Ship

Invoke `Skill("ship")`.

`/ship` takes it from here: lint/type-check → code review → fix loop → draft PR → autonomous review-loop → undraft → CI → READY TO MERGE.

### Feature Mode → Summary + Handoff

Stop. Do **not** invoke `/ship`. Deliver the Feature Summary to the user.

**Step 1 — Output the completion banner first:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ███████╗███████╗ █████╗ ████████╗██╗   ██╗██████╗ ███████╗
  ██╔════╝██╔════╝██╔══██╗╚══██╔══╝██║   ██║██╔══██╗██╔════╝
  █████╗  █████╗  ███████║   ██║   ██║   ██║██████╔╝█████╗
  ██╔══╝  ██╔══╝  ██╔══██║   ██║   ██║   ██║██╔══██╗██╔══╝
  ██║     ███████╗██║  ██║   ██║   ╚██████╔╝██║  ██║███████╗
  ╚═╝     ╚══════╝╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚══════╝

   ██████╗ ██████╗ ███╗   ███╗██████╗ ██╗     ███████╗████████╗███████╗
  ██╔════╝██╔═══██╗████╗ ████║██╔══██╗██║     ██╔════╝╚══██╔══╝██╔════╝
  ██║     ██║   ██║██╔████╔██║██████╔╝██║     █████╗     ██║   █████╗
  ██║     ██║   ██║██║╚██╔╝██║██╔═══╝ ██║     ██╔══╝     ██║   ██╔══╝
  ╚██████╗╚██████╔╝██║ ╚═╝ ██║██║     ███████╗███████╗   ██║   ███████╗
   ╚═════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚══════╝╚══════╝   ╚═╝   ╚══════╝

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Step 2 — What was built**

Run `git log --oneline origin/main..HEAD` and `git diff --stat origin/main..HEAD` to produce:
- A plain-English summary of each plan slice and what it delivers
- Key files created or modified
- Total tests written and what behaviours they cover

**Step 3 — How to test it manually**

Walk through the key user-facing flows step by step. Cover:
- The happy path end-to-end (exact steps: navigate to X, click Y, expect Z)
- Any required setup (seed data, env vars, feature flags)
- Notable edge cases worth verifying by hand

**Step 4 — Offer to start the dev server**

Discover the dev command from `package.json` scripts (typically `pnpm dev`). Then ask:

> "Want to test this in the browser? I can start the dev server for you — just say the word."

If the user confirms, run the dev command in the background.

**Step 5 — Next steps**

Remind the user:

> "When you're happy with it, run `/ship` to lint, type-check, get a code review, and open a PR."

---

## Bug-Fix Mode

### 0c-bug. Reproduce

- **UI bugs**: Playwright MCP — navigate, interact, capture screenshots, check console
- **Backend bugs**: Supabase MCP — logs, database state, advisors (dev branch only)
- **Can't reproduce?** Report findings, gather more context from user

### 0d-bug. Diagnose

Use Sequential Thinking for complex bugs. State clearly:
1. **What** is broken (specific code/logic)
2. **Why** it's failing (mechanism)

**Use Context7** if uncertain about framework behavior.

Present diagnosis to user. **Wait for confirmation.**

### RED — Write Two Diagnostic Tests

Spawn `tdd-test-writer` **twice** — once per test — to write **two** failing tests:

1. **API-level test** — reproduces the bug from the user's perspective (integration boundary)
2. **Narrowest test** — isolates the exact root cause (smallest possible reproduction)

Both must:
- Assert CORRECT/EXPECTED behavior
- FAIL against current broken code

**GATE:** Both tests must FAIL.
- Both fail with expected assertions → diagnosis confirmed, proceed
- **Either test PASSES → diagnosis is WRONG** — return to Diagnose
- Test errors → fix the test, stay in RED

### GREEN — Implement Fix

Same as Feature loop step 4. Minimal fix targeting root cause.

**GATE:** Both diagnostic tests + all previous tests must PASS.

**Commit the fix (behavioral):**

```
fix: [description]

Root cause: [1-2 sentences]
Tests: 2 diagnostic tests + [N] regression tests

Closes #[issue]
```

### REFACTOR

Same as Feature loop step 5. If refactoring done, commit separately as `refactor:`.

### Stage 7: Regression Analysis Heuristic

**Position:** After all RED-GREEN-REFACTOR cycles complete, before final commit

**Applies to:** Both feature and bug-fix modes

**Use Sequential Thinking to evaluate:**

> "Given test coverage and change scope, is deep edge case analysis (10 min, Opus) likely to find meaningful gaps?"

**Regression analysis looks for:**
- Untested edge cases (boundary conditions, null/empty, concurrent access)
- Integration risks (timing issues, race conditions, cascading effects)
- Error states not covered (network failures, timeouts, invalid responses)
- Assumptions that might not hold (user behavior, data shape, environment)

**Auto-SKIP (low ROI for analysis):**
- Narrow scope (1-3 files, single module, localized change)
- Simple logic with clear boundaries
- Tests confirm expected behavior and cover obvious edge cases
- Non-critical domain (UI components, utilities, formatting, documentation)
- No complex interactions (no async, no shared state, no external deps)

**Auto-RUN (high ROI for analysis):**
- Critical domain (security, auth, data integrity, payments, RLS policies)
- Complex interactions (async operations, state management, timing dependencies)
- Wide integration surface (API boundaries, database operations, external services)
- Bug fix (by definition we missed something - what else might we miss?)
- Likely untested edge cases (concurrent updates, error conditions, boundary values)
- New patterns with subtle failure modes
- Feature where acceptance test required a gap cycle → something was missed; check for more

**Ask user (genuinely uncertain, rare):**
- Medium complexity with moderate test coverage
- Non-critical but non-trivial (could benefit from analysis, but not essential)
- User might have context about risk tolerance

**Auto-skip outcome:**
```
✅ All tests passing

Regression analysis: SKIP
Reason: Localized fix (2 files, simple logic), tests confirm behavior, low-risk domain.

Proceeding to completion.
```

**Auto-run outcome:**
```
✅ All tests passing

Regression analysis: RUNNING
Reason: Critical domain (RLS policy), likely untested edge cases (cascades, soft deletes).

1. Spawn regression-analyst subagent (general-purpose, Opus model)
   - Reads regression-analysis.md for edge case methodology
   - Returns findings: untested scenarios, integration risks, boundary conditions

2. Auto-filter findings using QUICK WIN check (parallel subagents):
   For each finding, spawn general-purpose subagent to evaluate:
   ✓ Predictable cost (< 5 min, specific change, no cascading)
   ✓ Real benefit (fixes actual weakness, not speculative)
   ✓ Confident correctness (familiar pattern, testable)

   All three pass → QUICK WIN (implement now)
   Any fail → SKIP (defer to GitHub issue)

3. Process QUICK WIN findings:
   - Spawn tdd-test-writer for each QUICK WIN
   - Write tests using TDD approach
   - Commit: `test: regression coverage for [pattern class]`

4. Process SKIP findings:
   - Create GitHub issue for each
   - Tag: "regression-analysis", "test-gap"
   - Link to original PR/commit

No user prompt needed - automatic YAGNI-compliant filtering.
```

**Decision principle:**

Auto-skip for simple/localized changes. Auto-run for critical/complex changes. When genuinely uncertain (rare), default to AUTO-RUN — QUICK WIN filter will auto-skip speculative findings anyway, so there's minimal cost to running the analysis.

---

## Rules

### Hard Gates (NEVER bypass)

1. **ONE TEST gate**: RED phase writes EXACTLY ONE test per cycle. Never batch. No exceptions.
   - "They're testing the same function" — NOT an excuse
   - "They're simple tests" — NOT an excuse
   - "It's faster to batch" — BREAKS TDD, not allowed
2. **RED gate**: Test must FAIL before GREEN
3. **GREEN gate**: ALL tests must PASS before REFACTOR
4. **REFACTOR gate**: ALL tests must still PASS
5. **Bug diagnosis gate**: If diagnostic test passes, diagnosis is wrong

### Anti-Cheat

- Never modify tests during GREEN phase
- Never write implementation during RED phase
- **Never edit test + implementation in the same step** — see Mid-Session Refinements below
- Never delete or comment out tests
- Never use `@ts-ignore` or `as any` to force tests to pass
- Never skip gates

### Mid-Session Refinements (CRITICAL)

When the user requests changes during or after a TDD session (e.g. "change the text", "move the badge below the name", "make it consistent"):

1. **It's still TDD.** The session isn't over. Don't downgrade to ad-hoc editing.
2. **RED first**: Update the test to assert the new expected behavior. Watch it fail.
3. **GREEN**: Make the code change. Watch it pass.
4. **No excuses**: "It's just CSS", "it's just a label", "it's trivial" are not valid reasons to skip RED → GREEN. If there's a test covering the behavior, update the test first. Even "just updating a label" requires RED then GREEN.
5. **Commit prefix**: Refinements to an in-progress feature are `feat:`, not `fix:`. Reserve `fix:` for actual bugs with diagnostic tests.

### Commit Discipline

Per [`rules/commit-discipline.md`](~/.claude/rules/commit-discipline.md):
- **Commit after each GREEN** (behavioral: `test:` + `feat:` or `fix:`)
- **Commit after each REFACTOR** (structural: `refactor:`)
- **Never mix** structural and behavioral in one commit
- **Never commit** with failing tests or new warnings
- Small, frequent commits — one logical unit each

### Workflow

- Branch is created at the start of Stage 0, before exploration (see Branch Creation above)
- Tests before implementation (no exceptions)
- Wait for user approval after planning
- Use Context7 for framework behavior verification
- Follow project test conventions (check existing tests first)

