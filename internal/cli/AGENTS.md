<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# cli

## Purpose
This package owns CLI infrastructure shared by all commands: root command construction, global flags, output format selection, JSON envelope rendering, and exit code helpers.

## Key Files
| File | Description |
|------|-------------|
| `root.go` | Defines the root `yunxiao` command and persistent flags such as `--json`, `--human`, `--quiet`, `--timeout`, `--no-retry`, `--organization-id`, `--region`, and `--trace-id`. |
| `output.go` | Renders successful results and error details in JSON or human-readable format. |
| `exitcodes.go` | Defines public exit code constants. |
| `exit_error.go` | Carries explicit exit codes through Cobra execution. |
| `output_test.go` | Unit tests for output rendering. |
| `exit_error_test.go` | Unit tests for exit-code error helpers. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Shared CLI infrastructure package. |

## For AI Agents

### Working In This Directory
- Treat exit codes and JSON envelope behavior as public API.
- Keep stdout machine-readable under JSON mode; diagnostics belong on stderr.
- Do not add global flags without updating command introspection and relevant docs.

### Testing Requirements
- Run `go test ./internal/cli -count=1` for local changes.
- Run integration tests if output shape, global flags, or exit codes change.

### Common Patterns
- Commands call `cli.WriteResult` or `cli.WriteError` and then exit with the returned code when needed.
- `GetOutputFormat` gives `--json` priority over auto/human behavior.

## Dependencies

### Internal
- `internal/model/output/` provides envelope and error structs.

### External
- `github.com/spf13/cobra` for command construction.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
