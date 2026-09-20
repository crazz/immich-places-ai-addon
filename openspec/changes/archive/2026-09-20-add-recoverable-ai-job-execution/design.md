## Context

See [proposal.md](proposal.md), the [architecture](../../../../docs/engineering/architecture.md), [testing](../../../../docs/engineering/testing.md), [coding standards](../../../../docs/engineering/coding-standards.md) and ADR-07. CH05 already binds the configured endpoint/operator epoch to one persistent installation UUID and deletes stale snapshots on rotation. CH07 provides opaque validated proposals; CH09 provides an inactive internal Visual attempt. Migration 021 is the current last migration. Existing SQLite DSNs enable foreign keys and busy timeout per pooled connection.

The owner authorized sequential inline OpenSpec Plus planning, implementation, archive and separate commits without subagents. Reviews remain inline; no new library is needed.

## Goals / Non-Goals

**Goals:** Durable private work with atomic submission, application idempotency, finite claims/calls, cancellation/restart fencing and immutable validated history, exercised through a synthetic executor.

**Non-Goals:** No public job API, startup worker, real CH09 dispatch, selection/consent UI, token/monetary estimation, draft/write operation, automatic retention schedule or live performance claim. CH12 must supply fresh selection/consent/asset admission and enforce production call/token policy.

## Decisions

### Ownership and alternatives

Use `backend/internal/ai/jobs/` for normalized submission, deterministic state/retry policy, small consumer-owned persistence/executor contracts and a cancellable one-item worker. Root `aiJob*.go` adapters own SQLite transactions, private revision binding and canonical result persistence. No core SQL/driver/provider transport import is allowed. Results are the only cross-feature domain dependency. Existing deployment and Goose migration ownership remain unchanged.

A queue broker or external worker service would duplicate coordination and violate the adopted topology. Root-only workflow logic would mix scheduling policy with SQL. Choose a focused core plus the existing root persistence boundary; no speculative scheduler or generic repository framework is needed.

### Durable data and submission

Migration 022 adds owner-qualified `ai_jobs`, `ai_job_items` and `ai_analyses`, indexed for due work and private history. Jobs freeze installation, exact profile revision/model, Visual mode, normalized languages, primary language, selection digest, explicit `visual-v1` consent version, finite call cap and normalized-request digest. Items freeze exact deduplicated asset IDs and stable order; they retain state, claim/call counts, due time, random token/expiry and safe failure code. Analyses retain immutable canonical JSON, outcome, source/prepared digests, prompt/schema/validation versions and exact job/item/provider/language provenance. Image bytes, prompts, keys and raw errors are never stored.

Internal callers provide already-admitted asset IDs; CH11 validates shape, current installation, enabled AI and owner-qualified enabled provider revision. It does not claim to implement CH12's fresh snapshot/consent/access admission. Deduplicate IDs preserving order; normalize languages with CH07. Limit to 500 assets, three claims and three reserved calls per item; job call cap is positive and at most three times item count. One idempotency key identifies the normalized request for one owner/installation. Same request returns the existing job, different request conflicts; a new key is a new run. Key retention lasts as long as its job record.

Use short write transactions that acquire SQLite's write lock before read/decide/update, preventing deferred-read upgrade races. Queries and joins remain owner-qualified. No transaction crosses executor work. Request and result serialization have explicit bounds. Every error exposed by the adapter is a fixed category; internal SQL errors are not returned as private input text.

### Claims, leases and budgets

Defaults are two global active leases, one per owner and a 180-second lease. Callers pass the same validated policy to claims and workers; policy permits 1–16 global leases, 1–global per owner and a 1-second to 10-minute lease. The intended default heartbeat cadence is 30 seconds, supplied through the caller's tick channel; no scheduler or operator configuration is installed. Claim selects one due queued/retry_wait item under current installation authority and capacity, increments its execution count and assigns a random UUID token. Active canceled requests still consume capacity until finalized or expired because cancellation cannot recall a provider call.

Lease identity includes owner/job/item/installation/token. Authorization, heartbeat, reservation and finalization compare every field and current time. `Reserve` atomically increments item and job counts before a synthetic dispatch; it can conservatively charge an admission even if later network work fails. Counts are never refunded. Both claim and call counts are bounded so repeated pre-dispatch failures/restarts terminate. Call/token estimation beyond these durable counters is CH12 work.

Finalization checks exact current unexpired lease authority and cancellation before writing. Stale tokens cannot write any state. Canceled current work becomes canceled through failure finalization or expiry recovery without a result. Success requires at least one reservation on that lease and a nonzero CH07 proposal revalidated against the job's Visual language contract, with bounded provenance. Result insertion and succeeded transition share one transaction. The last job reservation fails remaining undispatched work as budget-exhausted. Immutable-result update protection is enforced in SQLite; no foreign key to catalog assets or ephemeral snapshots can erase history during sync.

### Recovery, worker and failure policy

Recovery processes at most 100 expired running items per call. Canceled jobs become canceled; exhausted claims/calls become failed; others become retry_wait with a new later due time and cleared token. A replacement claim always receives a new token. A fresh process cannot steal an unexpired lease. Job aggregate status is derived from item counts; analysis outcome remains separate.

`RunOne` claims, supplies durable authorize/reserve guards to an injected executor and monitors an injected heartbeat channel. On context cancellation, authority loss or heartbeat failure it cancels the executor context and joins the monitor/executor; it does not abandon work goroutines. Executors must honor context. Shutdown leaves interrupted ownership to bounded expiry/recovery if a context-safe finalization is not possible. No worker loop is registered at startup.

The executor returns only a validated completion or explicit safe failure category. Transient retry uses deterministic policy with injected bounded jitter and retry-after, capped at five minutes. Permanent errors fail one item. Authentication/policy systemic errors block the current item and remaining queued/retry items in that job; resumption is an explicit later run, not an automatic loop. Unknown/ambiguous valid results succeed. Late responses and ambiguous external charges are not exactly-once guarantees.

### Installation and retention

The existing root installation binder invokes a focused jobs invalidation adapter before replacing the identity in its existing transaction. This cancels old unfinished jobs/items and clears their leases atomically, using the binder's injected clock. It neither invents another instance key nor deletes terminal analyses. Current-installation reads reject stale bindings. Follow the existing operator replacement procedure: disable AI, change the endpoint/epoch as appropriate, restart, run full catalog sync for connected users, then enable AI. Identity rotation does not automatically rebuild the catalog; no new reindex endpoint is exposed.

Owner deletion cascades jobs/items/results, verified on pooled connections. An explicit owner-scoped cleanup method takes a nonfuture cutoff and a batch limit of 1–100. It removes jobs only when creation and all succeeded/failed/canceled completion times precede that cutoff, cascading their private results. Active/recent jobs remain; blocked jobs must first be canceled. Explicit cleanup covers retained old installations and works while AI is disabled. No age default or automatic scheduler is enabled; production retention requires owner approval before release. Later drafts/write audits must add restrictive references or retention exclusions before using this cleanup boundary. Local cleanup never changes Immich.

### Verification and rollout

Sequential TDD uses pure policy tests and real migrated file-backed SQLite for submission, competing stores, reservations, leases, reopen/recovery, rollback injection, immutability, isolation, rotation and cleanup. A deterministic executor proves cancellation, heartbeat and failure outcomes without external provider/Immich access. Test fresh and 021-upgrade databases, pool-wide foreign keys, unchanged selection/provider/catalog behavior, all shared gates, Docker packaging and fresh complete GitNexus change review. Synthetic timing is not a reference-NAS benchmark.

Migration rollback is intentionally not an operational recovery plan because down migration deletes AI history. Back up SQLite consistently before rollout; restore a compatible database snapshot if needed. The migration adds inactive internal functionality; CH12 must explicitly wire production admission and execution later.

## Risks / Trade-offs

- SQLite serializes short writes; bounded transactions and concurrency tests protect correctness, but reference-NAS latency remains unmeasured.
- Provider may charge after timeout/cancellation; reservations and fencing protect local state, not exactly-once billing.
- Heartbeat cadence depends on caller scheduling; expired authority fails closed and recovery is bounded.
- Installation invalidation preserves sensitive historical data until explicit owner cleanup/account deletion; automatic retention is deliberately unconfigured.
- A non-cooperative executor can delay shutdown; context compliance is part of its internal contract, not a goroutine timeout workaround.
