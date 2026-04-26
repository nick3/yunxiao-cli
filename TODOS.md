# TODOS

## Verify Projex Phase 2 capability probe

**What:** Verify the Projex API capabilities needed for cross-project and personal workitem queries.

**Why:** Phase 1 fixes read-only project and project-scoped workitem search, but user scenarios still need answers for "my unfinished tasks", "tasks completed today/this week", and "projects I participate in".

**Pros:**
- Confirms whether `workitems:search` supports organization-level search without `spaceId` before adding a public `projex workitems search` command.
- Confirms whether `org current` user IDs work for `assignedTo` / `creator` filters.
- Confirms the correct completion-time field before adding date filters.
- Prevents public flags from being based on MCP assumptions or mock-only behavior.

**Cons:**
- Requires real-token manual verification with strict desensitization.
- Does not ship new user-facing commands by itself.
- May reveal API limitations that require a smaller Phase 2 command design.

**Context:**
Phase 1 deliberately excludes cross-project search, assignee/status/date filters, and write operations. After Phase 1 lands, run real Yunxiao probes and record only desensitized summaries: command shape, exit code, `.error.code`, `.error.category`, data type, item count, field keys, and pagination keys. Do not record organization names, project names, workitem titles, descriptions, people names, tokens, or full IDs.

Capability probes:
- Does `workitems:search` allow omitting `spaceId` for organization-level or cross-project search?
- Does `org current.data.id` work as a Projex `assignedTo` / `creator` value?
- Is completion best represented by `finishTime`, `updateStatusAt`, status ID, or workflow final state?
- Which project participation filter is valid: `users`, `project.admin`, `collectMembers`, or another field?

**Depends on / blocked by:** Phase 1 raw-array search decode fix, Projex pagination header mapping, and real `projex projects list` verification.
