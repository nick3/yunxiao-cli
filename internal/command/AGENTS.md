<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# command

## Purpose
This directory contains Cobra command adapters for each Yunxiao CLI domain. Command packages parse flags, validate required inputs, resolve config/auth context, call domain functions, set metadata, and render output through the shared CLI envelope functions.

## Key Files
| File | Description |
|------|-------------|
| _(none)_ | Commands are split by domain and helper packages. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `auth/` | Interactive and non-interactive auth commands (see `auth/AGENTS.md`). |
| `codeup/` | Codeup repository commands (see `codeup/AGENTS.md`). |
| `flagmeta/` | Required flag metadata helpers (see `flagmeta/AGENTS.md`). |
| `flow/` | Flow pipeline commands (see `flow/AGENTS.md`). |
| `meta/` | Command introspection and JSON help (see `meta/AGENTS.md`). |
| `org/` | Current organization/user command (see `org/AGENTS.md`). |
| `packages/` | Package repository and artifact commands (see `packages/AGENTS.md`). |
| `projex/` | Projex project, workitem, and sprint commands (see `projex/AGENTS.md`). |
| `raw/` | Read-only raw API request command (see `raw/AGENTS.md`). |
| `testhub/` | Testhub testcase, directory, and testplan commands (see `testhub/AGENTS.md`). |
| `validation/` | Shared flag validation helpers (see `validation/AGENTS.md`). |

## For AI Agents

### Working In This Directory
- Keep command packages focused on CLI concerns; put API path and response mapping logic in `internal/domains/`.
- Mark required flags with `flagmeta.MustMarkRequired` so `commands --json` and `--help --json` stay accurate.
- For list commands, validate positive `--page-size` before auth when the command exposes pagination.

### Testing Requirements
- Run `go test ./internal/command/... -count=1` after command helper changes.
- Run `go test ./test/integration -run TestCommands -count=1` when command paths, flags, or help metadata change.
- Run domain-specific integration tests after changing command behavior.

### Common Patterns
- Each domain package exposes `New<Domain>Cmd()` for registration in `cmd/yunxiao/main.go`.
- Commands resolve trace ID, timeout, retry, quiet, token, and base URL before creating `httpx.Client`.

## Dependencies

### Internal
- `internal/cli/` for output and exit handling.
- `internal/auth/` and `internal/config/` for execution context.
- `internal/domains/*/` for API calls.
- `internal/model/output/` for metadata and errors.

### External
- `github.com/spf13/cobra` and `github.com/spf13/pflag`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
