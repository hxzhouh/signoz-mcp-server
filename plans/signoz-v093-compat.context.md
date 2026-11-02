# SigNoz v0.93 compatibility — context

## Original request
Fork the MCP server to the user's GitHub, adapt it to the existing SigNoz v0.93.0 instance, and iteratively verify compatibility.

## Discussion log
### 2026-11-02
- The local server is based on v0.10.0. Trace search and detail fail when optional fields absent from the instance are selected; an equivalent query with discovered fields succeeds.
- The metrics listing endpoint returns HTML instead of a metrics response on this instance.
- Live verification must be read-only. Administrative writes are covered by isolated tests, not production mutations.

### 2026-11-02 — verification
- Trace search/detail, metric catalog, top metrics, gauge/sum/histogram queries, and legacy dashboard and alert reads passed live read-only verification.
- SigNoz v0.93 rejects percentile queries with explicit temporality; the backend resolves temporality when omitted. Retry only that specific validation failure and disclose the fallback.
- Legacy dashboard schemas differ from v2. Mutating v2 dashboard and alert contracts are not translated or verified against the legacy instance.
- Legacy alert history supports offset pagination; use explicit v1-backed cursors and reject unsupported filter expressions.
- The old metric metadata endpoint cannot reliably answer time-bounded cardinality or complete dependency queries; those operations remain explicitly unsupported rather than returning misleading empty results.

## Open questions
- [ ] Which administrative mutation and optional integration APIs are supported by the deployed instance?
