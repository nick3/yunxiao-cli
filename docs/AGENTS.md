<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# docs

## Purpose
This directory contains project documentation beyond the README, including architecture decisions and phase status notes. It defines the intended CLI contract that implementation, tests, and skill documentation must follow.

## Key Files
| File | Description |
|------|-------------|
| `phase0-status.md` | Status notes for the initial project phase. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `adr/` | Architecture Decision Records for public CLI behavior (see `adr/AGENTS.md`). |

## For AI Agents

### Working In This Directory
- Treat ADRs as source-of-truth for public contract decisions unless superseded by explicit user direction.
- Keep docs synchronized with implementation, README, integration tests, and `skill/using-yunxiao-cli/SKILL.md`.
- Do not introduce aspirational behavior unless it is implemented or clearly marked as future work.

### Testing Requirements
- For contract documentation changes, verify matching Go tests or golden files exist.
- Run markdown-only checks if the project later adds a doc linter; currently no doc-specific test exists.

### Common Patterns
- ADR files use numbered names and describe stable external behavior such as exit codes, JSON envelope, flags, pagination, and retry.

## Dependencies

### Internal
- `internal/` implementation must conform to contract docs.
- `test/integration/` validates contract examples and edge cases.
- `skill/using-yunxiao-cli/` translates contract details into Agent-facing usage guidance.

### External
- No runtime dependencies.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
