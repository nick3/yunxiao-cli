# Projex Smoke Test Lifecycle Design

## Purpose

Use real user tasks to test whether Yunxiao CLI and `skill/using-yunxiao-cli/SKILL.md` can guide an AI Agent through common Projex workflows without guessing commands, writing to real projects by accident, or losing structured error information.

The central gap found during brainstorming is that safe write testing needs an isolated Projex project, but the CLI currently lacks project creation and archival commands. This design adds the minimum project lifecycle needed for safe smoke tests, then uses that lifecycle to validate workitem and comment flows.

## Scope

In scope:

- Discover current organization and participated projects.
- Query the current user's unfinished workitems by category.
- Read real workitem details and comments without modifying them.
- List project templates and template field config.
- Create a clearly named private test project.
- Create, update, and comment on workitems only inside the test project.
- Archive the test project as cleanup.
- Update README, ADR, command introspection tests, integration tests, and `SKILL.md` playbooks.

Out of scope:

- Hard delete of projects.
- Unarchive support.
- Project member management.
- Automatic template selection without user or upstream evidence.
- Broad project creation features unrelated to smoke testing.

## User Scenario Matrix

| Scenario | User intent | CLI coverage target |
|---|---|---|
| Discover context | Know current org, projects, and project IDs | `org current`, `projex projects list --mine`, `projex project get` |
| View my work | Find unfinished Req/Task/Bug items assigned to me | `projex workitems list --mine --unfinished --category <category>` |
| Inspect a workitem | Read fields, status, assignee, description, and comments | `projex workitem get`, `projex workitem comments list` |
| Prepare safe write area | Create a private test project with a test name prefix | new `projex project create` plus template discovery |
| Write safely | Create/update/comment only inside the test project | existing workitem create/update/comment commands |
| Cleanup safely | Archive test project after smoke test | new `projex project archive` |
| Recover from failures | Branch by JSON `error.category` and report IDs | existing stdout envelope contract plus SKILL playbook |

## Projex Project Lifecycle Commands

### `projex project create`

Shape:

```bash
yunxiao projex project create \
  --organization-id <org> \
  --name <name> \
  --custom-code <CODE> \
  --template-id <template> \
  --scope private \
  --description <text> \
  --yes \
  --json
```

Required flags:

- `--organization-id`
- `--name`
- `--custom-code`
- `--template-id`
- `--scope`
- `--yes`

Optional flags:

- `--description`
- `--description-file`
- `--custom-field key=value`, repeatable
- `--custom-fields-json <json-object>`

Validation:

- `--yes` is required before auth or HTTP because this writes to Yunxiao.
- `--scope` must be `public` or `private`.
- `--custom-code` must be 4 to 6 uppercase ASCII letters, matching the Swagger constraint.
- `--description` and `--description-file` are mutually exclusive.
- `--description-file` follows existing UTF-8 regular-file and 1 MiB safety rules.
- `--custom-fields-json` must decode to a JSON object.
- Duplicate custom field keys between pair and JSON inputs return `PARAM_INVALID`.

Domain request:

- `POST /oapi/v1/projex/organizations/{organizationId}/projects`
- Payload includes `name`, `customCode`, `scope`, `templateId`, optional `description`, and optional `customFieldValues`.

Response:

- Accept resource object with `id`.
- Accept `{ "result": { "id": "..." } }`.
- Surface upstream business errors through the JSON envelope.

### `projex project archive`

Shape:

```bash
yunxiao projex project archive \
  --organization-id <org> \
  --project-id <project> \
  --yes \
  --json
```

Required flags:

- `--organization-id`
- `--project-id`
- `--yes`

Optional flags:

- `--operator-id`; documented as ineffective for personal tokens, so the SKILL playbook should not pass it by default.

Domain request:

- `POST /oapi/v1/projex/organizations/{organizationId}/projects/{id}/archived`
- Payload is empty or contains `operatorId` when explicitly set.

Response:

- Accept resource object with `id` when provided.
- Accept explicit success confirmations such as `{ "success": true }` and `{ "success": true, "requestId": "..." }`.
- Do not add delete fallback if archive fails.

## Template Discovery Commands

Project creation requires `templateId`, so the CLI must expose template discovery instead of making agents guess.

### `projex project-templates list`

```bash
yunxiao projex project-templates list --organization-id <org> --json
```

Domain request:

- `GET /oapi/v1/projex/organizations/{organizationId}/projectTemplates`

Response:

- Decode an array or result-wrapped array.

### `projex project-template fields`

```bash
yunxiao projex project-template fields \
  --organization-id <org> \
  --template-id <template> \
  --json
```

Domain request:

- `GET /oapi/v1/projex/organizations/{organizationId}/projectTemplates/{id}/fields`

Response:

- Decode object or result-wrapped object.

## Smoke Test Playbook

1. Discover identity and organization:

   ```bash
   yunxiao auth status --json
   yunxiao org current --json
   ```

2. Discover current work context:

   ```bash
   yunxiao projex projects list --mine --json
   yunxiao projex workitems list --mine --unfinished --category Task --json
   yunxiao projex workitems list --mine --unfinished --category Req --json
   yunxiao projex workitems list --mine --unfinished --category Bug --json
   ```

3. Read a real workitem without writing:

   ```bash
   yunxiao projex workitem get --organization-id <org> --workitem-id <id> --json
   yunxiao projex workitem comments list --organization-id <org> --workitem-id <id> --json
   ```

4. Prepare test project:

   ```bash
   yunxiao projex project-templates list --organization-id <org> --json
   yunxiao projex project-template fields --organization-id <org> --template-id <template> --json
   yunxiao projex project create \
     --organization-id <org> \
     --name "yunxiao-cli-test-<timestamp>" \
     --custom-code <CODE> \
     --scope private \
     --template-id <template> \
     --description "Created by yunxiao-cli smoke test" \
     --yes \
     --json
   ```

5. Write only inside the test project:

   ```bash
   yunxiao projex workitem-types list --organization-id <org> --project-id <project> --category Task --json
   yunxiao projex workitem create --organization-id <org> --project-id <project> --workitem-type-id <type> --subject <subject> --assigned-to self --yes --json
   yunxiao projex workitem comment create --organization-id <org> --workitem-id <workitem> --content <content> --yes --json
   yunxiao projex workitem update --organization-id <org> --workitem-id <workitem> --subject <subject> --yes --json
   ```

6. Cleanup:

   ```bash
   yunxiao projex project archive --organization-id <org> --project-id <project> --yes --json
   ```

If archive fails, the agent reports the project ID and structured error. It must not attempt hard delete as a fallback.

## Documentation and Skill Updates

README updates:

- Add project template discovery examples.
- Add project create and archive examples.
- State that archive is the supported cleanup path for smoke tests.

`skill/using-yunxiao-cli/SKILL.md` updates:

- Add a Projex Safe Smoke Test Playbook.
- Require template discovery before project creation.
- Require obvious test name prefix, private scope, and explicit write confirmation.
- Require payload preview before every `--yes` write.
- Require project archival after smoke testing.
- State that failed archive must be reported, not replaced with delete.
- Keep raw request read-only and forbidden for write fallback.

ADR updates:

- Add the new command shapes if the public CLI contract section enumerates Projex commands.
- Preserve the existing grammar style: plural resources list collections; singular resources act on individual objects.

## Code Organization

Domain layer: `internal/domains/projex/projects.go`

- `CreateProject`
- `ArchiveProject`
- `ListProjectTemplates`
- `GetProjectTemplateFields`
- `projectTemplatesPath`
- `projectTemplateFieldsPath`

Command layer: `internal/command/projex/projects.go`

- Add `create` and `archive` under `project`.
- Add `project-templates list` collection command.
- Add `project-template fields` single-resource metadata command.
- Keep parameter validation and `--yes` enforcement in command layer.

Shared decode behavior:

- Prefer Projex-local helper names that describe generic resource decoding.
- Avoid adding workitem-specific dependencies to project/template code.

## Testing Strategy

Update command introspection tests:

- New command paths appear in `commands --json`.
- Help JSON exposes required and optional flags.

Add integration tests:

- `project create` sends the expected payload and decodes ID response.
- `project create` rejects missing `--yes` before auth/HTTP.
- `project create` rejects invalid `custom-code` before auth/HTTP.
- `project create` rejects invalid `scope` before auth/HTTP.
- `project create` supports `--description-file` with existing file safety rules.
- `project archive` POSTs to `/archived` and accepts explicit success confirmation.
- `project archive` rejects missing `--yes` before auth/HTTP.
- `project-templates list` decodes arrays and result-wrapped arrays.
- `project-template fields` decodes objects and result-wrapped objects.
- Upstream business errors remain structured JSON errors.

Verification after implementation:

```bash
gofmt -w <touched-go-files>
rtk go test ./test/integration -run 'TestProjexProject|TestCommands'
test -z "$(gofmt -l cmd internal test)"
rtk go vet ./...
rtk go test ./...
rtk go build -o yunxiao ./cmd/yunxiao
```

## Design Principles

- KISS: implement only project lifecycle and template reads needed for safe smoke testing.
- YAGNI: no delete, unarchive, member management, or automatic template choice.
- DRY: reuse existing text-file and custom-field parsing patterns where they fit, without forcing broad abstraction.
- SOLID: command layer owns CLI validation and confirmation; domain layer owns HTTP paths, payloads, and response decoding.

## Open Follow-Ups

- Decide after real smoke tests whether template selection can be made friendlier without guessing.
- Decide separately whether to add project delete for human-admin workflows. It is not part of the Agent smoke-test path.
- Review existing command grammar concerns around `relations`, `fields`, and `workflow` in a separate compatibility-focused change.
