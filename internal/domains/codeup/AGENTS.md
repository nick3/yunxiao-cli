<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# codeup

## Purpose
This package implements Codeup repository API calls for listing repositories and fetching repository details.

## Key Files
| File | Description |
|------|-------------|
| `repositories.go` | Builds Codeup repository paths, sends GET requests, maps `nextPage` to `output.Pagination`, and returns repository data. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single domain package. |

## For AI Agents

### Working In This Directory
- Preserve central and region endpoint path differences.
- Escape repository and organization IDs with `url.PathEscape`.
- Keep pagination conversion consistent with other `nextPage` APIs.

### Testing Requirements
- Run Codeup integration tests after changing request or pagination behavior.
- Add unit tests if path construction gains new branches.

### Common Patterns
- Central path: `/oapi/v1/codeup/organizations/{org}/repositories`.
- Region path: `/oapi/v1/codeup/repositories`.

## Dependencies

### Internal
- `internal/domains/shared/` for JSON requests and token conversion.
- `internal/httpx/` and `internal/model/output/`.

### External
- Go `net/http`, `net/url`, and `strconv`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
