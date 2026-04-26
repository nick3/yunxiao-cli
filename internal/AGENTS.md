<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# internal

## Purpose
This directory contains the non-exported application implementation for Yunxiao CLI. The code is organized into layers: CLI infrastructure, auth/config, Cobra command adapters, domain API logic, HTTP transport, and output models.

## Key Files
| File | Description |
|------|-------------|
| _(none)_ | Implementation is organized into subpackages. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `auth/` | Access-token resolution and token source precedence (see `auth/AGENTS.md`). |
| `cli/` | Root command flags, output rendering, and exit codes (see `cli/AGENTS.md`). |
| `command/` | Cobra command definitions and flag parsing (see `command/AGENTS.md`). |
| `config/` | Viper-backed config and environment handling (see `config/AGENTS.md`). |
| `domains/` | Yunxiao API domain logic and shared request helpers (see `domains/AGENTS.md`). |
| `httpx/` | HTTP client, retry, timeout, and HTTP error classification (see `httpx/AGENTS.md`). |
| `model/` | Public JSON envelope and related model packages (see `model/AGENTS.md`). |

## For AI Agents

### Working In This Directory
- Maintain the command/domain split: `internal/command/*` parses CLI flags and formats results; `internal/domains/*` builds API calls and maps data.
- Keep public behavior compatible with `docs/adr/001-cli-contract.md`; any contract change must update tests and docs.
- Avoid leaking tokens to stderr, logs, errors, or tests.

### Testing Requirements
- Run focused package tests for touched subpackages, then `go test ./...` for cross-package behavior.
- Run integration tests when changing command wiring, output envelopes, exit codes, auth, config, HTTP, or pagination behavior.

### Common Patterns
- Commands construct `httpx.Client` with `config.GetBaseURL()`, resolved token, timeout, retry, and trace options.
- Domain helpers return `*output.ErrorDetail` instead of panicking or printing directly.
- JSON envelope fields use snake_case and are defined in `internal/model/output/`.

## Dependencies

### Internal
- `internal/model/output/` is shared across CLI, HTTP, command, and domain layers.
- `internal/httpx/` is used by domain shared request helpers and command-level auth verification.
- `internal/config/` and `internal/auth/` feed command execution context.

### External
- Cobra and pflag for command and flag handling.
- Viper for environment/config resolution.
- Go standard `net/http` for transport.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
