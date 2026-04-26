---
name: using-yunxiao-cli
description: Use when an AI Agent calls Yunxiao CLI from shell, CI, subprocess, automation, or scripts and needs stable JSON output, command discovery, pagination handling, raw read-only API access, auth, retry, timeout, or error-code decisions.
---

# Using Yunxiao CLI

## Overview

Yunxiao CLI is Agent-first: always treat stdout JSON envelope, exit code, and `error.category` as the API contract. Do not infer command names, flags, pagination, or retry behavior from generic CLI conventions.

## When to Use

Use for `yunxiao` subprocess calls, CI automation, shell scripting, command discovery, JSON parsing, pagination, auth failures, retry handling, `raw request`, timeout, `PARAM_INVALID`, `AUTH_FAILED`, `COMMAND_FAILED`, `REQUEST_TIMEOUT`, `UPSTREAM_UNAVAILABLE`, or stderr/stdout separation.

## Required First Step

Discover commands from the CLI itself:

```bash
yunxiao commands --json
yunxiao <domain> <resource> <action> --help --json
```

`commands --json` returns a JSON envelope whose `.data` is a top-level command tree. Recursively walk every `.subcommands[]` node to find leaf commands; do not only read `.data[].path`.

```bash
yunxiao commands --json | jq -r '.. | objects | select(has("path") and ((.subcommands // []) | length == 0)) | .path'
```

If `yunxiao` is not installed on `PATH` and you are inside this repository, use `./yunxiao` or the absolute binary path. Never guess aliases such as `repo list`, `testplans --page`, `--format json`, `--limit`, or `--cursor`.

## Core Contract

| Area | Rule |
|---|---|
| Machine output | Pass `--json`; parse stdout only. |
| Failure output | Non-zero exits still write JSON envelope to stdout; capture it before exiting. |
| Diagnostics | stderr is only diagnostics; never use stderr as the structured error source. |
| Error fields | Branch on `.error.category`; log `.error.code`; treat missing `.error.retryable` as `false`. |
| Auth | Automation should use `YUNXIAO_ACCESS_TOKEN`; missing token maps to `AUTH_FAILED` / exit 3. |
| Organization | Prefer explicit `--organization-id`; env/config fallback may exist. |
| Page size | `--page-size` must be positive integer; invalid values are `PARAM_INVALID` / exit 2 before auth. |
| Pagination | Follow `meta.pagination.next_token` only when `meta.pagination.has_more` is true; use optional `meta.pagination.total` for counts when present. |
| Testplans | `testhub testplans list` is not paginated; do not pass `--page-size` or `--page-token`. |
| Raw request | Phase 2 raw is read-only: only `--method GET` and `--path /oapi/...`. |

## Auth Guidance

Use `YUNXIAO_ACCESS_TOKEN` for automation. Human users may run `yunxiao auth` in a private real terminal for visible interactive token entry. Agents, CI, and scripts must not trigger the interactive path; use env or explicit stdin instead:

```bash
printf '%s\n' "$YUNXIAO_ACCESS_TOKEN" | yunxiao auth login --token-stdin --json
yunxiao auth status --json
yunxiao auth status --verify --json
yunxiao auth logout --json
```

CI normally only needs the `YUNXIAO_ACCESS_TOKEN` environment variable, not `auth login`. Add `--force` to `auth login` only when intentionally overwriting an existing config token.

Only add `--skip-verify` for explicit offline/internal-network cases; it saves an unverified token. Never pass tokens as command-line arguments. `YUNXIAO_ACCESS_TOKEN` takes precedence over config, and `auth logout` only removes the config token.

## Command Quick Reference

```bash
yunxiao auth status --json
yunxiao org current --json
yunxiao org members list --organization-id <org> --json
yunxiao codeup repos list --organization-id <org> --json
yunxiao codeup repo get --organization-id <org> --repo-id <repo> --json
yunxiao codeup branches list --organization-id <org> --repo-id <repo> --json
yunxiao codeup commits list --organization-id <org> --repo-id <repo> --json
yunxiao codeup file get --organization-id <org> --repo-id <repo> --path <file> --ref <ref> --json
yunxiao codeup compare get --organization-id <org> --repo-id <repo> --from <ref> --to <ref> --json
yunxiao flow pipelines list --organization-id <org> --json
yunxiao flow pipeline get --organization-id <org> --pipeline-id <pipeline> --json
yunxiao flow runs list --organization-id <org> --pipeline-id <pipeline> --json
yunxiao flow run get --organization-id <org> --pipeline-id <pipeline> --run-id <run> --json
yunxiao projex projects list --organization-id <org> --json
# 中心站可省略 --organization-id；若 lastOrganizationId 为空会返回 PARAM_REQUIRED。当前账号参与项目可用：
yunxiao projex projects list --mine --json
yunxiao projex project get --organization-id <org> --project-id <project> --json
yunxiao projex workitems list --organization-id <org> --category <category> --space-id <space> --json
yunxiao projex workitem get --organization-id <org> --workitem-id <workitem> --json
yunxiao projex sprints list --organization-id <org> --project-id <project> --json
yunxiao packages repos list --organization-id <org> --json
yunxiao packages artifacts list --organization-id <org> --repo-id <repo> --repo-type <type> --json
yunxiao packages artifact get --organization-id <org> --repo-id <repo> --artifact-id <artifact> --repo-type <type> --json
yunxiao testhub testcases list --organization-id <org> --test-repo-id <repo> --json
yunxiao testhub testcase get --organization-id <org> --test-repo-id <repo> --testcase-id <case> --json
yunxiao testhub directories list --organization-id <org> --test-repo-id <repo> --json
yunxiao testhub testplans list --organization-id <org> --json
yunxiao raw request --method GET --path /oapi/... --json
```

## Projex Scope

Projex commands currently support project/space-scoped list enumeration and known-ID detail reads. `projex projects list` can auto-resolve the central-edition organization from `org current.lastOrganizationId`, supports MCP-compatible filters such as `--name`, `--status`, `--admin-user-id`, `--scenario-filter`, `--user-id`, `--advanced-conditions`, and `--extra-conditions`, and exposes `--mine` for projects participated in by the current user. `--mine` is mutually exclusive with explicit `--scenario-filter` / `--user-id`; `--advanced-conditions` overrides basic project filters, while `--scenario-filter` plus `--user-id` overrides `--extra-conditions`. `projex workitems list` requires explicit `--organization-id`, `--category`, and `--space-id`, and exposes MCP-compatible filters such as `--status`, `--assigned-to`, `--finish-time-after`, `--update-status-at-after`, `--order-by`, and `--sort`. Do not assume cross-project personal todo search, unfinished-state detection, or this-week completion aggregation are public CLI contracts. Use `commands --json` and `--help --json` before calling Projex commands, and treat unsupported business filters as capability gaps.

## Parsing Pattern

```bash
out=$(YUNXIAO_ACCESS_TOKEN="$YUNXIAO_ACCESS_TOKEN" yunxiao codeup repos list --organization-id "$ORG_ID" --json 2>diag)
status=$?

code=$(jq -r '.error.code // empty' <<<"$out")
category=$(jq -r '.error.category // empty' <<<"$out")

if [ "$status" -eq 0 ]; then
  jq '.data' <<<"$out"
else
  jq '.error' <<<"$out" >&2
  exit "$status"
fi
```

## Pagination Pattern

```bash
token=""
while :; do
  args=(codeup repos list --organization-id "$ORG_ID" --page-size 100 --json)
  [ -n "$token" ] && args+=(--page-token "$token")

  out=$(yunxiao "${args[@]}" 2>diag)
  status=$?
  if [ "$status" -ne 0 ]; then
    jq '.error' <<<"$out" >&2
    exit "$status"
  fi

  jq -c '.data[]' <<<"$out"
  has_more=$(jq -r '.meta.pagination.has_more // false' <<<"$out")
  [ "$has_more" = "true" ] || break

  token=$(jq -r '.meta.pagination.next_token // empty' <<<"$out")
  if [ -z "$token" ]; then
    echo "pagination error: has_more=true but next_token is empty" >&2
    exit 1
  fi
done
```

Do not use this loop for `testhub testplans list` because it has no pagination contract. Consume its `.data` array directly and ignore `meta.pagination` for that command.

When you only need a count, first check `meta.pagination.total`. It is optional and appears only when the upstream API exposes total metadata, but it avoids fetching every page just to count matches.

## Error Handling

Use `.error.category` for control flow and `.error.code` for logging or exact diagnostics. Do not compare `.error.code` to category names such as `param` or `auth`.

| Category | Exit | Agent behavior |
|---|---:|---|
| `param` | 2 | Fix command/flags; do not retry. |
| `auth` | 3 | Provide or refresh `YUNXIAO_ACCESS_TOKEN`; do not retry blindly. |
| `not_found` | 4 | Check IDs and organization. |
| `rate_limit` | 5 | Back off; respect retry metadata/logs. |
| `network` | 6 | Retry safe read commands if appropriate; inspect timeout. |
| `upstream` | 7 | Retry may already be exhausted; check `retryable`. |
| `forbidden` | 8 | Fix permissions; do not retry blindly. |

## Common Mistakes

| Mistake | Correction |
|---|---|
| `yunxiao repo list --format json` | Use `yunxiao codeup repos list --json`. |
| Parsing stderr for errors | Parse stdout envelope `.error`; stderr is diagnostics. |
| Treating non-zero stdout as empty | Non-zero stdout still contains JSON envelope. |
| Paginating `testhub testplans list` | Do not pass pagination flags; consume returned array. |
| Retrying `PARAM_INVALID` or `AUTH_FAILED` blindly | Fix params or auth first. |
| Exiting before parsing non-zero stdout | Capture stdout, then parse `.error` from the JSON envelope. |
| Reading only top-level `.data[].path` from `commands --json` | Recursively walk `.subcommands[]` to discover leaf commands. |
| Using `raw request` for POST/PUT/DELETE | Do not; Phase 2 raw only supports GET. |
| Guessing `--page`, `--limit`, `--cursor` | Use `--page-size` and `--page-token` only where help exposes them. |

## Red Flags

Stop and inspect `commands --json` or `--help --json` if you are about to:

- Invent a command path or flag.
- Treat `commands --json` as a flat list instead of a command tree.
- Use stderr as structured output.
- Exit on a non-zero status before parsing stdout `.error`.
- Add pagination to a command whose help does not expose it.
- Use `raw request` to mutate data.
- Retry without checking `error.category` and `error.retryable`.
- Continue after `PARAM_INVALID` without changing arguments.

## Example

Need all Codeup repos for automation:

1. Confirm the command shape: `yunxiao codeup repos list --help --json`.
2. Run with `--json`, explicit `--organization-id`, and positive `--page-size`.
3. Parse stdout envelope. If `.error != null`, branch by `.error.category`.
4. Continue with `--page-token` only while `meta.pagination.has_more` is true.
