<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# testhub

## Purpose
This package implements Testhub commands for testcases, testcase details, directories, and testplans. It handles paginated testcase lists while keeping `testplans list` intentionally non-paginated.

## Key Files
| File | Description |
|------|-------------|
| `testcases.go` | Defines Testhub command tree, testcase list/get, directory list, testplan list, flag validation, and API client setup. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single command package. |

## For AI Agents

### Working In This Directory
- Do not add `--page-size` or `--page-token` to `testhub testplans list` unless the public contract changes and tests/docs are updated.
- Keep required `test-repo-id` metadata for testcase and directory commands.
- Keep API path logic in `internal/domains/testhub/`.

### Testing Requirements
- Run Phase 2 integration tests and command help tests after Testhub command changes.
- Run path tests in `internal/domains/testhub` when endpoint construction changes.

### Common Patterns
- Testcase list is paginated; directories and testplans currently return arrays without pagination metadata.

## Dependencies

### Internal
- `internal/domains/testhub/` for API calls and paths.
- `internal/command/validation/` for testcase pagination.
- `internal/command/flagmeta/` for required flags.

### External
- Cobra for command definitions.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
