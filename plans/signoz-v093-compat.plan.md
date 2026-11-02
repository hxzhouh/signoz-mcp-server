# SigNoz v0.93 compatibility plan

## Status
In Progress

## Approach
- Discover available trace attributes before selecting optional fields; preserve intrinsic fields and propagate authorization failures.
- Diagnose and adapt the metrics catalog endpoint without accepting HTML as a successful response.
- Fall back to v1 alert and dashboard reads without translating legacy write contracts.
- Check read-only tools iteratively and use mock-based tests for mutation paths.

## Verification
- Run focused tests, the full Go test suite, and build the server.
- Delegate read-only verification against the configured instance, then correct any confirmed incompatibilities.
