# Testing Rules

**Core:** TDD always. Tests from day 1. Every bug fix starts with a failing test.

With AI assistants, writing tests is cheap—skipping them is expensive. Tests guide implementation, catch regressions immediately, and cost ~30 seconds overhead per feature.

---

## 1. Vitest/Integration Testing

### Mock Safety

**Critical:** `vi.clearAllMocks()` only clears call history, NOT implementations. Mock contamination is the #1 source of flakiness.

**Mandatory Config (vitest.config.ts):**

```typescript
export default defineConfig({
  test: {
    restoreMocks: true, // Auto vi.restoreAllMocks() after each test
    mockReset: true, // Auto vi.resetAllMocks() before each test
    unstubGlobals: true, // Auto vi.unstubAllGlobals() after each test
    sequence: { shuffle: true }, // Detect order-dependent tests early
  },
});
```

**Prefer vi.spyOn() as Default:**

```typescript
// ✅ Surgical, per-test mocking (restored automatically)
import * as userService from "./userService";
vi.spyOn(userService, "getUser").mockRejectedValue(new Error("Not found"));
```

**vi.mock() Hoisting Trap:**

```typescript
// ❌ This vi.mock() inside a test block is STILL hoisted to file top
test("test 1", () => {
  vi.mock("./module"); // Runs BEFORE imports, affects ALL tests in file
});
```

**Reserve vi.mock() for:** Silencing entire modules (loggers, analytics), import-time side effects.

### Timing & Async

**Microtask vs Macrotask Deadlock:**

```typescript
// ❌ Timer fires but Promise.then hasn't resolved yet
vi.useFakeTimers();
render(<ComponentThatFetchesOnMount />);
vi.advanceTimersByTime(1000); // Advances setTimeout, but fetch().then() is stuck

// ✅ Use async timer APIs that flush microtasks
await vi.advanceTimersByTimeAsync(1000);
```

**Fake Timers + MSW Deadlock:**

```typescript
// ✅ Selective timer mocking (NEVER mock queueMicrotask)
vi.useFakeTimers({
  toFake: [
    "setTimeout",
    "setInterval",
    "clearTimeout",
    "clearInterval",
    "Date",
  ],
});
```

**React Testing Library:**

```typescript
// ❌ Double-polling — findBy is already waitFor + getBy
await waitFor(async () => {
  expect(await screen.findByText("Done")).toBeInTheDocument();
});

// ✅ Use findBy* directly for async elements
expect(await screen.findByText("Done")).toBeInTheDocument();

// ✅ Use waitFor with SYNC queries only
await waitFor(() => {
  expect(screen.getByRole("button", { name: /save/i })).toBeEnabled();
});
```

**Always Use userEvent:**

```typescript
// ❌ NEVER use fireEvent (single synthetic event, unrealistic)
fireEvent.click(button);

// ✅ ALWAYS use userEvent (simulates real browser behavior)
const user = userEvent.setup();
await user.click(button);
await user.type(input, "hello@test.com");
```

### Worker-Scoped Fixtures (MANDATORY)

**Problem:** `afterAll` doesn't run when process crashes, CI times out, or developer hits Ctrl+C.

**Solution:** Use worker-scoped fixtures for guaranteed cleanup:

```typescript
// Import worker-scoped fixture
import { test, expect } from "@/tests/helpers/vitest-worker-fixture";

test("my test", async ({ workerHousehold }) => {
  // workerHousehold provides: householdId, userId, authUserId, etc.
  // Cleanup ALWAYS runs (even on crash/timeout/Ctrl+C)
  const result = await someFunction(workerHousehold.householdId);
  expect(result).toBeDefined();
});
```

**Why Worker-Scoped?**

- Cleanup runs in guaranteed teardown phase (even on crash/interrupt)
- Parallel execution completely safe (each worker isolated)
- Simpler test code (no manual beforeAll/afterAll)

**Exception:** Sequential tests (rare) can use `describe.sequential()` with manual cleanup, but MUST include comment explaining why tests can't be parallel.

---

## 2. Playwright/E2E Testing

### React Hydration Safety (CRITICAL)

**Problem:** SSR HTML appears ready but React event handlers aren't attached yet (hydration gap). Playwright sees button as clickable, clicks it, nothing happens.

**Why It Flakes:** SSR HTML has no event handlers. React hydration attaches handlers AFTER Playwright's actionability checks pass. Click is a no-op.

**Solution 1: Retry Until State Changes (BEST - No Code Changes):**

```typescript
await expect(async () => {
  await page.getByRole("button", { name: "Submit" }).click();
  // If hydration incomplete, this will fail and retry
  await expect(page).toHaveURL(/\/success/);
}).toPass({ timeout: 15000 });
```

**Solution 2: Hydration Marker:**

```typescript
// app/layout.tsx - Add to root layout
<script dangerouslySetInnerHTML={{
  __html: `document.body.dataset.hydrated = 'true';`
}} />

// E2E test
await page.goto('/form');
await page.waitForSelector('body[data-hydrated="true"]');
await page.getByRole('button', { name: 'Submit' }).click();
```

**Solution 3: Disable Controls Until Hydrated:**

```typescript
// Component: <button disabled={!isMounted}>Submit</button>
// Playwright auto-waits for enabled state
```

### Timing Anti-Patterns

**NEVER Use waitForTimeout():**

```typescript
// ❌ Hardcoded wait — too long (wastes time) and too short (fails on slow CI)
await page.waitForTimeout(3000);

// ✅ Auto-retrying web-first assertion
await expect(page.getByRole("button", { name: "Submit" })).toBeVisible();
```

**Web-First Assertions (ALWAYS await expect):**

```typescript
// ❌ Checks once, immediately — fails if element hasn't rendered yet
expect(await page.getByText("welcome").isVisible()).toBe(true);

// ✅ Retries until condition met or timeout
await expect(page.getByText("welcome")).toBeVisible();
```

**CI Timeout Guidance (30s Minimum):**

```typescript
// ❌ Too tight for CI (cold starts, resource contention, network latency)
await expect(page).toHaveURL(/\/chat\/.+/, { timeout: 10000 });

// ✅ Accounts for CI latency (Lambda cold starts, shared runners)
await expect(page).toHaveURL(/\/chat\/.+/, { timeout: 30000 });
```

**Default timeouts by operation type:**

- DOM interactions: 5s (Playwright default)
- Navigation: 30s
- External API calls: 30s
- Database operations: 30s

### Dynamic Port Handling

**Problem:** Dev servers start on random ports when default is busy. Hardcoded ports cause failures.

```typescript
// playwright.config.ts
export default defineConfig({
  webServer: {
    command: "npm run dev",
    wait: { stdout: /Local:.*http:\/\/localhost:(?<next_dev_port>\d+)/ },
    stdout: "pipe", // REQUIRED to capture port
  },
  use: {
    baseURL: `http://localhost:${process.env.NEXT_DEV_PORT ?? 3000}`,
  },
});
```

### Playwright Strict Mode - Scope Selectors

```typescript
// ❌ Could match <title> tag
page.getByText("City Name");

// ✅ Role-based (best for semantic elements)
page.getByRole("heading", { name: /City Name/ });

// ✅ Scope to body (exclude <head>)
page.locator("body").getByText("City Name");

// ✅ First match (last resort, document why)
page.getByText("City Name").first(); // City appears in title + content
```

---

## 3. Data Management

### Test Data Isolation (Worker-Scoped)

**Tests must own all their data — never query for pre-existing records:**

```typescript
// ❌ Depends on seed data being present — throws on a clean DB
const { data: profile } = await supabase.from("profiles").select("id").limit(1).single();
if (!profile) throw new Error("No profiles found — run db:reset");

// ✅ Create what you need, clean it up yourself
const profileId = crypto.randomUUID();
await supabase.from("profiles").insert({ id: profileId, household_id: workerHousehold.householdId, ... });
// use profileId in test, delete in finally
```

This is the same principle as not hardcoding IDs — querying for existing records couples the test to environment state outside its control. CI on a fresh DB, a `db:reset`, or a parallel worker that deleted the row can all cause silent skips or false passes.

**Never Hardcode User IDs:**

```typescript
// ❌ Brittle, no isolation, FK errors
const USER_ID = "hardcoded-uuid";

// ✅ Dynamic, isolated, parallel-safe
const userId = crypto.randomUUID();
await supabase.from("auth.users").insert({ id: userId, ... });
```

**E2E: Worker-Scoped Fixtures:**

```typescript
import { test, expect } from "@/tests/e2e/fixtures/auth-fixture";

test("my test", async ({ workerAuth }) => {
  // workerAuth provides: email, password, userId, householdId
  // Cleanup automatic after worker completes
});
```

**Unique Test Data Naming:**

```typescript
export function testName(desc: string) {
  return `__TEST__${crypto.randomUUID().slice(0, 8)}__${desc}`;
}

// Usage
const record = await supabase
  .from("records")
  .insert({ name: testName("Record") }); // "__TEST__abc12345__Record"
```

### Cleanup Patterns

**Defense-in-Depth (Both Layers):**

1. **Pre-test cleanup script** (catches orphaned data from dev interruptions)
2. **Worker-scoped fixtures** (handles normal test execution)

```json
{
  "scripts": {
    "pretest:integration": "tsx tests/scripts/cleanup-orphaned-households.ts"
  }
}
```

**Deletion Order (CRITICAL):**
Always delete children before parents (check schema for FK constraints):

```typescript
// ✅ DELETE CHILDREN FIRST (reverse FK order)
await supabase.from("child_records").delete().in("id", childIds);
await supabase.from("parent_records").delete().in("id", parentIds);
```

**CASCADE DELETE Race Conditions:**

```typescript
// ❌ CASCADE automatically deletes ALL child records, including other workers'!
await supabase.from("households").delete().eq("id", householdAId);
// ↑ Worker B's meals suddenly disappear mid-test

// ✅ Manual child-first deletion (explicit > implicit CASCADE)
await supabase.from("meals").delete().eq("household_id", householdId);
await supabase.from("households").delete().eq("id", householdId);
```

**Non-Throwing Cleanup (E2E):**

```typescript
afterAll(async () => {
  if (!context) {
    console.warn("Setup failed, skipping cleanup");
    return;
  }

  const results = await Promise.allSettled([cleanupUser1(), cleanupUser2()]);

  // Log failures without throwing
  results.forEach((result, i) => {
    if (result.status === "rejected") {
      console.warn(`Cleanup ${i} failed:`, result.reason);
    }
  });
});
```

### Auth Users & Rate Limiting

**Supabase Auth Rate Limiting at Scale:**

Parallel tests hit Supabase auth rate limits causing cascading failures. Solutions:

1. **Reduce API calls** - Return email/password from setup (avoid getUserById calls)
2. **Retry with exponential backoff** - 1s, 2s, 4s delays
3. **Guard undefined context** - Check if setup succeeded before cleanup
4. **Promise.allSettled** - Both cleanups run even if one fails
5. **Rollback on partial failure** - Delete successfully created users if setup crashes

---

## 4. Architecture

### Server Actions (Extract Logic Functions)

**Never call async server actions directly in Vitest.** Extract business logic, test that instead.

```typescript
// actions.ts — thin wrapper, NOT unit tested
"use server";
export async function addTopLevelCategory(name: string) {
  await requireAdmin();
  const supabase = await createClient();
  const result = await addTopLevelCategoryLogic(supabase, name);
  if (result.success) revalidatePath("/admin/categories-review");
  return result;
}

// logic.ts — business logic, takes supabase as param
export async function addTopLevelCategoryLogic(
  supabase: SupabaseClient,
  name: string,
) {
  // all database operations here - TESTABLE
}

// test — zero mocks
const supabase = createClient(URL, SERVICE_ROLE_KEY);
const result = await addTopLevelCategoryLogic(supabase, testName("Category"));
expect(result.success).toBe(true);
```

**File Convention:**

- `actions.ts` — server action wrappers (auth + revalidate + delegate)
- `logic.ts` — exported business logic functions (testable, no framework deps)

### Mock Guidance

| Mock target             | Keep?      | Reason                             |
| ----------------------- | ---------- | ---------------------------------- |
| `@/lib/supabase/server` | **REMOVE** | Pass real client to logic function |
| `@/lib/dal`             | **REMOVE** | Auth tested via E2E                |
| `next/cache`            | **REMOVE** | Revalidation tested via E2E        |
| `ai` (Vercel AI SDK)    | **KEEP**   | External service boundary          |
| `@langfuse/*`           | **KEEP**   | External service boundary          |

**Principle:** Integration over mocks. Use real Supabase client in tests.

---

## 5. Debugging Quick Reference

### Common Symptoms → Fix

| Symptom                          | Likely Cause                        | Fix                                            |
| -------------------------------- | ----------------------------------- | ---------------------------------------------- |
| Passes alone, fails in suite     | Shared state / parallel race        | Worker-scoped fixtures                         |
| `Cannot read properties of null` | Data deleted by other worker        | Check CASCADE DELETE                           |
| Inconsistent failures in CI      | Rate limiting / resource contention | Add retries, reduce parallelism                |
| Works locally, fails CI          | Environment differences             | Check timeouts (30s), cold starts              |
| vi.clearAllMocks() contamination | Module-scoped mock state            | Remove clearAllMocks with module mocks         |
| Click does nothing (E2E)         | Hydration gap                       | Retry pattern or hydration marker              |
| E2E timeout on admin routes      | Missing auth fixture                | Import from auth-fixture, not @playwright/test |

### Parallel Execution Checklist

- ✅ Use worker-scoped fixtures for shared resources
- ✅ Test locally with forced parallelism: `--poolOptions.threads.singleThread=false`
- ✅ Check schema for CASCADE DELETE relationships
- ✅ Use unique identifiers in all test data (testName())
- ✅ Add retry logic for external service calls
- ✅ Guard cleanup against undefined context

### Flaky Test Checklist

- ✅ Reproduce: Does it pass alone but fail in suite?
- ✅ Timeouts: Are CI timeouts generous enough? (30s for external services)
- ✅ Hydration: Does E2E test wait for React to attach handlers?
- ✅ Mocks: Using `vi.clearAllMocks()` with module-level mocks?
- ✅ Cleanup: Using worker-scoped fixtures or manual beforeAll/afterAll?

---

## Related

- **Full diagnostics:** `/docs/avoiding-flakey-tests.md` (comprehensive anti-flakiness guide)
- **Deep debugging:** `/docs/test-debugging-deep-dive.md` (debugging decision tree, framework gotchas)
- **Migration guide:** `tests/MIGRATION-GUIDE-WORKER-FIXTURES.md` (legacy pattern migration)
