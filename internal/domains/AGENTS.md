<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# domains

## Purpose
This directory contains domain-level Yunxiao API logic. Domain packages construct API paths, encode query strings or search payloads, call shared HTTP helpers, and convert upstream pagination into the CLI `output.Pagination` model.

## Key Files
| File | Description |
|------|-------------|
| _(none)_ | Domain code is split by Yunxiao product area. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `codeup/` | Code repository API logic (see `codeup/AGENTS.md`). |
| `flow/` | Pipeline API logic (see `flow/AGENTS.md`). |
| `packages/` | Package repository and artifact API logic (see `packages/AGENTS.md`). |
| `projex/` | Project, workitem, and sprint API logic (see `projex/AGENTS.md`). |
| `raw/` | Raw read-only request validation and execution (see `raw/AGENTS.md`). |
| `shared/` | Shared JSON request, pagination, and base URL helpers (see `shared/AGENTS.md`). |
| `testhub/` | Testhub API logic (see `testhub/AGENTS.md`). |

## For AI Agents

### Working In This Directory
- Keep CLI flag parsing out of domain packages; accept already-validated parameters.
- Preserve central-vs-region path branching where APIs differ by base URL.
- Return `*output.ErrorDetail` instead of printing or exiting.

### Testing Requirements
- Run package-specific unit tests for path logic.
- Run integration tests for affected commands because domains are mostly validated through subprocess flows.

### Common Patterns
- Central OpenAPI endpoints include `/organizations/{organizationID}` while region endpoints often omit organization from the path.
- Search-style endpoints use POST with payloads and `nextPage`; page-header endpoints read `x-next-page`.

## Dependencies

### Internal
- `internal/httpx/` for transport.
- `internal/model/output/` for error and pagination structs.
- `internal/domains/shared/` for common helpers.

### External
- Go standard `net/http`, `net/url`, and context packages.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
