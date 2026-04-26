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

Never guess aliases such as `repo list`, `testplans --page`, `--format json`, `--limit`, or `--cursor`.

## Core Contract

| Area | Rule |
|---|---|
| Machine output | Pass `--json`; parse stdout only. |
| Failure output | Non-zero exits still write JSON envelope to stdout. |
| Diagnostics | stderr is only diagnostics; never use stderr as the structured error source. |
| Auth | Automation should use `YUNXIAO_ACCESS_TOKEN`; missing token maps to `AUTH_FAILED` / exit 3. |
| Organization | Prefer explicit `--organization-id`; env/config fallback may exist. |
| Page size | `--page-size` must be positive integer; invalid values are `PARAM_INVALID` / exit 2 before auth. |
| Pagination | Follow `meta.pagination.next_token` only when `meta.pagination.has_more` is true. |
| Testplans | `testhub testplans list` is not paginated; do not pass `--page-size` or `--page-token`. |
| Raw request | Phase 2 raw is read-only: only `--method GET` and `--path /oapi/...`. |

## Auth Guidance

Use `YUNXIAO_ACCESS_TOKEN` for automation. Human users may run `yunxiao auth` in a private real terminal for visible interactive token entry. Agents, CI, and scripts must not trigger the interactive path; use env or explicit stdin instead:

```bash
printf '%s\n' "$YUNXIAO_ACCESS_TOKEN" | yunxiao auth login --token-stdin --force --json
yunxiao auth status --json
yunxiao auth status --verify --json
yunxiao auth logout --json
```

Only add `--skip-verify` for explicit offline/internal-network cases; it saves an unverified token. Never pass tokens as command-line arguments. `YUNXIAO_ACCESS_TOKEN` takes precedence over config, and `auth logout` only removes the config token.

## Command Quick Reference

```bash
yunxiao auth status --json
yunxiao org current --json
yunxiao codeup repos list --organization-id <org> --json
yunxiao codeup repo get --organization-id <org> --repo-id <repo> --json
yunxiao flow pipelines list --organization-id <org> --json
yunxiao flow pipeline get --organization-id <org> --pipeline-id <pipeline> --json
yunxiao projex projects list --organization-id <org> --json
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
  out=$(yunxiao "${args[@]}") || exit $?
  jq -c '.data[]' <<<"$out"
  has_more=$(jq -r '.meta.pagination.has_more // false' <<<"$out")
  [ "$has_more" = true ] || break
  token=$(jq -r '.meta.pagination.next_token // empty' <<<"$out")
  [ -n "$token" ] || break
 done
```

Do not use this loop for `testhub testplans list` because it has no pagination contract.

## Error Handling

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
| Using `raw request` for POST/PUT/DELETE | Do not; Phase 2 raw only supports GET. |
| Guessing `--page`, `--limit`, `--cursor` | Use `--page-size` and `--page-token` only where help exposes them. |

## Red Flags

Stop and inspect `commands --json` or `--help --json` if you are about to:

- Invent a command path or flag.
- Use stderr as structured output.
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
