<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# packages

## Purpose
This package implements package repository and artifact CLI commands: `packages repos list`, `packages artifacts list`, and `packages artifact get`. It validates repository identifiers, repository types, pagination flags, and optional filters before delegating to the packages domain layer.

## Key Files
| File | Description |
|------|-------------|
| `packages.go` | Defines package repository/artifact command tree, flags, required metadata, pagination handling, and API client setup. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single command package. |

## For AI Agents

### Working In This Directory
- Preserve required flags for organization, repository ID, artifact ID, and repository type.
- Keep optional filters (`--repo-types`, `--repo-categories`, `--search`) documented through command introspection.
- Keep path and payload mapping in `internal/domains/packages/`.

### Testing Requirements
- Run Phase 2 integration tests after package command changes.
- Run command help/introspection tests after flag changes.

### Common Patterns
- List operations expose pagination and attach `meta.pagination`; get operations return a single object.

## Dependencies

### Internal
- `internal/domains/packages/` for API calls.
- `internal/command/validation/` for `--page-size`.
- `internal/command/flagmeta/` for required flag annotations.

### External
- Cobra for command definitions.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
