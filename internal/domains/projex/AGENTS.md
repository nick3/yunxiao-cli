<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# projex

## Purpose
This package implements Projex API calls for projects, workitems, and sprints. It handles both search-style POST endpoints and GET endpoints with query/header pagination.

## Key Files
| File | Description |
|------|-------------|
| `projects.go` | Implements project/workitem/sprint list and get calls, search payloads, central-vs-region paths, and pagination mapping. |
| `paths_test.go` | Unit tests for Projex endpoint path selection. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single domain package. |

## For AI Agents

### Working In This Directory
- Maintain central-vs-region path tests when changing endpoint construction.
- Use `shared.ApplyPageToken` for search payload page tokens.
- Keep workitem list payload keys (`category`, `spaceId`, `perPage`) aligned with Yunxiao API expectations.

### Testing Requirements
- Run `go test ./internal/domains/projex -count=1` after path changes.
- Run Phase 2 integration tests after behavior changes.

### Common Patterns
- Projects and workitems use `:search` POST responses with `nextPage`.
- Sprints use GET query parameters and `x-next-page` header pagination.

## Dependencies

### Internal
- `internal/domains/shared/`, `internal/httpx/`, and `internal/model/output/`.

### External
- Go `net/http`, `net/url`, and `strconv`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
