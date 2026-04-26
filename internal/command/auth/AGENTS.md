<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# auth

## Purpose
This package implements `yunxiao auth`, `yunxiao auth login`, `yunxiao auth status`, and `yunxiao auth logout`. It supports visible interactive token entry for humans, stdin-based token configuration for automation, token verification, and config token removal.

## Key Files
| File | Description |
|------|-------------|
| `auth.go` | Defines auth subcommands, token reading/normalization, verification via `/oapi/v1/platform/user`, config file writes, status detection, and logout behavior. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single command package. |

## For AI Agents

### Working In This Directory
- Never print token values to stdout/stderr except the intentional visible interactive terminal echo covered by integration tests.
- Keep non-interactive usage explicit: scripts and CI should use `auth login --token-stdin` or `YUNXIAO_ACCESS_TOKEN`.
- Preserve `--force` semantics so existing config tokens are not overwritten accidentally.

### Testing Requirements
- Run `go test ./test/integration -run TestAuth -count=1` after auth behavior changes.
- Run `go test ./test/integration -run TestCommands -count=1` after auth flags or command help changes.

### Common Patterns
- `verifyToken` uses `config.GetBaseURL()` and GET `/oapi/v1/platform/user`.
- Config writes preserve YAML map contents and set restrictive permissions where possible.

## Dependencies

### Internal
- `internal/cli/` for output and exit handling.
- `internal/config/` for config path, timeout, and base URL.
- `internal/domains/shared/` and `internal/httpx/` for token verification.

### External
- Cobra for command definitions.
- Viper and YAML libraries for config token storage.
- `golang.org/x/term` for terminal detection.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
