<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# org

## Purpose
This package implements organization/user context commands. Currently it provides `yunxiao org current`, which returns the current user and organization by calling Yunxiao's user endpoint.

## Key Files
| File | Description |
|------|-------------|
| `current.go` | Defines `org current`, resolves auth/config context, calls `/oapi/v1/platform/user`, and renders the response envelope. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single command package. |

## For AI Agents

### Working In This Directory
- Keep `org current` aligned with auth verification endpoint behavior.
- Isolate integration tests from the user's real config token when changing auth or config behavior.
- Do not require `--organization-id` for `org current`; it discovers current context from the token.

### Testing Requirements
- Run `go test ./test/integration -run TestOrgCurrent -count=1` after changes.
- Run auth tests too when token source or verification behavior changes.

### Common Patterns
- Uses the standard command-level `newAPIClient` flow for trace ID, timeout, retry, quiet mode, token, and base URL.

## Dependencies

### Internal
- `internal/domains/shared/` for JSON request helpers.
- `internal/auth/`, `internal/config/`, `internal/httpx/`, and `internal/model/output/`.

### External
- Cobra for command definitions.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
