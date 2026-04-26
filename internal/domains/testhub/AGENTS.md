<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# testhub

## Purpose
This package implements Testhub API calls for testcases, testcase details, directories, and testplans. It covers search-style testcase pagination and non-paginated testplan lists.

## Key Files
| File | Description |
|------|-------------|
| `testcases.go` | Implements testcase list/get, directory list, testplan list, central-vs-region paths, and pagination mapping. |
| `paths_test.go` | Unit tests for Testhub endpoint path selection. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single domain package. |

## For AI Agents

### Working In This Directory
- Keep `testplansPath` behavior aligned with the non-paginated CLI command contract.
- Escape organization, test repo, and testcase IDs in paths.
- Preserve testcase search POST response handling with `nextPage`.

### Testing Requirements
- Run `go test ./internal/domains/testhub -count=1` after path changes.
- Run Phase 2 integration and command help tests after Testhub behavior changes.

### Common Patterns
- Testcase lists use `:search` and return pagination.
- Directory and testplan commands do not expose CLI pagination metadata.

## Dependencies

### Internal
- `internal/domains/shared/`, `internal/httpx/`, and `internal/model/output/`.

### External
- Go `net/http` and `net/url`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
