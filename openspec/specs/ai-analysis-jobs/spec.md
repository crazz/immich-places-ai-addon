# AI analysis jobs Specification

## Purpose

Keep AI analysis work and validated history durable, owner-scoped and bounded across cancellation, competing workers and backend restarts without granting Immich write authority.

## Requirements

### Requirement: Freeze bounded private submissions atomically

Internal submission SHALL persist one job and its exact deduplicated asset membership atomically, with owner, current installation, exact private provider revision/model, Visual mode, normalized languages, selection digest, explicit consent version and finite call limits. It SHALL accept at most 500 distinct assets, require enabled AI and provider authority and reject malformed or foreign references. Production selection/consent/access admission and public routes belong to the later integration change.

#### Scenario: Valid submission survives reopening
- **WHEN** a valid synthetic submission is stored and the database is reopened
- **THEN** its exact configuration and deduplicated membership remain queued without provider or Immich calls

#### Scenario: Invalid or foreign submission is rejected atomically
- **WHEN** submission has invalid bounds, missing consent binding, disabled AI/provider, a foreign revision or stale installation
- **THEN** no job or partial membership is published

#### Scenario: Insertion failure rolls back membership
- **WHEN** persistence fails after a job header is inserted
- **THEN** neither its header nor any item survives

### Requirement: Deduplicate application submission without overwriting history

An idempotency key SHALL be scoped by owner and installation. Repeating the same normalized request SHALL return the same job; reusing its key with different execution configuration or membership SHALL conflict. Concurrent submissions SHALL not create duplicate jobs. A new key SHALL create a separate run without replacing earlier results. Keys remain reserved for the lifetime of the retained job.

#### Scenario: Concurrent identical submissions converge
- **WHEN** independent connections submit the same key and normalized request concurrently
- **THEN** exactly one job and one set of items exist and all successful callers receive its ID

#### Scenario: Changed request conflicts
- **WHEN** an existing key is reused with a different asset set, language set, provider revision or limit
- **THEN** the original job remains unchanged and the new submission conflicts

#### Scenario: Explicit new run preserves earlier results
- **WHEN** the same asset is submitted under a new key after completion
- **THEN** the new run has new work while the original immutable result remains intact

### Requirement: Claim and reserve bounded work under exclusive leases

A claim SHALL atomically select one due item in the current installation, respect finite global/per-owner concurrency, increment its bounded execution-attempt count and assign an unpredictable lease token and expiry. Authorization, heartbeat, reservation and completion SHALL require the current unexpired token. Each provider reservation SHALL durably increment item and job call counts before external work, never exceeding their caps. Execution claims SHALL also be bounded so repeated pre-dispatch failures cannot retry forever. No transaction SHALL span executor work.

#### Scenario: Competing claims respect capacity and ownership
- **WHEN** independent workers compete for due work from multiple owners
- **THEN** each item has one lease and current global/per-owner limits are respected

#### Scenario: Future retry is not claimable
- **WHEN** a retry is scheduled in the future
- **THEN** it is not claimed before its due time

#### Scenario: Reservation cannot exceed item or job budget
- **WHEN** concurrent reservations reach a configured cap
- **THEN** accepted calls are counted durably and every excess reservation is denied

#### Scenario: Heartbeat requires the exact active lease
- **WHEN** a current lease is renewed or a foreign, expired or replaced lease is presented
- **THEN** only the current unexpired lease extends ownership

### Requirement: Recover and cancel without late publication

Expired leases SHALL be reconciled in bounded batches: canceled work becomes canceled, exhausted work becomes failed and otherwise interrupted work becomes retry_wait. Recovered work SHALL use a new token. Cancellation SHALL stop queued/retry/blocked work immediately and request cancellation of running work, preserving completed results. Cancellation, shutdown and lost authority SHALL prevent subsequent reservations; stale workers SHALL never finalize another lease. The executor SHALL honor context cancellation and heartbeat failures without abandoned work goroutines.

#### Scenario: Restart recovers expired work and fences old worker
- **WHEN** a running lease expires across reopen and recovery runs
- **THEN** eligible work becomes retryable and an old token cannot reserve, renew or publish over its replacement

#### Scenario: Repeated interruptions reach a terminal bound
- **WHEN** work repeatedly loses leases or fails before dispatch
- **THEN** the execution-attempt cap eventually fails it without unbounded retries or invented call counts

#### Scenario: Cancellation rejects late completion
- **WHEN** a job with queued, running and completed items is canceled
- **THEN** no new dispatch is admitted, a late response cannot create a result, and completed history is preserved

#### Scenario: Worker observes cancellation and heartbeat failure
- **WHEN** the execution context is canceled or a heartbeat cannot retain authority
- **THEN** the synthetic executor receives cancellation and the worker waits for it to return without publishing success

### Requirement: Persist only immutable validated terminal results

Success SHALL atomically store a canonical validated proposal, exact job/item provenance, bounded image/source and prompt/schema metadata, and mark the matching item succeeded. Located, ambiguous and unknown SHALL remain distinct valid outcomes. Failed attempts SHALL retain only allowlisted safe failure categories. Results SHALL be immutable, owner-scoped, independent of synchronized asset rows and have no draft or write authority. Invalid or mismatched proposals SHALL never become success.

#### Scenario: Valid unknown completion is durable
- **WHEN** an active reserved attempt returns a valid unknown proposal
- **THEN** one immutable analysis and a succeeded item survive reopen with their exact provenance

#### Scenario: Invalid result or failed insertion cannot partially succeed
- **WHEN** completion has an uninitialized proposal, wrong requested languages, invalid metadata or a result insertion failure
- **THEN** no analysis or succeeded item is published

#### Scenario: Terminal result cannot be overwritten
- **WHEN** a completed lease is finalized again or a stored analysis is updated directly
- **THEN** its original validated content and terminal state remain unchanged

#### Scenario: History is private and survives source removal
- **WHEN** a synchronized asset disappears or a different owner requests the job/result
- **THEN** local history survives source removal and the foreign read is denied without exposing content

### Requirement: Isolate failures and retry only within explicit policy

The worker SHALL use explicit safe success, transient, permanent and systemic-blocked outcomes. Transient failures SHALL use bounded retry delay with injected jitter and bounded retry-after input; permanent failures SHALL terminate the item. A systemic block SHALL stop new work in the affected job without consuming its remaining items. Other jobs SHALL remain usable. Provider dispatch is at least once under bounded reservations; no exactly-once billing guarantee or hidden retry is allowed.

#### Scenario: Transient failure retries with bounded delay
- **WHEN** the synthetic executor reports a retryable failure and budget remains
- **THEN** the item enters retry_wait with a bounded due time and its consumed reservation is retained

#### Scenario: Permanent failure leaves unrelated items usable
- **WHEN** one item fails permanently
- **THEN** it becomes failed and other eligible work can still complete

#### Scenario: Systemic failure blocks remaining job work
- **WHEN** execution reports an authentication or policy block
- **THEN** undispatched items in that job become blocked and no further calls are spent there

### Requirement: Reuse installation identity and provide controlled lifecycle cleanup

Installation rotation through the existing endpoint/epoch binding SHALL invalidate all unfinished work from the previous identity and fence its leases atomically while retaining terminal history. Account deletion SHALL cascade private job/item/result data. Explicit cleanup SHALL delete only a bounded set of terminal jobs older than a caller-supplied cutoff, scoped to one owner, and SHALL not contact Immich. No automatic retention period, cleanup scheduler, public endpoint or production executor SHALL be enabled by this change.

#### Scenario: Installation rotation invalidates unfinished work
- **WHEN** the existing controlled installation binding rotates
- **THEN** old queued/running work is canceled, old tokens cannot publish and completed analyses remain retained

#### Scenario: Account deletion removes private lifecycle data
- **WHEN** an account is deleted through the existing database boundary
- **THEN** its jobs, items and analyses are removed without affecting another account

#### Scenario: Explicit cleanup is bounded and terminal only
- **WHEN** an owner requests cleanup with a valid cutoff and batch limit
- **THEN** only that owner's eligible terminal history is removed, active/recent work remains and no external mutation occurs

#### Scenario: Existing database upgrades safely
- **WHEN** a database at the previous migration is upgraded and reopened
- **THEN** its catalog, provider and selection data remain intact and new lifecycle constraints are enforced on pooled connections
