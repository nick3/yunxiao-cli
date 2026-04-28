<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# yunxiao-cli

## Purpose
Yunxiao CLI is an Agent-first Go command line tool for Alibaba Cloud Yunxiao DevOps automation. It wraps read-only Yunxiao capabilities behind a stable subprocess contract: JSON envelopes on stdout, diagnostics on stderr, deterministic exit codes, explicit pagination, auth, timeout, and retry semantics.

## Key Files
| File | Description |
|------|-------------|
| `README.md` | User-facing installation, auth, configuration, command, and automation guide. |
| `go.mod` | Go module definition and dependency list for Cobra, Viper, terminal handling, and tests. |
| `go.sum` | Locked module checksums. |
| `.goreleaser.yaml` | Release build configuration for multi-platform CLI artifacts. |
| `.gitignore` | Ignore rules for local binaries, AI tool state, and generated artifacts. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `.github/` | GitHub Actions CI and release workflows (see `.github/AGENTS.md`). |
| `cmd/` | CLI executable entrypoints (see `cmd/AGENTS.md`). |
| `docs/` | Architecture decisions and phase status documentation (see `docs/AGENTS.md`). |
| `internal/` | Application source code split into CLI, command, domain, HTTP, config, auth, and model layers (see `internal/AGENTS.md`). |
| `skill/` | AI Agent skill documentation for using the built CLI correctly (see `skill/AGENTS.md`). |
| `test/` | Integration tests, golden JSON fixtures, and test fixtures (see `test/AGENTS.md`). |

## For AI Agents

### Working In This Directory
- Preserve the Agent-first CLI contract: `--json` stdout must be machine-readable JSON, and stderr must remain diagnostics-only.
- Do not add command aliases or inferred flags without updating command introspection, README, ADR, tests, and skill documentation.
- Keep implementation changes minimal and aligned with KISS/YAGNI; prefer extending existing command/domain patterns over adding new abstractions.
- When discovering Alibaba Cloud Yunxiao DevOps capabilities or how to call them, reference `alibabacloud-devops-mcp-server` at `/Users/nick/Workspace/alibabacloud-devops-mcp-server` to understand Yunxiao API coverage, authentication, error-handling experience, domain boundaries, and API invocation patterns.
- Never commit local auth material, tokens, `.omc/`, or built binaries such as `yunxiao`.

### Testing Requirements
- After Go changes, run `gofmt` on touched Go files and verify `test -z "$(gofmt -l cmd internal test)"`.
- Run `go vet ./...`, `go test ./...`, and `go build -o yunxiao ./cmd/yunxiao` before claiming completion.
- For CLI contract changes, update or add integration tests and golden JSON fixtures under `test/`.

### Common Patterns
- CLI wiring starts in `cmd/yunxiao/main.go`, root flags live in `internal/cli/`, Cobra commands live in `internal/command/`, and API/domain logic lives in `internal/domains/`.
- Output uses `internal/model/output.Envelope`; failures still write structured JSON to stdout when JSON mode is active.
- Auth prefers `YUNXIAO_ACCESS_TOKEN` over config file tokens; config defaults are centralized in `internal/config/config.go`.

## Dependencies

### Internal
- `cmd/yunxiao/` wires all command packages and configuration.
- `internal/cli/` owns global flags, output rendering, and exit-code mapping.
- `internal/domains/shared/` centralizes API request and pagination helpers used by multiple domains.
- `test/integration/` validates the public CLI contract with subprocess-level tests.

### External
- `github.com/spf13/cobra` - command tree and flag handling.
- `github.com/spf13/viper` - environment and config file loading.
- `github.com/stretchr/testify` - test assertions.
- `golang.org/x/term` and `github.com/creack/pty` - terminal/auth integration tests and input behavior.
- `go.yaml.in/yaml/v3` - YAML config file handling.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->

## Skill routing

When the user's request matches an available skill, invoke it via the Skill tool. When in doubt, invoke the skill.

Key routing rules:
- Product ideas/brainstorming → invoke /office-hours
- Strategy/scope → invoke /plan-ceo-review
- Architecture → invoke /plan-eng-review
- Design system/plan review → invoke /design-consultation or /plan-design-review
- Full review pipeline → invoke /autoplan
- Bugs/errors → invoke /investigate
- QA/testing site behavior → invoke /qa or /qa-only
- Code review/diff check → invoke /review
- Visual polish → invoke /design-review
- Ship/deploy/PR → invoke /ship or /land-and-deploy
- Save progress → invoke /context-save
- Resume context → invoke /context-restore
