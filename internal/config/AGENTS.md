<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# config

## Purpose
This package initializes and reads Yunxiao CLI configuration from environment variables, YAML config files, and built-in defaults. It centralizes API base URL, region, timeout, and organization ID fallback behavior.

## Key Files
| File | Description |
|------|-------------|
| `config.go` | Initializes Viper, reads `~/.yunxiao/config.yaml` or `YUNXIAO_CONFIG_FILE` directory, and exposes base URL, timeout, and organization ID helpers. |
| `config_test.go` | Unit tests for default OpenAPI endpoint and organization fallback behavior. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single configuration package. |

## For AI Agents

### Working In This Directory
- Preserve `https://openapi-rdc.aliyuncs.com` as the central default API base URL unless the contract changes.
- Remember that `YUNXIAO_CONFIG_FILE` currently points to a config directory used by `viper.AddConfigPath`, not a literal file path.
- Reset Viper in tests to avoid global state leakage.

### Testing Requirements
- Run `go test ./internal/config -count=1` after config logic changes.
- Run auth and org integration tests after base URL, token, or organization fallback changes.

### Common Patterns
- Environment variables override config values where helper functions explicitly check `os.Getenv`.
- `YUNXIAO_REGION_DEFAULT_ORG_ID` only applies when the base URL is treated as a region endpoint.

## Dependencies

### Internal
- Used by all command packages and auth verification.

### External
- `github.com/spf13/viper` for environment and config loading.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
