<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-26 | Updated: 2026-04-26 -->

# httpx

## Purpose
This package provides the Yunxiao HTTP client used by command and domain layers. It handles base URL concatenation, token headers, trace IDs, request timeout, retry behavior, Retry-After handling, and HTTP status classification into CLI error details.

## Key Files
| File | Description |
|------|-------------|
| `client.go` | Defines `Client`, request execution, retry/backoff logic, and HTTP error-to-category mapping. |
| `client_test.go` | Unit tests for retry, timeout, and error classification behavior. |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| _(none)_ | Shared HTTP transport package. |

## For AI Agents

### Working In This Directory
- Never print or expose token values from request headers.
- Keep retries limited to idempotent methods (`GET`, `HEAD`) unless the public contract changes.
- Preserve status-to-category mapping: 401 auth, 403 forbidden, 404 not_found, 429 rate_limit, 5xx upstream.

### Testing Requirements
- Run `go test ./internal/httpx -count=1` after transport changes.
- Run integration tests for timeout, rate-limit, upstream, and forbidden scenarios after classification changes.

### Common Patterns
- `Client.Do` builds requests as `BaseURL + path` and sets `x-yunxiao-token`.
- Retry logs go to stderr unless quiet mode is active.

## Dependencies

### Internal
- `internal/model/output/` for `ErrorDetail` classification results.

### External
- Go standard `net/http`, `time`, and context-aware request handling.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
