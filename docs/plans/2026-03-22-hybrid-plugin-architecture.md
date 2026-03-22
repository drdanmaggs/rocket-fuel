# TDD Plan: Hybrid Plugin Architecture

## Context
Rocket Fuel's AI intelligence (Integrator/Worker prompts, skills) is currently embedded in the Go binary via go:embed and dynamic prompt building. Migrating agent definitions and skills to a proper Claude Code plugin — embedded in the Go binary and extracted on launch — gives native integration (tool restrictions, model overrides, auto-invocation) while preserving single-artifact distribution.

## Architecture
The Go binary embeds a `plugin/` directory containing Claude Code plugin files (agents, skills). On `rf launch`, `internal/plugin.ExtractPlugin(targetDir)` copies these files to `~/.claude/plugins/rocket-fuel/`. Claude Code natively loads the plugin on session start.

**Critical scope decision:** Hooks MUST stay project-scoped (in `.claude/settings.json` per repo, written by `EnsureClaudeSettings`). The Stop hook (`rf should-continue`) would block Claude from stopping in ALL sessions if applied globally via plugin. `EnsureClaudeSettings` stays but is simplified — only responsible for hooks, not agents.

| Component | Where | Why |
|-----------|-------|-----|
| Agent definitions | Plugin (user-scoped) | Static personality, tools, model config |
| Skills | Plugin (user-scoped) | Board-setup, dispatch workflows |
| Hooks | Project `.claude/settings.json` | Must be project-scoped (Stop hook dangerous globally) |
| Hook handlers | Go binary (rf commands) | Mechanical operations, unchanged |

Always-overwrite strategy — no local customization, fork the repo to modify. Two ADRs document decisions: ADR-006 (plugin architecture) and ADR-007 (dedicated board per repo).

## Session Constants
Test command: `go test -race ./...`
Test file pattern: colocated `*_test.go`
Test helpers: `internal/testutil/` (real tmux sockets, real git repos)
Acceptance test path: none (CLI tool — manual verification required)

## Slice 1: Plugin skeleton + extraction
Type: unit | Status: done
Files: `internal/plugin/.claude-plugin/plugin.json` (new), `internal/plugin/extract.go` (new), `internal/plugin/extract_test.go` (new), `cmd/up.go`

- [x] ExtractPlugin(targetDir) creates plugin directory structure at the given path
- [x] Extraction creates .claude-plugin/plugin.json with valid manifest (name, version, description fields)
- [x] Extraction overwrites existing files on every call (always-overwrite strategy)
- [x] ExtractPlugin returns error if target directory is not writable
- [x] rf launch calls ExtractPlugin before session setup (warning on failure, not fatal)

## Slice 2: Integrator agent definition
Type: unit | Status: done
Files: `plugin/agents/integrator.md` (new), `internal/launch/launch.go`, `internal/prime/prime.go`, `internal/prime/prime_test.go`
Builds on: Slice 1

- [x] Integrator agent file has valid YAML frontmatter (name, description, tools)
- [x] Integrator prompt content preserved from current internal/prime/integrator.md
- [x] IntegratorCommand returns claude invocation referencing the plugin agent
- [x] prime.Build no longer includes static integrator prompt (only dynamic state: board, workers, repo)
- [x] go:embed of integrator.md removed from prime.go
- [x] internal/prime/integrator.md deleted (single source of truth is plugin/agents/integrator.md)

## Slice 3: Worker agent definition
Type: unit | Status: pending
Files: `plugin/agents/worker.md` (new), `internal/worker/worker.go`, `internal/worker/worker_test.go`
Builds on: Slice 2

- [ ] Worker agent file has valid YAML frontmatter (name, description, tools)
- [ ] Worker spawn command references plugin agent with issue context via initial message
- [ ] Skill routing preserved in worker.go (labels to skills mapping in initial message)
- [ ] prompts/worker.md deleted (single source of truth is plugin/agents/worker.md)

## Slice 4: Board-setup skill
Type: unit | Status: pending
Files: `plugin/skills/board-setup/SKILL.md` (new)
Builds on: Slice 1 (not Slice 3 — independent)

- [ ] Skill file has valid YAML frontmatter (name, description)
- [ ] Skill content instructs creating board with columns: Backlog, Ready, Scoped, In Progress, In Review, Done
- [ ] Skill is extracted as part of plugin

## Slice 5: ADRs + CLAUDE.md update
Type: docs | Status: pending
Files: `docs/adr/006-hybrid-plugin-architecture.md` (new), `docs/adr/007-dedicated-board-per-repo.md` (new), `CLAUDE.md`
Builds on: nothing (parallel)

- [ ] ADR-006 documents hybrid plugin decision: context, alternatives (pure Go / pure plugin / hybrid), decision, scope split (agents+skills in plugin, hooks stay project-scoped), migration path
- [ ] ADR-007 documents board-per-repo decision: context, column structure, relationship to #68
- [ ] CLAUDE.md updated with plugin structure, agent references, new architecture
