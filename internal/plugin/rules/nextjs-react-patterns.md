# Next.js & React Patterns

## Next.js 16: `middleware.ts` → `proxy.ts`

**Next.js 16 renamed `middleware.ts` to `proxy.ts`** (Node.js runtime, not Edge). Export `proxy` not `middleware`. Codemod: `npx @next/codemod@latest upgrade`. Old `middleware.ts` is deprecated.

## Syncing Props to State (Client Components)

### ❌ Anti-Patterns

```typescript
// BAD: useState only uses initial value on mount
const [data, setData] = useState(initialData);

// BAD: useEffect causes double render (stale → updated)
useEffect(() => setData(initialData), [initialData]);
```

### ✅ Correct: setState During Render (DEFAULT)

```typescript
const [data, setData] = useState(initialData);
const [prevData, setPrevData] = useState(initialData);

if (initialData !== prevData) {
  setPrevData(initialData);
  setData(initialData);  // Single render cycle
}
```

**Benefits:** Single render, no stale state, efficient.

**Source:** [React docs - You Might Not Need an Effect](https://react.dev/learn/you-might-not-need-an-effect#adjusting-some-state-when-a-prop-changes)

### ⚠️ Exception: useInsertionEffect Libraries

**MUST use `useEffect` when using:**
- Drag-and-drop (@dnd-kit/react)
- CSS-in-JS (styled-components, emotion)
- Any library that injects styles at runtime

**Why:** React 19 crashes with `"useInsertionEffect must not schedule updates"` if setState-during-render is used.

```typescript
// ✅ Safe with useInsertionEffect libraries
useEffect(() => setData(initialData), [initialData]);
```

**Trade-off:** One extra render vs. app crash.

### Alternative: Key Prop

Reset **ALL** state when prop changes:

```typescript
<Profile userId={userId} key={userId} />  // Forces complete remount
```

**Use when:** All state should reset together.
**Don't use when:** Only specific state needs syncing.

## router.refresh() with Client State

```typescript
router.refresh()  // Re-renders Server Components, preserves client useState
```

**Common pitfall:** Client state doesn't sync with fresh server data.

```typescript
// Server Component passes fresh data
<ClientComponent initialData={serverData} />

// Client Component
const [data, setData] = useState(initialData);  // Doesn't sync on refresh!
```

**Fix:** Use setState-during-render pattern above.

## Quick Reference

| Pattern | Use Case |
|---------|----------|
| **setState during render** | Default - sync props to state |
| **useEffect** | REQUIRED with useInsertionEffect libraries |
| **key prop** | Reset ALL component state |
| **Direct prop** | No state needed (computed values) |
