<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# validation

## Purpose
This package contains shared command flag validators used before network or auth work. It currently validates pagination page sizes.

## Key Files
| File | Description |
|------|-------------|
| `page.go` | Validates positive page sizes and returns `PARAM_INVALID` error details for invalid values. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single helper package. |

## For AI Agents

### Working In This Directory
- Keep validation errors structured as `output.ErrorDetail` so command packages can preserve JSON envelopes.
- Validate parameter errors before auth/network where the public contract expects `PARAM_INVALID` first.

### Testing Requirements
- Run affected integration tests for list commands after validation changes.
- Run `go test ./...` if adding shared validators used across domains.

### Common Patterns
- Category `param` maps to exit code 2 through command rendering.

## Dependencies

### Internal
- `internal/model/output/` for error details.

### External
- No third-party dependencies.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
