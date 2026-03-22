# TDD Plan: Merge claude-code-config into Rocket Fuel plugin

## Context
All personal Claude Code skills, agents, and rules currently live in ~/.claude/ (unversioned). Moving them into the Rocket Fuel plugin makes them version-controlled, portable (go install + rf launch), and available in all Claude sessions via the plugin system.

## Architecture
Copy files from ~/.claude/skills/, ~/.claude/agents/, ~/.claude/rules/ into internal/plugin/. The existing go:embed + extractDir pattern already handles recursive directories. Only rules/ needs a new embed directive. After verification, personal copies are deleted — plugin is single source of truth.

## Session Constants
Test command: `go test -race ./...`
Test file pattern: colocated `*_test.go`
Test helpers: `internal/testutil/`
Acceptance test path: none

## Slice 1: Copy all agents into plugin + verify count
Type: unit | Status: done
Files: `internal/plugin/agents/*.md` (15 new files), `internal/plugin/extract_test.go`

- [x] ExtractPlugin extracts all 17 agent definitions (15 new + 2 existing)
- [x] Each agent file has valid content (non-empty, starts with --- or #)

## Slice 2: Copy all skills into plugin + verify complex extraction
Type: unit | Status: pending
Files: `internal/plugin/skills/*/` (23 new skill directories), `internal/plugin/extract_test.go`
Builds on: Slice 1

- [ ] ExtractPlugin extracts all 24 skill directories (23 new + board-setup)
- [ ] Skills with references/ subdirectories extract correctly (verify tdd/references/ exists)

## Slice 3: Add rules support
Type: unit | Status: pending
Files: `internal/plugin/rules/*.md` (8 new), `internal/plugin/extract.go`, `internal/plugin/extract_test.go`, `CLAUDE.md`
Builds on: Slice 2

- [ ] extract.go embeds and extracts rules/ directory
- [ ] All 8 rule files present after extraction
- [ ] CLAUDE.md updated with rules in plugin architecture table
