<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# shared

## Purpose
This package contains shared domain helpers for JSON HTTP requests, response decoding, upstream error conversion, pagination payloads, base URL mode detection, and token normalization.

## Key Files
| File | Description |
|------|-------------|
| `http.go` | Sends JSON requests through `httpx.Client`, decodes responses, maps network/read/decode failures to `output.ErrorDetail`, and exposes base URL/token helpers. |
| `pagination.go` | Defines search-style response shape and applies page tokens to POST search payloads. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Shared helper package. |

## For AI Agents

### Working In This Directory
- Keep helpers side-effect free except for HTTP calls; do not print or exit from domain helpers.
- Preserve non-2xx handling through `httpx.ClassifyHTTPError` so error categories stay consistent.
- Treat empty JSON response bodies as errors for JSON requests.

### Testing Requirements
- Run domain and integration tests after changing request, decode, or pagination helper behavior.
- Add focused unit tests when adding helper branches that are not covered by integration tests.

### Common Patterns
- `RequestJSONWithBodyAndHeaders` returns response headers so domains can read pagination headers.
- `StringToken` normalizes `nextPage` values from `nil`, string, number, or other JSON values.

## Dependencies

### Internal
- `internal/httpx/` for transport and HTTP error classification.
- `internal/model/output/` for error details.

### External
- Go standard `encoding/json`, `net/http`, and `io` packages.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
