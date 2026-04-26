<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# flow

## Purpose
This package implements Flow pipeline API calls for listing pipelines and fetching pipeline details.

## Key Files
| File | Description |
|------|-------------|
| `pipelines.go` | Builds pipeline paths, sends GET requests, converts `nextPage` to pagination metadata, and returns pipeline data. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single domain package. |

## For AI Agents

### Working In This Directory
- Keep central and region path behavior aligned with tests and ADR examples.
- Escape organization and pipeline IDs.
- Preserve `perPage` and `page` query names used by Yunxiao Flow APIs.

### Testing Requirements
- Run Flow integration tests after path, query, or pagination changes.

### Common Patterns
- List returns `data` plus `nextPage`; get returns a single map.

## Dependencies

### Internal
- `internal/domains/shared/`, `internal/httpx/`, and `internal/model/output/`.

### External
- Go `net/http`, `net/url`, and `strconv`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
