<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# codeup

## Purpose
This package implements Codeup repository CLI commands: `yunxiao codeup repos list` and `yunxiao codeup repo get`. It validates organization and repository flags, handles pagination flags for list operations, and delegates API work to the Codeup domain package.

## Key Files
| File | Description |
|------|-------------|
| `repos.go` | Defines Codeup command tree, list/get flag handling, API client creation, pagination metadata, and error rendering. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single command package. |

## For AI Agents

### Working In This Directory
- Keep required flag metadata in sync with actual validation.
- Validate `--page-size` before making auth or network calls.
- Do not put URL path construction here; keep it in `internal/domains/codeup/`.

### Testing Requirements
- Run `go test ./test/integration -run TestCodeup -count=1` after Codeup command changes.
- Run `go test ./test/integration -run TestCommands -count=1` after changing flags or command hierarchy.

### Common Patterns
- `newAPIClient` resolves trace ID, timeout, retry, quiet mode, token, and base URL before domain calls.
- List commands attach `output.Pagination` to `meta.Pagination`.

## Dependencies

### Internal
- `internal/domains/codeup/` for Codeup API paths and response mapping.
- `internal/command/validation/` for pagination validation.
- `internal/command/flagmeta/` for required flag metadata.

### External
- Cobra for commands and flags.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
