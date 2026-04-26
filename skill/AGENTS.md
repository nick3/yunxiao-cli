<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# skill

## Purpose
This directory contains Agent skill documentation that teaches AI agents how to call Yunxiao CLI safely and correctly from shells, CI, subprocesses, and automation.

## Key Files
| File | Description |
|------|-------------|
| _(none)_ | Skill content is organized into subdirectories. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `using-yunxiao-cli/` | Skill for machine-safe Yunxiao CLI usage (see `using-yunxiao-cli/AGENTS.md`). |

## For AI Agents

### Working In This Directory
- Skill files are product documentation for other agents; keep examples exact and verified against the real CLI.
- Do not guess command names, flags, pagination, or error categories; verify via `yunxiao commands --json` and `--help --json`.
- Avoid examples that expose tokens in command-line arguments.

### Testing Requirements
- Validate skill guidance by running representative CLI commands or pressure-testing with an agent before claiming the skill is ready.
- For behavior changes, update the skill alongside README, ADR, and integration tests.

### Common Patterns
- Skill descriptions should state when to use the skill, not summarize its full workflow.
- Agent-facing CLI examples should use `--json` and parse stdout JSON envelopes.

## Dependencies

### Internal
- `cmd/yunxiao/` and `internal/command/meta/` provide command introspection used by the skill.
- `docs/adr/001-cli-contract.md` defines the contract that the skill explains.
- `test/integration/` provides executable examples of CLI behavior.

### External
- No runtime dependencies; examples assume shell and `jq` for parsing.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
