<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# auth

## Purpose
This package resolves the access token used by business commands. It enforces token source precedence so automation can use `YUNXIAO_ACCESS_TOKEN` while human login can persist a config token.

## Key Files
| File | Description |
|------|-------------|
| `token.go` | Returns the access token, preferring `YUNXIAO_ACCESS_TOKEN` over Viper config `access_token`, and reports auth failure when neither exists. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single-purpose package. |

## For AI Agents

### Working In This Directory
- Do not log, print, or wrap token values into errors.
- Preserve environment-token priority over config-token priority unless the public contract changes.
- Keep missing-token errors actionable for CLI users and CI.

### Testing Requirements
- Run `go test ./test/integration -run 'TestAuthStatus|TestOrgCurrentPrefersEnvTokenOverConfigToken' -count=1` after token precedence changes.
- Run `go test ./...` before claiming project-level correctness.

### Common Patterns
- Token lookup is side-effect free; writing/removing tokens belongs to `internal/command/auth/`.

## Dependencies

### Internal
- `internal/config/` initializes Viper before command execution.

### External
- `github.com/spf13/viper` for config-backed token lookup.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
