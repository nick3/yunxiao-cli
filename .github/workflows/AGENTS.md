<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# workflows

## Purpose
This directory contains GitHub Actions workflows for continuous integration and release automation.

## Key Files
| File | Description |
|------|-------------|
| `ci.yml` | Runs on pull requests and pushes to `main`; checks gofmt, `go vet`, `go test`, and `go build`. |
| `preview-release.yml` | Runs on pushes to `main`; checks formatting, vet, tests, builds preview artifacts, and publishes the `preview` prerelease. |
| `release.yml` | Runs on `v*` tags; checks vet/tests and publishes release artifacts through GoReleaser. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Workflow YAML files only. |

## For AI Agents

### Working In This Directory
- Keep workflow permissions minimal; avoid broad write scopes except release publishing where required.
- Keep CI commands in sync with local verification guidance in root `AGENTS.md`.
- Do not add secrets directly to YAML; use GitHub secrets references only when needed.

### Testing Requirements
- Run equivalent local commands before claiming workflow changes are safe.
- For release changes, validate `.goreleaser.yaml` and GoReleaser command assumptions.

### Common Patterns
- `actions/setup-go` reads `go.mod` through `go-version-file`.
- Preview and release publishing workflows use GitHub-provided tokens from the Actions secrets context.

## Dependencies

### Internal
- `go.mod` supplies Go version.
- `.goreleaser.yaml` supplies release packaging.

### External
- GitHub Actions, checkout/setup-go actions, and GoReleaser action.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
