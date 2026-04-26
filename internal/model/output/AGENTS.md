<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# output

## Purpose
This package defines the public JSON envelope returned by Yunxiao CLI. These types are part of the automation contract consumed by AI agents, scripts, and CI.

## Key Files
| File | Description |
|------|-------------|
| `envelope.go` | Defines `Envelope`, `Meta`, `Pagination`, and `ErrorDetail` with JSON field names used on stdout. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single model package. |

## For AI Agents

### Working In This Directory
- Treat JSON tags and field semantics as public API; update ADR, README, skill docs, and tests for any change.
- Keep `error` nullable on success and `data` nullable on failure.
- Keep `retryable` explicit; do not rely on omitted fields for control flow.

### Testing Requirements
- Run `go test ./internal/cli -count=1` and integration tests after envelope changes.
- Update golden JSON fixtures under `test/golden/` for intentional output shape changes.

### Common Patterns
- Pagination always uses `next_token`, `page_size`, and `has_more`; `page`, `total_pages`, `total`, and `prev_token` are optional upstream metadata.
- Error details use `code`, `category`, `retryable`, `message`, and `upstream_status`.

## Dependencies

### Internal
- Consumed by CLI output, command adapters, HTTP classification, and domain helpers.

### External
- No third-party dependencies.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
