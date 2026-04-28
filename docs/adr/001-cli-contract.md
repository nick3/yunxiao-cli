# ADR-001: Yunxiao CLI Contract

Status: Accepted
Date: 2026-04-25

## Context

Yunxiao CLI is an Agent-first command line tool that wraps Alibaba Cloud Yunxiao DevOps capabilities as a standalone binary. The primary consumer is AI Agents calling via subprocess, not human developers. This ADR freezes the CLI's public contract before any business command implementation begins.

## 1. Command Namespace & Initial Domains

Top-level commands organized by Yunxiao business domain:

```
yunxiao org current
yunxiao codeup repos list --organization-id <id>
yunxiao codeup repo get --organization-id <id> --repo-id <id>
yunxiao flow pipelines list --organization-id <id>
yunxiao flow pipeline get --organization-id <id> --pipeline-id <id>
```

Phase 2 additions:

```
yunxiao projex projects list --organization-id <id>
yunxiao projex projects list --mine
# In central edition, projects list may omit --organization-id when org current returns lastOrganizationId.
yunxiao projex project get --organization-id <id> --project-id <id>
yunxiao projex project-templates list --organization-id <id>
yunxiao projex project-template fields --organization-id <id> --template-id <id>
yunxiao projex project create --organization-id <id> --name <name> --custom-code <CODE> --template-id <id> --scope private --yes
yunxiao projex project archive --organization-id <id> --project-id <id> --yes
yunxiao projex workitems list --organization-id <id> --category <category> --project-id <id>
yunxiao projex workitems list --mine --category <category>
yunxiao projex workitems list --mine --unfinished --category <category>
yunxiao projex workitem get --organization-id <id> --workitem-id <id>
yunxiao projex workitem create --organization-id <id> --project-id <id> --workitem-type-id <id> --subject <text> --assigned-to <user-id|self> --yes
yunxiao projex workitem update --organization-id <id> --workitem-id <id> --assigned-to <user-id|self> --yes
yunxiao projex workitem comments list --organization-id <id> --workitem-id <id>
yunxiao projex workitem comment create --organization-id <id> --workitem-id <id> --content <text> --yes
yunxiao projex workitem-types list --organization-id <id> --project-id <id> --category <category>
yunxiao projex workitem-types list --organization-id <id> --all
yunxiao projex workitem-types relations --organization-id <id> --workitem-type-id <id>
yunxiao projex workitem-type get --organization-id <id> --workitem-type-id <id>
yunxiao projex workitem-type fields --organization-id <id> --project-id <id> --workitem-type-id <id>
yunxiao projex workitem-type workflow --organization-id <id> --project-id <id> --workitem-type-id <id>
yunxiao projex sprints list --organization-id <id> --project-id <id>
yunxiao packages repos list --organization-id <id>
yunxiao packages artifacts list --organization-id <id> --repo-id <id> --repo-type <type>
yunxiao packages artifact get --organization-id <id> --repo-id <id> --artifact-id <id> --repo-type <type>
yunxiao testhub testcases list --organization-id <id> --test-repo-id <id>
yunxiao testhub testcase get --organization-id <id> --test-repo-id <id> --testcase-id <id>
yunxiao testhub directories list --organization-id <id> --test-repo-id <id>
yunxiao testhub testplans list --organization-id <id>
yunxiao raw request --method GET --path /oapi/...
```

Grammar rules:
- Collection operations use plural resource names: `repos list`, `pipelines list`
- Singular operations use singular resource names: `repo get`, `pipeline get`
- Action verbs: `list` / `get` / `create` / `update` / `delete` / `run` / `archive` (no synonyms)
- Metadata/subresource reads may use a singular resource followed by a noun subresource, such as `project-template fields`, `workitem-type fields`, and `workitem-type workflow`.
- Projex safe write testing uses a private project lifecycle: discover templates, create a clearly named private test project, write only inside it, and archive it for cleanup. Hard delete remains outside the Agent smoke-test path.
- No abbreviation aliases in v1
- Mismatched singular/plural returns a suggestion: "did you mean 'repos list'?"

Phase 1 domains: `org`, `codeup`, `flow`
Phase 2 additions: `projex`, `packages`, `testhub`, `raw`

### Projex Personal Workitems

`projex workitems list --mine` is a Phase 2 read-only aggregation command. It resolves the current user with `org current`, lists projects participated in by that user, then performs project-scoped `workitems:search` calls with `assignedTo` set to the current user. Plain `projex workitems list` accepts `--project-id` and the API-shaped `--space-id` alias for the same project/space identifier; if both are set they must match. `--unfinished` is only valid with `--mine`, filters completed workitems from the aggregated result, and fails instead of guessing when a workitem completion status is not recognizable. The command drains upstream pages for every participated project and returns one complete aggregate with `has_more: false`; `--page-size` only controls upstream fetch size in this mode. If any project-scoped workitem search fails, the aggregate command fails rather than returning partial results. Direct organization-level `workitems:search` without `spaceId` is not part of the v1 contract because the verified upstream API rejects that shape.

### Projex Workitem Writes

`projex workitem create`, `projex workitem update`, and `projex workitem comment create` are explicit write commands. They must fail with `PARAM_REQUIRED` / `param` / exit code 2 before auth or HTTP unless `--yes` is present. `workitem create` accepts `--project-id` and the API-shaped `--space-id` alias for the same project/space identifier; if both are set they must match. `workitem create/update --assigned-to self` resolves the current user ID before sending the write. Non-idempotent write HTTP methods are not auto-retried. Text file inputs such as `--description-file` and `--content-file` must read UTF-8 regular files only, reject empty files, cap input at 1MiB, and avoid leaking full local paths in errors. Create sends custom fields as nested `customFieldValues`; update expands custom fields to top-level request body fields to match the official Yunxiao MCP server call shape.

### Raw Request Boundary

`raw request` is a Phase 2 escape hatch for read-only API coverage gaps:
- Only `GET` is supported in Phase 2; non-read methods return `PARAM_INVALID` / `param` / exit code 2
- `--path` must start with `/oapi/`; absolute URLs are rejected
- Output still uses the standard JSON envelope and exit-code mapping
- Raw request does not bypass auth, timeout, retry, trace, or stdout/stderr rules
- Raw request cannot be used to bypass Projex write command confirmation

## 2. JSON Output Envelope & Version Field

All `--json` output uses this envelope:

```json
{
  "version": "v1",
  "data": <result or null>,
  "meta": {
    "trace_id": "<string, optional>",
    "pagination": {
      "next_token": "<string or null>",
      "page_size": <int>,
      "has_more": <bool>,
      "page": <int, optional>,
      "total_pages": <int, optional>,
      "total": <int, optional>,
      "prev_token": "<string, optional>"
    }
  },
  "error": {
    "code": "<ERROR_CODE>",
    "category": "<category>",
    "retryable": <bool>,
    "message": "<english message>",
    "upstream_status": <int or null>
  }
}
```

Field naming: all fields use `snake_case`.

On success: `data` is populated, `error` is `null`.
On failure: `data` is `null`, `error` is populated. stdout always emits the full JSON envelope regardless of success or failure.

## 3. stdout / stderr / Exit Code / Error Category

**stdout**: result data only. Always valid JSON when `--json` is active (or non-TTY default).
**stderr**: diagnostics only (warnings, retry logs, debug info, human-readable errors).

Exit codes:

| Code | Meaning | Error Category |
|------|---------|---------------|
| 0 | Success | - |
| 1 | General failure | general |
| 2 | Parameter error | param |
| 3 | Auth failed (401, token missing/invalid) | auth |
| 4 | Resource not found (404) | not_found |
| 5 | Rate limited (429) | rate_limit |
| 6 | Network failure (timeout, DNS, connection) | network |
| 7 | Upstream service error (5xx) | upstream |
| 8 | Forbidden (403, valid token but no permission) | forbidden |

Fine-grained error codes within a category are conveyed via `error.code` in JSON (e.g., `NETWORK_UNREACHABLE` vs `REQUEST_TIMEOUT` both map to exit code 6).

### Stderr Verbosity

| Flag | Level | Output |
|------|-------|--------|
| (default) | warning | warnings + errors + retry info |
| `--quiet` | error | errors only |
| `--verbose` | info | default + info-level diagnostics |
| `--debug` | debug | all diagnostics including HTTP request/response |

Stderr must never leak tokens or Authorization headers.

## 4. Flag / Env / Config Precedence

Priority (highest to lowest):
1. Command-line flag
2. Environment variable (`YUNXIAO_` prefix)
3. Config file (`~/.yunxiao/config.yaml` via Viper)
4. Built-in default

Key variables:

| Flag | Env | Config Key | Default |
|------|-----|-----------|---------|
| `--organization-id` | `YUNXIAO_ORGANIZATION_ID` | `organization_id` | (none) |
| `--region` | `YUNXIAO_REGION` | `region` | (none) |
| `--timeout` | `YUNXIAO_TIMEOUT` | `timeout` | 30 |
| - | `YUNXIAO_ACCESS_TOKEN` | `access_token` | (none, required) |
| - | `YUNXIAO_API_BASE_URL` | `api_base_url` | `https://openapi-rdc.aliyuncs.com` |

Rules:
- `YUNXIAO_API_BASE_URL` takes absolute precedence over `--region` for endpoint resolution
- `--organization-id` flag overrides `YUNXIAO_ORGANIZATION_ID` env, `YUNXIAO_REGION_DEFAULT_ORG_ID`, and config
- `YUNXIAO_REGION_DEFAULT_ORG_ID` is an environment-provided organization fallback for region endpoints and takes precedence over config `organization_id`

## 5. List / Pagination Contract

All list commands support: `--page-size`, `--page-token`

JSON response always includes `meta.pagination`:
- `next_token`: string or null
- `page_size`: current page size
- `has_more`: boolean
- `page`: current page number when the upstream response exposes it
- `total_pages`: total page count when the upstream response exposes it
- `total`: total matched item count when the upstream response exposes it
- `prev_token`: previous page token when the upstream response exposes it

Rules:
- `has_more: false` implies `next_token: null`
- Total metadata is optional and must be omitted when the upstream response does not expose it
- v1 does not return partial results; a list command either succeeds with a full page or fails entirely

## 6. Compatibility / Versioning

- CLI contract follows SemVer
- JSON envelope includes `"version": "v1"`
- Minor versions: additive only (new fields, new commands, new error codes)
- Major versions: may remove/rename fields, must provide migration guide
- Once public: exit codes, top-level command names, flag names, error categories are managed as public API
- Compatibility tests must cover: golden JSON, stderr separation, exit code semantics, pagination structure

## 7. Timeout Contract

| Setting | Default | Override |
|---------|---------|---------|
| Connection timeout | 10s | (internal) |
| Request timeout | 30s | `--timeout <seconds>` |

- All HTTP requests have a deadline; CLI never hangs indefinitely
- Timeout maps to exit code 6 (network), `error.code: "REQUEST_TIMEOUT"`

## 8. Retry Contract

| Setting | Default | Override |
|---------|---------|---------|
| Max retries | 3 | `--no-retry` disables |
| Backoff | Exponential (1s, 2s, 4s) | Respects `Retry-After` header |
| Retryable status | 429 and all 5xx | - |
| Retryable methods | GET, HEAD only | Non-idempotent methods never auto-retry |

- Retry logs go to stderr: `[RETRY] upstream returned 503, attempt 2/3`
- After retry budget exhausted, returns the error with `retryable: true`

## 9. Command Introspection

- `yunxiao commands --json`: outputs structured list of all commands with paths, required/optional params, param types
- `yunxiao <domain> <resource> <action> --help --json`: structured help for a single command
- Introspection is Phase 1 infrastructure, not deferred

## Golden Flow Validation

### Flow 1: org current (success)
```bash
$ YUNXIAO_ACCESS_TOKEN=valid-token yunxiao org current --json
```
stdout:
```json
{"version":"v1","data":{"id":"user-001","name":"agent-user","organization":{"id":"org-123","name":"demo-org"}},"meta":{"trace_id":""},"error":null}
```
exit code: 0

### Flow 1: org current (auth failure)
```bash
$ yunxiao org current --json
```
stdout:
```json
{"version":"v1","data":null,"meta":{"trace_id":""},"error":{"code":"AUTH_FAILED","category":"auth","retryable":false,"message":"YUNXIAO_ACCESS_TOKEN is missing or invalid","upstream_status":null}}
```
exit code: 3

### Flow 2: codeup repos list (success with pagination)
```bash
$ YUNXIAO_ACCESS_TOKEN=valid-token yunxiao codeup repos list --organization-id org-123 --json --page-size 2
```
stdout:
```json
{"version":"v1","data":[{"id":"repo-1","name":"frontend"},{"id":"repo-2","name":"backend"}],"meta":{"trace_id":"","pagination":{"next_token":"abc123","page_size":2,"has_more":true}},"error":null}
```
exit code: 0

### Flow 3: flow pipeline get (not found)
```bash
$ YUNXIAO_ACCESS_TOKEN=valid-token yunxiao flow pipeline get --organization-id org-123 --pipeline-id nonexistent --json
```
stdout:
```json
{"version":"v1","data":null,"meta":{"trace_id":""},"error":{"code":"RESOURCE_NOT_FOUND","category":"not_found","retryable":false,"message":"resource not found","upstream_status":404}}
```
exit code: 4

### Flow 3: flow pipeline get (upstream 503, retry exhausted)
```bash
$ YUNXIAO_ACCESS_TOKEN=valid-token yunxiao flow pipeline get --organization-id org-123 --pipeline-id p-1 --json
```
stdout:
```json
{"version":"v1","data":null,"meta":{"trace_id":""},"error":{"code":"UPSTREAM_UNAVAILABLE","category":"upstream","retryable":true,"message":"upstream service error","upstream_status":503}}
```
stderr:
```
[RETRY] upstream returned 503, attempt 1/3
[RETRY] upstream returned 503, attempt 2/3
[RETRY] upstream returned 503, attempt 3/3
```
exit code: 7
