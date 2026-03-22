# ADR-006: Hybrid Plugin Architecture

## Status: Active

## Context

Rocket Fuel's AI intelligence — the Integrator prompt, Worker prompt, hook declarations, and skills — was embedded in the Go binary via `go:embed` and programmatically written to `.claude/settings.json` via `EnsureClaudeSettings`. This had several problems:

1. **EnsureClaudeSettings was a hack** — programmatically writing JSON into project repos is fragile and pollutes repos with Rocket Fuel configuration
2. **Prompt iteration was slow** — changing the Integrator personality required rebuilding the Go binary
3. **Reinventing Claude Code features** — skill routing, agent definitions, and tool restrictions are all native Claude Code capabilities we weren't using
4. **No agent isolation** — Workers and Integrator ran as generic Claude sessions with injected prompts, without tool restrictions or model overrides

Meanwhile, Claude Code provides a native plugin system with agents, skills, hooks, and a distribution model.

## Alternatives Considered

### Option A: Pure Go Binary (status quo)
Everything embedded in Go. Simple but limited — can't use native Claude Code agent features.

### Option B: Pure Claude Code Plugin
Everything in a plugin, minimal Go binary. Two install steps, version sync nightmares, still need Go for tmux/worktree orchestration.

### Option C: Hybrid (chosen)
Go binary embeds the plugin via `go:embed` and extracts it on launch. Single artifact, version sync guaranteed, native Claude Code integration for what it's good at.

## Decision

**Hybrid architecture with scope split:**

| Component | Where | Why |
|-----------|-------|-----|
| Agent definitions | Plugin (`~/.claude/plugins/rocket-fuel/`) | Native agent features: tools, model, permissions |
| Skills | Plugin | Auto-invocation, YAML frontmatter, composability |
| Hooks | Project `.claude/settings.json` | **Must be project-scoped** — Stop hook would block all sessions if global |
| Hook handlers | Go binary (`rf` commands) | Mechanical operations, testable |
| Orchestration | Go binary | tmux, worktrees, watchdog, dashboard, self-update |

### Critical: Hooks Stay Project-Scoped

The Stop hook (`rf should-continue`) forces the Integrator to keep working when there's queued work. If this were in a global plugin, it would block Claude from stopping in ALL sessions — including non-Rocket Fuel ones. Hooks remain in `.claude/settings.json`, installed per-project by `rf launch` via `EnsureClaudeSettings`.

### Plugin Extraction

- Plugin source files live in `internal/plugin/` (agents/, skills/, .claude-plugin/)
- Embedded via `//go:embed` directives in `internal/plugin/extract.go`
- `ExtractPlugin(targetDir)` copies to `~/.claude/plugins/rocket-fuel/` on every `rf launch`
- Always-overwrite strategy — fork the repo to customize
- Version sync guaranteed (plugin compiled into binary)

### Agent Invocations

- Integrator: `claude --agent integrator --dangerously-skip-permissions "<startup message>"`
- Worker: `claude --agent worker --dangerously-skip-permissions "<issue context>"`
- Dynamic context (issue number, skill, board state) passed via initial message
- Static personality and instructions live in the agent definition

## Consequences

- Single `go install` — no separate plugin installation
- Prompt iteration is instant (edit markdown, restart session — or rebuild for permanent change)
- Native Claude Code agent features available (tool restrictions, model overrides in future)
- `EnsureClaudeSettings` stays (simplified — hooks only)
- `prompts/` directory deleted — single source of truth is `internal/plugin/agents/`
- `internal/prime/integrator.md` deleted — prompt lives in plugin agent definition
- `prime.Build()` simplified to dynamic state only (board, workers, repo)
