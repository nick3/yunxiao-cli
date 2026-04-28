<!-- Parent: ../AGENTS.md -->

# installer

## Purpose
Hermetic tests for installer scripts and release artifact naming contracts.

## For AI Agents
- Keep tests offline by default. Use local fixtures and `httptest`, not real GitHub Releases.
- Exercise install, update, uninstall, checksum failures, missing binaries, dry-run mapping, and release naming separation.
- Do not commit generated archives or binaries; build fixtures in temp directories during tests.

## Testing Requirements
- Run `go test ./test/installer -count=1` while iterating.
- Run `go test ./...` before claiming project-level correctness.
