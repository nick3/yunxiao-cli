<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# packages

## Purpose
This package implements package repository and artifact API calls, including repository lists, artifact lists, artifact detail lookup, optional filters, and header-based pagination.

## Key Files
| File | Description |
|------|-------------|
| `packages.go` | Builds package repository/artifact paths, encodes query filters, reads `x-next-page`, and returns package/artifact data. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single domain package. |

## For AI Agents

### Working In This Directory
- Preserve `repoType`, ordering, search, and repository filter query parameters.
- Escape path identifiers and keep query values in `url.Values`.
- Keep `x-next-page` pagination semantics consistent with package APIs.

### Testing Requirements
- Run Phase 2 integration tests after package API changes.

### Common Patterns
- Package APIs return arrays directly and expose next page through headers, not `nextPage` in JSON body.

## Dependencies

### Internal
- `internal/domains/shared/`, `internal/httpx/`, and `internal/model/output/`.

### External
- Go `net/http`, `net/url`, and `strconv`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
