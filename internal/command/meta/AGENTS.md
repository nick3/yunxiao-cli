<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# meta

## Purpose
This package implements command introspection for agents and automation. It provides `yunxiao commands --json` and JSON-formatted command help so callers can discover command paths, flags, flag types, and required metadata without guessing.

## Key Files
| File | Description |
|------|-------------|
| `commands.go` | Builds recursive command specs, installs JSON help behavior, and serializes flag metadata. |
| `commands_test.go` | Unit tests for command spec construction and metadata behavior. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single command package. |

## For AI Agents

### Working In This Directory
- Preserve recursive `subcommands` output; AI consumers depend on walking the command tree.
- Keep required flag annotations visible in JSON help.
- Do not include Cobra help/completion internals as public command entries.

### Testing Requirements
- Run `go test ./internal/command/meta -count=1` after metadata logic changes.
- Run `go test ./test/integration -run TestCommands -count=1` after public introspection behavior changes.

### Common Patterns
- `BuildSpec` recursively includes available subcommands and merges inherited/non-inherited flags.
- `InstallJSONHelp` switches help output to JSON when `--json` is present.

## Dependencies

### Internal
- `internal/cli/` for output rendering.
- `internal/command/flagmeta/` for required flag annotations.

### External
- Cobra and pflag for command and flag inspection.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
