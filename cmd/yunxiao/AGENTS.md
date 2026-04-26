<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# yunxiao

## Purpose
This directory contains the main `yunxiao` binary entrypoint. It initializes configuration, installs JSON help, registers all top-level commands, executes Cobra, and maps command-level errors into the standard CLI error envelope.

## Key Files
| File | Description |
|------|-------------|
| `main.go` | Boots config, constructs the root command, registers `auth`, `org`, `codeup`, `flow`, `projex`, `packages`, `testhub`, `raw`, and `commands`, then handles top-level failures. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Entrypoint package only. |

## For AI Agents

### Working In This Directory
- Keep `main.go` focused on wiring and top-level error translation.
- Register new top-level commands here only after adding the command package and integration tests.
- Preserve `meta.InstallJSONHelp(root)` and `meta.NewCommandsCmd(root)` so command discovery remains available to automation.

### Testing Requirements
- Run `go build -o yunxiao ./cmd/yunxiao` after changes.
- Run `go test ./test/integration -run TestCommands -count=1` when command registration changes.

### Common Patterns
- Cobra `RunE` errors with known exit codes pass through `cli.ExitCode`.
- Unknown command/flag failures are converted to JSON errors with category `param`.

## Dependencies

### Internal
- `internal/cli/` for root command, output, and exit code handling.
- `internal/config/` for startup configuration.
- `internal/command/*/` for registered command trees.
- `internal/model/output/` for startup error envelopes.

### External
- Cobra is used indirectly through internal command packages.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
