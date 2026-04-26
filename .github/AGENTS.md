<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# .github

## Purpose
This directory contains GitHub automation for CI and releases. Workflows enforce formatting, vetting, tests, builds, and release packaging for Yunxiao CLI.

## Key Files
| File | Description |
|------|-------------|
| _(none)_ | Workflow files live under `workflows/`. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `workflows/` | GitHub Actions workflow definitions (see `workflows/AGENTS.md`). |

## For AI Agents

### Working In This Directory
- Treat workflow edits as shared automation changes; keep permissions minimal and commands aligned with local verification.
- Do not add secrets, tokens, or production credentials to workflow files.

### Testing Requirements
- Mirror workflow commands locally where possible: gofmt check, `go vet ./...`, `go test ./...`, and `go build`.
- For release workflow changes, verify `.goreleaser.yaml` remains compatible.

### Common Patterns
- CI runs on pull requests and pushes to `main`.
- Releases should be tag-driven and use GoReleaser configuration from the repository root.

## Dependencies

### Internal
- `.goreleaser.yaml` for release packaging.
- `go.mod` for Go version selection.

### External
- GitHub Actions, `actions/checkout`, `actions/setup-go`, and GoReleaser actions.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
