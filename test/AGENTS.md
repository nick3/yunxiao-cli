<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# test

## Purpose
This directory contains black-box and fixture-based tests for the public CLI contract. Integration tests build and execute the `yunxiao` binary, compare JSON envelopes against golden files, and validate command behavior from an automation consumer's perspective.

## Key Files
| File | Description |
|------|-------------|
| _(none)_ | Test assets are organized into subdirectories. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `golden/` | Expected JSON envelope outputs for integration tests (see `golden/AGENTS.md`). |
| `integration/` | Subprocess-level CLI integration tests (see `integration/AGENTS.md`). |

## For AI Agents

### Working In This Directory
- Treat golden files as public contract fixtures; update them only when behavior intentionally changes.
- Keep integration tests isolated from the user's real `~/.yunxiao/config.yaml` by setting `YUNXIAO_CONFIG_FILE` to a temp directory.
- Avoid relying on real Yunxiao network calls; use `httptest` servers and controlled environment variables.

### Testing Requirements
- Run `go test ./test/integration -count=1` for integration-only changes.
- Run `go test ./...` before claiming project-level correctness.
- When changing golden files, verify JSON equality rather than string formatting assumptions where possible.

### Common Patterns
- Integration tests build a temporary binary with `go build -o yunxiao-test ./cmd/yunxiao`.
- `testEnv` strips inherited `YUNXIAO_*` variables before adding explicit overrides.

## Dependencies

### Internal
- `cmd/yunxiao/` is built and executed by integration tests.
- `internal/cli/`, `internal/httpx/`, and `internal/model/output/` behavior is validated indirectly through CLI subprocesses.

### External
- Go `testing`, `net/http/httptest`, and `os/exec`.
- `github.com/stretchr/testify/require` for assertions.
- `github.com/creack/pty` for auth terminal interaction tests.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
