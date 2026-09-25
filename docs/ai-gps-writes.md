# Confirmed AI GPS writes

CH19 adds explicit, durable confirmation of one reviewed GPS pair for one photo. Ordinary dispatch remains disabled: live GATE-02 compatibility has not been verified. Analysis, local draft acceptance/editing and preview creation still make no Immich mutations.

## Review and recovery

1. Review the current photo and GPS, acknowledge the baseline, stage the saved draft and create an [exact preview](ai-write-previews.md).
2. Inspect the photo, approved revision, before/intended values and five-minute expiry. Resolve any overlapping manual pending edit explicitly. **Confirm GPS write** authorizes only that pair and photo. Descriptions, direction, timestamps, rating, favorite, sidecars and stack siblings are outside this operation.
3. If confirmation acknowledgement is lost, **Reconcile confirmation** retrieves the same submission identity. **Retry same confirmation** explicitly repeats that same identity and plan; it does not create a second operation. Recovery works on HTTP origins without `crypto.randomUUID`. The browser retains only the unresolved identity, preview ID and digest in account/draft-scoped session storage.
4. Inspect status and private audit independently of the current result list or Missing GPS membership. **Check status** requests read-only reconciliation; it never grants another mutation. **Retry GPS write** appears only for a completed unsuccessful attempt and requires explicit action, unchanged fresh baseline, current revision/authority and remaining budget.

Late initial history cannot replace a newer confirmation or discard its unresolved identity. A definitive rejection of **Retry same confirmation**, such as an expired preview, releases that rejected identity so a fresh comparison can be explicitly confirmed. A missing lookup or uncertain repeat keeps the same identity and blocks another approval.

Unsaved edits survive a `WRITE_IN_PROGRESS` response. Queued or safely retryable approvals are canceled atomically when a new revision is saved or rejected. A reserved or possibly active request keeps its approved revision fixed. Account/result navigation fences late private responses, and independent manual pending edits remain intact.

A successful HTTP response is not a saved label. Source-consistent readback must observe the intended pair within **1e-7 degrees per coordinate**. This is numerical serialization tolerance, not geographic accuracy or causal proof. An unchanged plan is verified without sending a mutation. Missing local rows and failed local publication remain **local refresh pending**; no asset is fabricated and reconciliation does not resend to repair local state. Verified catalog publication precedes marker/count refresh. Failure to refresh the page offers a local refresh action.

Observed desired GPS and sender completion are separate facts. An unknown sender can leave GPS verified while continuing to block another write and revision changes. Seeing the old baseline does not establish that a timed-out request cannot still act. There is no automatic override for this exclusion.

## Authority, transport and bounds

- `POST /ai/write-operations` accepts only `previewId`, `digest` and `idempotencyKey`. Session/origin checks and a 4 KiB strict body apply. A short SQLite transaction revalidates the canonical server plan, protects/consumes the preview and creates immutable approval, target exclusion and audit before dispatch.
- Private `GET /ai/write-operations/{id}` and `/by-key/{key}` return the stored operation without upstream I/O. `GET /ai/write-operations?draftId=...&before=...` pages 100 private summaries. Reads remain available when execution is disabled. Foreign or obsolete installation references do not disclose private values.
- `POST /ai/write-operations/{id}/reconcile` accepts `{}` and renews only the exhausted read budget. `POST /ai/write-operations/{id}/retry` accepts the inspected `generation`; concurrent requests cannot allocate additional attempts. No operation may reserve more than two mutation attempts.
- Repeating an already-accepted retry generation returns its current private operation locally, even after success, restart, expiry or disablement. It performs no new upstream read or attempt allocation. A duplicate already reading metadata also recovers a concurrently accepted retry; new retries still require every fresh eligibility check.
- The dedicated `immich-v3.2.2` adapter sends one `PATCH /api/assets/{approved-id}` containing only finite `latitude` and `longitude`. Zero is a coordinate. It has no redirects, hidden retries, alternate route, replay body or legacy stack-expanding path. Response headers are bounded to 16 KiB, drained bodies to 64 KiB and mutation I/O to 20 seconds. Read metadata is bounded to 1 MiB and ten seconds.
- Current installation, owner, credentials, local visibility, reviewed image, exact nullable baseline, revision, expiry and enablement are checked before reservation; local authority and worker lease are checked again immediately before send. Another client can still change Immich between the final read and write: there is no remote compare-and-swap or exactly-once guarantee.
- One OS lock in the data directory admits a writer runtime. Failure to acquire it disables dispatch. Two global workers, one active operation per owner and installation/asset exclusion bound concurrency. A worker turn/lease is 45 seconds. Expired or restarted work reconciles through reads; expiry alone never permits resend. Automatic recovery allows three reads with persisted 1/5/15-second delays. UI polling performs at most 30 reads at two-second intervals; later checks are explicit.
- Verified publication drains older user sync work through the existing pause lock, with a five-second deadline, before final readback and local transaction. Failure leaves reconciliation pending. Shutdown cancels I/O, retains reservations and waits up to seven seconds for the writer before database teardown.

## Operator controls and live release gate

`AI_WRITE_ENABLED` defaults to `false`; `AI_WRITE_PROFILE` defaults to empty. Dispatch additionally requires `AI_ENABLED=true`, the exact profile `immich-v3.2.2` and the runtime lock. Unsupported profiles fail closed. These are startup controls; restart to change them. Disabling prevents new sends and cancels queued approvals while preserving read-only recovery of possibly sent work. It cannot retract a request already received by Immich.

Keep ordinary `AI_WRITE_ENABLED=false` until separately authorized GATE-02 evidence records all of the following for the actual configured endpoint/version:

- An explicitly disposable synthetic photo, update rights and authorization to mutate it. Analysis-only photographs are not write fixtures.
- Read-only before values and unrelated metadata, the exact approved GPS pair and one observed mutation on the supported route.
- Source-consistent API readback, unchanged unrelated fields/siblings and the corresponding private operation audit.

No live runner is added or invoked by CH19. The automated browser fixture is loopback-only and rejects write enablement without its explicit disposable-synthetic marker; tests cannot establish actual Immich compatibility. A future live check requires explicit authorization and a separately recorded procedure/evidence before enabling ordinary dispatch. Do not commit private IDs, coordinates, credentials or photo payloads to a public report.

## Persistence and rollback

Migration 028 adds immutable operation approvals, fenced target state, bounded append-only events and opaque target exclusion. Owner/installation-qualified references retain exact preview/draft/history through restart, cleanup and catalog reset. Each operation retains at most 100 audit events; read checks never erase attempt history. Account deletion cascades private records and prevents late recreation. A possibly active deleted sender retains only opaque target exclusion; definite completion can retire that exclusion. Installation rotation never transfers old approval to the new installation.

Take a consistent SQLite backup before deployment. Roll back binaries with additive tables retained; destructive down migrations remove audit and are not recovery. No rollback claims to undo upstream GPS, clear GPS or finish sidecars. Deployment and live rollback were not performed. See the [CH19 verification record](../openspec/changes/execute-confirmed-ai-gps-writes/implementation-verification.md) for executed synthetic checks and remaining limits.

The [recovery follow-up verification](../openspec/changes/execute-confirmed-ai-gps-writes/recovery-verification.md) records late-history, repeat-rejection and durable retry replay regressions.
