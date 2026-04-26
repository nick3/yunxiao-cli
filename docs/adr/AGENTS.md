<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# adr

## Purpose
This directory contains Architecture Decision Records that define stable project and CLI contract decisions. ADRs should explain why public behavior exists and what compatibility constraints apply.

## Key Files
| File | Description |
|------|-------------|
| `001-cli-contract.md` | Accepted contract for command namespace, JSON envelope, stdout/stderr separation, exit codes, config precedence, pagination, timeout, retry, and command introspection. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | ADR files only. |

## For AI Agents

### Working In This Directory
- Keep ADRs factual and aligned with current implementation unless marking a future decision explicitly.
- For public CLI behavior changes, update ADRs together with README, integration tests, golden files, and skill documentation.
- Do not silently change accepted contract text without verifying implementation and tests.

### Testing Requirements
- No direct test runner exists for ADRs; validate contract claims through Go integration tests.

### Common Patterns
- ADR names use numeric prefixes and concise titles.
- Public contract details should be precise enough for automation consumers.

## Dependencies

### Internal
- `internal/cli/`, `internal/command/`, `internal/model/output/`, and `test/integration/` implement and verify these decisions.

### External
- No runtime dependencies.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
