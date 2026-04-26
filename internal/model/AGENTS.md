<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# model

## Purpose
This directory contains shared data model packages used by CLI output and domain behavior. The most important model is the public JSON envelope returned to automation consumers.

## Key Files
| File | Description |
|------|-------------|
| _(none)_ | Model types are organized into subpackages. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `output/` | Public output envelope, pagination, and error details (see `output/AGENTS.md`). |

## For AI Agents

### Working In This Directory
- Treat exported JSON fields as public contract fields.
- Use snake_case JSON tags for envelope-facing structs.
- Avoid adding fields unless tests and docs cover the additive contract change.

### Testing Requirements
- Run output and integration tests after changing envelope-related models.

### Common Patterns
- Commands pass `output.Meta` through result rendering so trace IDs and pagination stay with the response.

## Dependencies

### Internal
- Used by `internal/cli/`, `internal/httpx/`, `internal/command/`, and `internal/domains/`.

### External
- No direct third-party dependencies at this container level.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
