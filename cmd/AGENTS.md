<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# cmd

## Purpose
This directory contains executable entrypoints for Yunxiao CLI. It should stay thin: initialize configuration, construct the root Cobra command, register subcommands, and translate top-level errors into the public CLI output contract.

## Key Files
| File | Description |
|------|-------------|
| _(none)_ | Source files are kept in executable-specific subdirectories such as `yunxiao/`. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `yunxiao/` | Main CLI binary entrypoint (see `yunxiao/AGENTS.md`). |

## For AI Agents

### Working In This Directory
- Keep executable packages small and delegate business behavior to `internal/command/` and `internal/domains/`.
- Add new binaries only when the project explicitly needs a separate executable; otherwise extend `cmd/yunxiao/`.

### Testing Requirements
- Run `go build -o yunxiao ./cmd/yunxiao` after entrypoint changes.
- Run `go test ./...` when command registration or error handling changes.

### Common Patterns
- Entrypoints should wire dependencies and return exit codes, not perform domain API work directly.

## Dependencies

### Internal
- `internal/cli/` for root command and exit behavior.
- `internal/command/` for Cobra command packages.
- `internal/config/` for initialization.

### External
- Standard Go runtime only at this container level.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
