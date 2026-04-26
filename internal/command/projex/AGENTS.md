<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# projex

## Purpose
This package implements Projex project collaboration commands for projects, workitems, and sprints. It exposes list/get style commands with required organization and entity identifiers, plus pagination where the API supports it.

## Key Files
| File | Description |
|------|-------------|
| `projects.go` | Defines Projex command tree, project/workitem/sprint commands, flags, pagination, API client creation, and error rendering. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Single command package. |

## For AI Agents

### Working In This Directory
- Keep command grammar consistent: plural resources list collections, singular resources get individual objects.
- Validate `category` and `space-id` for workitem lists before network calls.
- Keep API path details in `internal/domains/projex/`.

### Testing Requirements
- Run Phase 2 integration tests after Projex command changes.
- Run `go test ./internal/domains/projex -count=1` when path behavior changes.

### Common Patterns
- List commands use page tokens and positive page sizes; get commands require entity IDs.

## Dependencies

### Internal
- `internal/domains/projex/` for path construction and API calls.
- `internal/command/validation/` and `internal/command/flagmeta/` for flag behavior.

### External
- Cobra for command definitions.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
