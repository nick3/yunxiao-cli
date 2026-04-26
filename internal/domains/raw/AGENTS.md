<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# raw

## Purpose
This package validates and executes read-only raw Yunxiao OpenAPI requests for endpoints not yet wrapped by typed commands.

## Key Files
| File | Description |
|------|-------------|
| `request.go` | Validates Phase 2 raw request boundaries and executes GET requests through shared JSON helpers. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single domain package. |

## For AI Agents

### Working In This Directory
- Preserve GET-only and `/oapi/`-only validation unless the public safety boundary changes.
- Keep validation reusable so command tests and domain tests can exercise the same rules.

### Testing Requirements
- Run raw/Phase 2 integration tests after changes.
- Confirm invalid methods and invalid paths return `PARAM_INVALID`.

### Common Patterns
- `Request` re-validates inputs before calling `shared.RequestJSON`.

## Dependencies

### Internal
- `internal/domains/shared/`, `internal/httpx/`, and `internal/model/output/`.

### External
- Go `net/http` and `strings`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
