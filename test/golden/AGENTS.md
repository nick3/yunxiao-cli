<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# golden

## Purpose
This directory stores expected JSON envelope outputs used by integration tests. Golden files make public CLI behavior explicit and easy to compare across success and failure cases.

## Key Files
| File | Description |
|------|-------------|
| `org_current_success.json` | Expected `org current` success envelope. |
| `org_current_auth_failed.json` | Expected missing-token auth failure envelope. |
| `org_current_forbidden.json` | Expected forbidden response envelope. |
| `org_current_decode_failed.json` | Expected invalid JSON upstream response envelope. |
| `org_current_timeout.json` | Expected timeout response envelope. |
| `codeup_repos_list_success.json` | Expected Codeup repositories list envelope. |
| `codeup_repo_get_success.json` | Expected Codeup repository detail envelope. |
| `codeup_repo_get_empty_response.json` | Expected empty upstream body failure envelope. |
| `flow_pipelines_list_success.json` | Expected Flow pipelines list envelope. |
| `flow_pipeline_get_success.json` | Expected Flow pipeline detail envelope. |
| `flow_pipeline_get_not_found.json` | Expected not-found envelope. |
| `flow_pipeline_get_rate_limited.json` | Expected rate-limit envelope. |
| `flow_pipeline_get_upstream_500.json` | Expected upstream 500 envelope. |
| `flow_pipeline_get_upstream_503.json` | Expected upstream 503 envelope. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | JSON fixtures only. |

## For AI Agents

### Working In This Directory
- Update golden files only for intentional contract changes.
- Preserve JSON envelope fields and semantic values; formatting can differ if tests use `require.JSONEq`.
- Do not include real tokens or sensitive upstream data.

### Testing Requirements
- Run the integration test that consumes a changed golden file.
- Run `go test ./test/integration -count=1` after broad fixture updates.

### Common Patterns
- File names include command and outcome, such as success, auth failure, not found, timeout, or upstream error.

## Dependencies

### Internal
- `test/integration/` loads these files and compares them to CLI subprocess stdout.

### External
- No runtime dependencies.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
