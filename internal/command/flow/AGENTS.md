<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# flow

## Purpose
This package implements Flow pipeline CLI commands: `yunxiao flow pipelines list` and `yunxiao flow pipeline get`. It adapts Cobra flags into domain calls and preserves pagination metadata for list responses.

## Key Files
| File | Description |
|------|-------------|
| `pipeline.go` | Defines Flow command tree, list/get command behavior, required flags, pagination validation, API client creation, and error handling. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single command package. |

## For AI Agents

### Working In This Directory
- Keep plural `pipelines list` and singular `pipeline get` command grammar intact.
- Validate organization and pipeline IDs before network calls.
- Keep pipeline API path logic in `internal/domains/flow/`.

### Testing Requirements
- Run `go test ./test/integration -run TestFlow -count=1` after Flow command changes.
- Run command introspection tests after flag or command path changes.

### Common Patterns
- List commands expose `--page-size` and `--page-token`; get commands require an ID and do not paginate.

## Dependencies

### Internal
- `internal/domains/flow/` for API interaction.
- `internal/command/validation/` and `internal/command/flagmeta/` for flag behavior.

### External
- Cobra for command definitions.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
