# User Feedback Flywheel

User feedback is the highest-leverage product input. This philosophy applies to every project.

## The Flywheel

Prominent feedback UI → low-friction submission → fast triage → visible action → user feels heard → submits more feedback → product improves faster.

**Every step matters.** A buried feedback link kills the loop. Slow triage kills trust. No visible response kills motivation.

## Design Principles

1. **Feedback UI is always visible** — persistent button in main layout, not hidden in settings. Users report bugs in the moment or not at all.
2. **Submission is instant** — minimal required fields. Category + description. Everything else (page URL, user context, browser info) is captured automatically.
3. **Triage is fast** — admin sees feedback within hours, not days. Build for quick scan and action.
4. **Close the loop** — users should know their feedback was received and acted on. Even "won't fix" is better than silence.
5. **Lower the bar** — every piece of feedback (bug, suggestion, confusion, praise) is valuable. Don't gate behind formal templates.

## Implementation Implications

- Feedback button placement: navbar/FAB, not buried in menus
- Store with full context (user, household, current page, timestamp)
- Admin triage view is a launch requirement, not a follow-up
- Toast confirmation on submit ("Thanks! We'll look at this shortly")
- Status tracking so users can see their feedback was acknowledged
