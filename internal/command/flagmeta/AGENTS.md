<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# flagmeta

## Purpose
This helper package marks Cobra flags as required in a way that command introspection can expose to agents and automation.

## Key Files
| File | Description |
|------|-------------|
| `required.go` | Defines the required-flag annotation key and `MustMarkRequired` helper. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single helper package. |

## For AI Agents

### Working In This Directory
- Keep the annotation stable because `internal/command/meta/` reads it for JSON command specs.
- Use this helper in command packages whenever a flag is required by runtime validation.

### Testing Requirements
- Run `go test ./internal/command/meta -count=1` after annotation behavior changes.
- Run command integration tests after changing required flag metadata.

### Common Patterns
- `MustMarkRequired` panics on programmer error so command setup fails fast during tests.

## Dependencies

### Internal
- `internal/command/meta/` consumes the annotation.

### External
- Cobra command flags.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
