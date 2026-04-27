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
| Writes | Projex write commands require explicit `--yes`; never use raw request for writes. |

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
yunxiao projex project-templates list --organization-id <org> --json
yunxiao projex project-template fields --organization-id <org> --template-id <template> --json
yunxiao projex project create --organization-id <org> --name <test-project-name> --custom-code <CODE> --template-id <template> --scope private --description "Created by yunxiao-cli smoke test" --yes --json
yunxiao projex project archive --organization-id <org> --project-id <project> --yes --json
yunxiao projex workitems list --organization-id <org> --category <category> --project-id <project> --json
yunxiao projex workitems list --mine --unfinished --category Task --json
yunxiao projex workitem get --organization-id <org> --workitem-id <workitem> --json
yunxiao projex workitem create --organization-id <org> --project-id <project> --workitem-type-id <type> --subject <subject> --assigned-to self --yes --json
yunxiao projex workitem update --organization-id <org> --workitem-id <workitem> --status <status> --assigned-to self --yes --json
yunxiao projex workitem comments list --organization-id <org> --workitem-id <workitem> --json
yunxiao projex workitem comment create --organization-id <org> --workitem-id <workitem> --content <content> --yes --json
yunxiao projex workitem-types list --organization-id <org> --project-id <project> --category Task --json
yunxiao projex workitem-types relations --organization-id <org> --workitem-type-id <type> --json
yunxiao projex workitem-type get --organization-id <org> --workitem-type-id <type> --json
yunxiao projex workitem-type fields --organization-id <org> --project-id <project> --workitem-type-id <type> --json
yunxiao projex workitem-type workflow --organization-id <org> --project-id <project> --workitem-type-id <type> --json
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

Projex commands support project/space-scoped list enumeration, known-ID detail reads, project template discovery, safe test project lifecycle writes, workitem type metadata, comments, and explicit-confirmation workitem writes. `projex projects list` can auto-resolve the central-edition organization from `org current.lastOrganizationId`, supports MCP-compatible filters such as `--name`, `--status`, `--admin-user-id`, `--scenario-filter`, `--user-id`, `--advanced-conditions`, and `--extra-conditions`, and exposes `--mine` for projects participated in by the current user. `projects list --mine` is mutually exclusive with explicit `--scenario-filter` / `--user-id`; `--advanced-conditions` overrides basic project filters, while `--scenario-filter` plus `--user-id` overrides `--extra-conditions`. `projex project-templates list` and `projex project-template fields` are required before project creation because `template-id` must come from upstream data or the user. `projex project create` and `projex project archive` write to Yunxiao and must include `--yes`; use them only for clearly named private test projects unless the user explicitly confirms another target. Plain `projex workitems list` requires explicit `--organization-id`, `--category`, and `--project-id` or `--space-id`, and exposes MCP-compatible filters such as `--status`, `--assigned-to`, `--finish-time-after`, `--update-status-at-after`, `--order-by`, and `--sort`; prefer `--project-id` in Agent workflows and use `--space-id` only for compatibility with API wording. `projex workitems list --mine --category Task` resolves the current user, enumerates participated projects, and returns workitems assigned to that user across those projects; add `--unfinished` to filter completed workitems from that aggregated result. In `--mine` mode, `--project-id`, `--space-id`, `--page-token`, `--assigned-to`, and `--advanced-conditions` are rejected with `PARAM_INVALID`. `--unfinished` is only valid with `--mine` and fails instead of guessing when a workitem completion status is not recognizable. `projex workitem create`, `projex workitem update`, and `projex workitem comment create` write to Yunxiao and must include `--yes`; without `--yes` they fail before auth or HTTP. `workitem create` accepts `--project-id` or `--space-id`; if both are set they must match. `workitem create/update --assigned-to self` resolves the current user ID before sending the write. `--description-file` and `--content-file` read UTF-8 regular files only, reject empty files, and cap input at 1MiB. Create sends `customFieldValues` as a nested object; update expands `--custom-field` / `--custom-fields-json` entries to top-level fields, matching the official Yunxiao MCP server. Do not assume direct organization-level workitem search or this-week completion aggregation are public CLI contracts. Use `commands --json` and `--help --json` before calling Projex commands, and treat unsupported business filters as capability gaps.

## Projex Safe Smoke Test Playbook

Use this playbook when validating whether Yunxiao CLI and this skill can complete common Projex workflows without touching real project data.

1. Discover identity and organization:

   ```bash
   yunxiao auth status --json
   yunxiao org current --json
   ```

2. Discover current read context without writing:

   ```bash
   yunxiao projex projects list --mine --json
   yunxiao projex workitems list --mine --unfinished --category Task --json
   yunxiao projex workitems list --mine --unfinished --category Req --json
   yunxiao projex workitems list --mine --unfinished --category Bug --json
   ```

3. Read a real workitem only when the user has approved that read target:

   ```bash
   yunxiao projex workitem get --organization-id <org> --workitem-id <workitem> --json
   yunxiao projex workitem comments list --organization-id <org> --workitem-id <workitem> --json
   ```

4. Discover templates before creating a project. Do not guess `template-id`:

   ```bash
   yunxiao projex project-templates list --organization-id <org> --json
   yunxiao projex project-template fields --organization-id <org> --template-id <template> --json
   ```

5. Before creating the test project, present the exact target and payload to the user: organization, template, name, custom code, scope, description, and custom fields. Continue only after explicit confirmation. Use an obvious test name prefix and `--scope private`:

   ```bash
   yunxiao projex project create \
     --organization-id <org> \
     --name "yunxiao-cli-test-<timestamp>" \
     --custom-code <CODE> \
     --template-id <template> \
     --scope private \
     --description "Created by yunxiao-cli smoke test" \
     --yes \
     --json
   ```

6. Write only inside the test project. Before every command that includes `--yes`, preview the exact target and payload and wait for explicit confirmation:

   ```bash
   yunxiao projex workitem-types list --organization-id <org> --project-id <project> --category Task --json
   yunxiao projex workitem create --organization-id <org> --project-id <project> --workitem-type-id <type> --subject <subject> --assigned-to self --yes --json
   yunxiao projex workitem comment create --organization-id <org> --workitem-id <workitem> --content <content> --yes --json
   yunxiao projex workitem update --organization-id <org> --workitem-id <workitem> --subject <subject> --yes --json
   ```

7. Archive the test project as cleanup:

   ```bash
   yunxiao projex project archive --organization-id <org> --project-id <project> --yes --json
   ```

8. If archive fails, report the project ID and the structured `.error` object. Do not attempt hard delete as a fallback, and do not use `raw request` for write fallback.

## Projex Workitem Write Playbook

Use this playbook when the user asks to create or update a Yunxiao requirement, task, defect, bug, or work item.

1. Discover capabilities first: run `yunxiao projex workitem create --help --json`, `yunxiao projex workitem-types list --help --json`, and any needed `projects list` / `workitem-types list` command. Do not invent flags.
2. Map Chinese work item words to Projex categories for lookup: `需求` -> `Req`, `任务` -> `Task`, `缺陷` / `Bug` -> `Bug`. If the user gives a different category, use their explicit value.
3. Resolve target IDs before writing: organization ID, project ID, workitem type ID, subject, and assignee. Prefer `--project-id`; use `--space-id` only when the user or existing data names it that way. Use `--assigned-to self` when the user wants the current account as assignee.
4. Before adding `--yes`, present the exact target and payload to the user: organization, project, category/type, subject, assignee, description, parent, labels, sprint, custom fields, and whether this creates or updates data. Continue only after explicit confirmation.
5. Execute the write with `--json` and `--yes`. Capture stdout even on non-zero exit and branch on `.error.category`.
6. After create succeeds, run `yunxiao projex workitem get --organization-id <org> --workitem-id <id> --json` to verify the created item when the response contains an ID. Report the ID, serial number if present, subject, status, project, workitem type, and URL only if the CLI or upstream response provides one.
7. Recovery: on `param`, fix flags or missing IDs; on `auth`, ask for token refresh; on `forbidden`, stop and report permissions; on `upstream`, report upstream message and retry only when `.error.retryable` is true.

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
| Running Projex create/update/comment create without `--yes` | Add `--yes` only after confirming the write target and payload. |
| Guessing `--page`, `--limit`, `--cursor` | Use `--page-size` and `--page-token` only where help exposes them. |

## Red Flags

Stop and inspect `commands --json` or `--help --json` if you are about to:

- Invent a command path or flag.
- Treat `commands --json` as a flat list instead of a command tree.
- Use stderr as structured output.
- Exit on a non-zero status before parsing stdout `.error`.
- Add pagination to a command whose help does not expose it.
- Use `raw request` to mutate data.
- Add `--yes` to a Projex write command without confirming the target organization, workitem, and payload.
- Retry without checking `error.category` and `error.retryable`.
- Continue after `PARAM_INVALID` without changing arguments.

## Example

Need all Codeup repos for automation:

1. Confirm the command shape: `yunxiao codeup repos list --help --json`.
2. Run with `--json`, explicit `--organization-id`, and positive `--page-size`.
3. Parse stdout envelope. If `.error != null`, branch by `.error.category`.
4. Continue with `--page-token` only while `meta.pagination.has_more` is true.
