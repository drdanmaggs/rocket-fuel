# Code Quality Standards

**CRITICAL:** These rules are non-negotiable for all TypeScript code.

## Type Safety

### No `any` Types
- `any` types are **banned** without exception
- Never use `as any` or `@ts-ignore` to silence type warnings
- Type warnings catch real bugs — fix the underlying type instead

### Use `unknown` for Untyped Data
- When a proper type genuinely isn't possible, use `unknown` with runtime validation
- Always validate `unknown` at the boundary (parsing, API responses, etc.)
- Example:
  ```typescript
  function processData(data: unknown) {
    if (typeof data === 'string') {
      // Now data is narrowed to string
      return data.toUpperCase();
    }
    throw new Error('Invalid data type');
  }
  ```

### When Type Issues Arise
1. Try to infer the correct type from usage
2. If external (API, library), check docs or use `unknown` + validation
3. If genuinely stuck, ask before using escape hatches

## ESLint Configuration

### Required Rule
```json
{
  "@typescript-eslint/no-explicit-any": "error"
}
```

### Inline Disables
Permitted **only as a last resort**:
- Must be line-level (never file-level)
- Must include a comment explaining:
  - **Why** the disable is needed
  - **Whether** it's fixable in future

Example:
```typescript
// eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Legacy API returns untyped data, tracked in issue #123
const result = await legacyApi.fetch();
```
