# Form Field Removal Testing

When a field is removed from a form (moved to another form, deleted), two
tests are always required. One without the other is incomplete.

## Required test pair

### Test 1 — DOM absence (field does not render)
```typescript
expect(screen.queryByLabelText(/field name/i)).not.toBeInTheDocument();
```

### Test 2 — Submit payload preservation (data still sent to action)
```typescript
await user.click(screen.getByRole("button", { name: /save/i }));
const callArgs = mockAction.mock.calls[0][0];
expect(callArgs.preferences.removedField).toBe(existingValue);
```

## Why

Verifying a field doesn't render ≠ verifying its data is preserved.
When the underlying action does a full object replacement (not partial
update), omitting the field in the submit payload silently deletes it
from the database.

## Trigger

Any plan slice containing "preserve", "maintain", "keep", "not lose",
or "still include" requires a submit payload assertion, not just a DOM
assertion.
