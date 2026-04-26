<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# integration

## Purpose
This directory contains subprocess-level integration tests for Yunxiao CLI. Tests build the CLI binary, run real commands with controlled environment variables, use local `httptest` servers, and verify stdout/stderr, exit codes, auth behavior, command introspection, and JSON envelopes.

## Key Files
| File | Description |
|------|-------------|
| `auth_test.go` | Tests interactive/non-interactive auth, token sources, config writes, verification, logout, and token secrecy. |
| `commands_test.go` | Tests `commands --json` and structured `--help --json` metadata. |
| `org_current_test.go` | Tests `org current` success and auth/error edge cases. |
| `codeup_test.go` | Tests Codeup repository commands and golden output. |
| `flow_pipeline_get_test.go` | Tests Flow pipeline success and error classifications. |
| `phase2_test.go` | Tests Phase 2 domains such as Projex, Packages, Testhub, and raw request behavior. |
| `config_test.go` | Tests configuration-related integration behavior. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Go integration tests only. |

## For AI Agents

### Working In This Directory
- Keep tests hermetic: strip inherited `YUNXIAO_*` env vars and set `YUNXIAO_CONFIG_FILE` to `t.TempDir()` where config state matters.
- Prefer `httptest.NewServer` over real Yunxiao network calls.
- Assert exit codes, stdout JSON, and stderr behavior for CLI contract changes.

### Testing Requirements
- Run focused tests with `go test ./test/integration -run <Pattern> -count=1` while iterating.
- Run `go test ./...` before completion when integration behavior changes.

### Common Patterns
- `buildTestBinary` compiles `./cmd/yunxiao` to a temporary test binary and cleans it up.
- `testEnv` removes inherited `YUNXIAO_*` variables before adding explicit overrides.
- Golden comparisons use `require.JSONEq` to avoid formatting coupling.

## Dependencies

### Internal
- `cmd/yunxiao/` binary and all internal command/domain packages are exercised through subprocess execution.
- `test/golden/` provides expected JSON outputs.

### External
- Go `testing`, `os/exec`, `httptest`, `github.com/stretchr/testify/require`, and `github.com/creack/pty`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
