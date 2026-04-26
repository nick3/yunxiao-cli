<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# using-yunxiao-cli

## Purpose
This directory contains the reusable Agent skill for calling Yunxiao CLI from shell, CI, subprocesses, and automation. The skill teaches command discovery, JSON parsing, pagination, auth, retry, timeout, raw request, and error-category handling.

## Key Files
| File | Description |
|------|-------------|
| `SKILL.md` | Agent-facing usage guide for stable Yunxiao CLI automation, including examples and common mistakes. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single skill package. |

## For AI Agents

### Working In This Directory
- Verify examples against the real CLI before changing the skill.
- Keep the skill focused on when/how agents should call `yunxiao`; avoid project implementation details that belong in README or ADRs.
- Do not add unsafe auth examples that place tokens in command-line arguments.

### Testing Requirements
- Run representative `./yunxiao commands --json` and `./yunxiao <cmd> --help --json` checks when command discovery guidance changes.
- Pressure-test changes with an agent or at minimum validate examples manually before claiming the skill is reliable.

### Common Patterns
- The skill requires agents to discover commands from the CLI instead of guessing aliases or flags.
- Failure handling examples capture stdout before acting on non-zero exit codes.

## Dependencies

### Internal
- `cmd/yunxiao/` builds the CLI whose behavior the skill documents.
- `internal/command/meta/` supplies command introspection.
- `docs/adr/001-cli-contract.md` defines the contract explained by the skill.

### External
- Shell and `jq` appear in examples for JSON parsing.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
