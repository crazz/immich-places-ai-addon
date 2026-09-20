## 1. Persist exact private submissions without duplication

- [x] 1.1 Preserve normalized execution configuration and deduplicated membership across reopen, while rejecting malformed, disabled, stale or foreign submissions atomically.
- [x] 1.2 Make concurrent identical submissions converge, reject changed-key reuse and preserve separate explicit runs.
- [x] 1.3 Upgrade existing databases without changing catalog/provider/selection data and enforce scoped persistence constraints across pooled connections.

## 2. Execute and publish only under current bounded authority

- [x] 2.1 Claim due work under global/per-owner capacity, renew only current leases and durably enforce item/job dispatch and execution-attempt caps.
- [x] 2.2 Atomically retain only immutable canonical results with exact provenance, preserving unknown outcomes, tenant isolation and history after source removal.

## 3. Recover interrupted work and isolate execution failures

- [x] 3.1 Recover expired leases in bounded batches, reject stale workers and terminate repeated interruptions at the configured bounds.
- [x] 3.2 Cancel undispatched and active work without late publication, and make the synthetic worker honor context/heartbeat authority loss.
- [x] 3.3 Retry transient failures with bounded scheduling, terminate permanent failures independently and block systemic job failures without further calls.

## 4. Close installation and private-history lifecycle boundaries

- [x] 4.1 Fence unfinished work on existing installation rotation and cascade private lifecycle data on account deletion.
- [x] 4.2 Provide bounded explicit terminal-history cleanup without an automatic retention policy or external mutations.
- [x] 4.3 Document the internal caller and recovery/retention limits, map all scenarios to tests and complete shared verification, packaging and graph review while preserving legacy workflows.
