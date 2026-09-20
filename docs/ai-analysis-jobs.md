# Internal durable AI jobs

CH11 implements the [job lifecycle contract](../openspec/specs/ai-analysis-jobs/spec.md) with real SQLite persistence and a synthetic executor. It does not register a public job API or startup worker. CH12 must connect [Visual analysis](ai-visual-analysis.md) to frozen selections, current access, consent and production budgets. No job or stored analysis grants Immich write authority.

## Ownership and caller contract

`backend/internal/ai/jobs/` owns normalized inputs, bounded state transitions, safe failures and the one-item worker. Root `aiJob*.go` adapters own SQL, installation/profile binding and result persistence. Core code has no SQL driver, HTTP transport or concrete provider dependency. Migration `022_ai_analysis_jobs.sql` creates jobs, items and immutable analyses in the existing database.

`Submit` accepts already-admitted asset IDs. It requires enabled AI, the current installation identity, an enabled owner-qualified provider and an existing exact revision; the model is loaded from that revision. It freezes Visual mode, normalized requested/primary languages, a SHA-256 selection digest, the `visual-v1` consent version and a finite call cap. These shape checks do not establish that a user actually consented or still has image access. CH12 must derive the owner from authenticated server state and verify the retained selection, consent, profile/capabilities, asset access and source freshness before dispatch and publication.

At most 10,000 raw IDs may normalize to at most 500 distinct UUIDs. Deduplication preserves order. A job permits at most three calls per distinct item in total; a smaller positive job cap is allowed. Each item also permits at most three claims, including interrupted attempts that never reserved a call. Languages use the canonical CH07 normalization rules.

An idempotency key belongs to one owner and installation. The same normalized request returns the same retained job, concurrent submissions converge, and changed configuration or membership conflicts. A new key creates a separate run. Cleanup releases the deleted job's key; idempotency does not survive deliberate history deletion. No result-reuse cache is implemented.

## Claims, reservations and cancellation

All write operations acquire SQLite's write lock before their read/decide/update sequence and use a five-second operation context. Reads are also bounded. Transactions finish before executor work. Callers must use a consistent validated `jobs.Policy` across workers: defaults are two global active leases, one per owner and a 180-second lease. Allowed bounds are 1–16 global leases, 1–global per owner and a 1-second to 10-minute lease. No new operator configuration or scheduler is installed.

`Claim` chooses due work, increments the claim count and assigns an unpredictable token. Every guard, heartbeat, reservation and finalization checks the exact owner, installation, job, item, asset, token, running state and expiry. Active canceled/blocked attempts continue to consume concurrency until finalized or expired. Expiry fences local publication; it cannot prove that an external provider stopped processing.

`Reserve` durably increments both item and job counters before a provider call. Counters are never refunded, even if later admission or transmission fails. The last job reservation fails remaining undispatched items as budget-exhausted instead of leaving them queued forever. Active attempts that already reserved may still finish. A replacement lease must make its own reservation before success.

`Cancel` immediately cancels queued, retrying and blocked items, fences running guards and preserves completed results. Running work can finalize as canceled through the failure path or be reconciled after expiry. Late completion cannot publish a result. Cancellation cannot recall an already transmitted image or guarantee that no charge occurred.

## Recovery and executor behavior

`Worker.RunOne` recovers at most 100 expired items, claims at most one due item and invokes one injected executor under a 120-second context, honoring earlier caller deadlines. It supplies durable authorize/reserve callbacks bound to that context. The caller supplies heartbeat ticks; the intended default cadence is 30 seconds and must remain shorter than the chosen lease. The caller also owns the outer worker loop and shutdown.

The worker monitors renewal while execution runs. A heartbeat error, closed tick channel or canceled context cancels execution and prevents success. The executor must honor context and return; the worker joins it rather than abandoning an execution goroutine. Shutdown/storage faults can leave a running item for bounded expiry/recovery. No transaction spans execution.

Recovery clears expired tokens and schedules eligible interrupted work one second later. Canceled work becomes canceled, exhausted work fails and a systemic job block stays blocked. A process cannot claim another worker's unexpired lease. Claims and reservations independently bound repeated restarts and pre-dispatch failures.

| Executor outcome | Durable behavior |
|---|---|
| Valid completion | Atomically insert one immutable analysis and mark the item succeeded |
| Transient failure | Retry while budget remains; delay starts at 5 seconds and doubles by claim, with injected jitter clamped to 0–5 seconds and retry-after clamped to 0–5 minutes |
| Permanent or unrecognized failure | Fail this item with a safe `execution` category; unrelated work remains usable |
| Systemic authentication/policy block | Block the job's remaining undispatched work and fence active guards/publication; a later explicit run is required |
| Budget exhaustion | Fail the item without a hidden retry |
| Lost lease/storage authority | Return a safe operational error; leave unresolved work to recovery |

Raw executor/SQL errors are not stored or exposed by the adapter. Provider-specific classification and token/cost admission belong to CH12's integration. Dispatch is at least once within finite reservations; neither leases nor submission idempotency guarantee exactly-once billing.

## Results and private lifecycle

Completion requires a reservation on the current lease and a nonzero opaque CH07 proposal, revalidated against the stored Visual language contract. Located, ambiguous and unknown are distinct successful analysis outcomes. The result stores exact owner/job/item/asset binding, installation, profile revision/model, languages, selection/source/image digests and prompt/schema/validation versions. CH11 does not yet persist provider usage or the full planned context/review envelope.

Result insertion and the succeeded transition commit together. A SQLite trigger prevents updates to retained analyses. Reads return independent payload/metadata values; default Go formatting redacts an analysis record, but its explicit payload fields remain private data and must never be logged or serialized into telemetry. Image bytes, API keys, raw prompts and raw errors are not persisted. There is no catalog/snapshot foreign key that could erase history during synchronization.

Ordinary job/result reads require the current installation and enabled AI. Installation rotation retains old terminal history in SQLite but makes it unavailable through these current-installation readers. The existing installation binder cancels old unfinished work and clears leases in the same transaction as identity replacement. Follow the [controlled replacement procedure](ai-selection-snapshots.md#operator-settings-and-lifecycle): disable AI, change the endpoint/epoch as appropriate, restart, run full catalog sync for connected users and then enable AI. The epoch is an operator declaration; it does not automatically rebuild the catalog or attest remote identity.

Account deletion cascades private jobs, items and analyses. `PurgeBefore` accepts a trusted owner, a nonfuture cutoff and a batch size of 1–100. It deletes only jobs whose creation and every item's terminal completion precede that cutoff; eligible terminal states are succeeded, failed and canceled. Blocked jobs must first be canceled. Cleanup also covers retained old-installation history and works while AI is disabled. It never contacts Immich. No retention age, public cleanup route or automatic scheduler is enabled. Later draft/write references must restrict cleanup before they depend on this history.

Back up SQLite consistently before migration. Destructive down migrations are not an operational rollback strategy; restore a compatible backup when rollback is necessary. Keep the existing local SQLite deployment.

## Verification boundary

The [verification record](engineering/ai-analysis-jobs-verification.md) maps all 25 scenarios to automated tests. Synthetic execution and real file-backed SQLite establish lifecycle behavior, not real-model quality, deployed provider compatibility, reference-NAS concurrency/submission latency or production consent admission. Those remain separate integration/release evidence.
