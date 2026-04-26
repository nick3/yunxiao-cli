<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# raw

## Purpose
This package implements the `yunxiao raw request` escape hatch for read-only Yunxiao OpenAPI coverage gaps. It validates raw request boundaries before dispatching to the raw domain layer.

## Key Files
| File | Description |
|------|-------------|
| `request.go` | Defines `raw request`, validates `--method` and `--path`, creates the API client, and renders raw response data. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single command package. |

## For AI Agents

### Working In This Directory
- Preserve Phase 2 safety: only `GET` and paths beginning with `/oapi/` are allowed.
- Do not allow absolute URLs or mutating methods unless the raw request contract is explicitly changed.
- Keep parameter validation before network calls.

### Testing Requirements
- Run `go test ./test/integration -run TestRaw -count=1` or the relevant Phase 2 integration tests after raw changes.
- Confirm invalid methods return `PARAM_INVALID` with exit code 2.

### Common Patterns
- Param validation errors are returned as JSON envelopes and should not emit unnecessary stderr noise.

## Dependencies

### Internal
- `internal/domains/raw/` for request validation and execution.
- `internal/auth/`, `internal/config/`, `internal/httpx/`, and `internal/model/output/`.

### External
- Cobra for command definitions.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
