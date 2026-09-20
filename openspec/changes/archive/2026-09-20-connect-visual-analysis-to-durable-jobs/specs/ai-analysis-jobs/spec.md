## ADDED Requirements

### Requirement: Admit exact consented production selections

Authenticated production submission SHALL bind a current eligible frozen selection of 1–500 assets, owner, installation, active exact provider revision/model, explicit image consent, Visual mode, output format, normalized languages and finite limits. It SHALL atomically retain the exact selection/filter and consent provenance without inference or upstream image work during admission. Stale selection facts SHALL reject the entire new submission. No submitted identity or replacement membership SHALL override server authority.

#### Scenario: Admit without waiting for inference
- **GIVEN** a current eligible snapshot and all required authority, consent and limits
- **WHEN** the owner submits a Visual job
- **THEN** the server returns its durable job ID after atomic admission without waiting for inference or fetching images

#### Scenario: Reject stale or incomplete authority
- **GIVEN** changed selection facts, a foreign reference, expired snapshot, revoked provider, unsupported format, absent consent or invalid limits
- **WHEN** a new submission is attempted
- **THEN** no job, membership or provider call is created and a safe actionable failure is returned

#### Scenario: Retry an admitted key after snapshot expiry
- **GIVEN** an admitted request whose snapshot has expired and whose job is retained
- **WHEN** its owner repeats the same key and normalized input while mutation admission is enabled
- **THEN** the existing job is returned without new work, while changed input conflicts and a new key requires a current snapshot

### Requirement: Protect bounded job progress and cancellation

Submission and cancellation SHALL enforce authenticated ownership and mutation-origin protection with bounded bodies. Private progress/detail and paginated job listing SHALL expose safe counts, item states, result IDs and explicit usage status without secrets, payloads or cross-owner existence disclosure. Reads and cancellation SHALL remain available with AI execution disabled and SHALL never dispatch work.

#### Scenario: Read and cancel an owned job
- **GIVEN** an owned job with mixed queued, running and completed items
- **WHEN** its owner reloads progress or cancels it, including while AI is disabled
- **THEN** bounded durable progress remains readable and cancellation preserves completed history without new dispatch

#### Scenario: Reject foreign or unprotected access
- **GIVEN** an unauthenticated request, foreign job ID, invalid origin, oversized body or invalid pagination
- **WHEN** the corresponding protected operation is attempted
- **THEN** it fails safely without disclosing private data or mutating work

#### Scenario: Reload stable private job pages
- **GIVEN** more retained jobs than one page and concurrent new submissions
- **WHEN** the owner follows a valid continuation cursor
- **THEN** pages retain deterministic ordering and cursor scope without duplicating or crossing owners

### Requirement: Execute real Visual work under current authority

The production worker SHALL compose authorized image preparation, one Visual attempt and immutable validated completion under the durable lease, consent, exact provider and resource policy. It SHALL recheck current source/access and authority at sensitive dispatch/publication boundaries. It SHALL preserve empty Visual evidence and never invoke an Immich mutation, hidden provider retry, repair or format fallback.

#### Scenario: Complete a real composed attempt
- **GIVEN** an admitted Visual item and currently authorized image/provider access
- **WHEN** the worker receives a valid located, ambiguous or unknown proposal
- **THEN** the matching result and item completion persist atomically with exact provenance and no Immich write

#### Scenario: Lose authority during an attempt
- **GIVEN** queued or running work whose source, consent, provider revision, installation or execution policy changes
- **WHEN** the next dispatch or publication check runs
- **THEN** no unauthorized call or stale success is published and the safe item/job outcome reflects the lost authority

#### Scenario: Keep failure scope and retry bounds explicit
- **GIVEN** transient transport/rate-limit failures, invalid output, inaccessible assets or a systemic provider block
- **WHEN** execution classifies each failure
- **THEN** only eligible transient work retries within existing limits, permanent item failures remain isolated, and systemic blocks stop the affected job

### Requirement: Reserve finite usage before every production call

Every provider call SHALL atomically consume a durable call reservation and conservative token allowance under the current lease before dispatch. Concurrent attempts and retries SHALL not exceed admitted item/job limits. Reservations SHALL remain consumed when actual usage is smaller, unavailable or delivery is uncertain. Reported usage and monetary estimates SHALL remain separate from enforced allowances; unknown values SHALL not become zero or exact billing claims.

#### Scenario: Concurrent calls reach the allowance
- **GIVEN** multiple workers whose next reservations would exceed a call or token limit
- **WHEN** they attempt dispatch
- **THEN** only reservations within both limits succeed and all others make no provider call

#### Scenario: Lose the response after reservation
- **GIVEN** a reserved attempt with timeout, crash, absent usage or smaller reported usage
- **WHEN** recovery or a retry evaluates remaining budget
- **THEN** the full prior allowance remains consumed and new work requires another permitted reservation

#### Scenario: Report cost or policy uncertainty honestly
- **GIVEN** missing tariffs, incomplete usage or usage exceeding the attested allowance
- **WHEN** accounting is recorded and exposed
- **THEN** unknown cost/usage remains explicit, estimates remain labeled, and a violated policy blocks further dispatch without claiming charges were reversed

### Requirement: Start and stop a bounded recoverable production consumer

Worker startup SHALL require enabled AI and valid finite execution settings. Fixed bounded concurrency, deadlines, heartbeats, idle polling and joined shutdown SHALL preserve existing recovery and lease fencing. Browser closure SHALL not cancel accepted jobs. Upgrade SHALL preserve old history without promoting internal fixture jobs lacking production consent into executable work.

#### Scenario: Execute after selection snapshot cleanup
- **GIVEN** an admitted job whose preview token expired and snapshot was cleaned up
- **WHEN** an eligible item executes under its retained membership, consent and current authority
- **THEN** snapshot cleanup does not invalidate the accepted job or widen its targets

#### Scenario: Continue after browser closure and backend restart
- **GIVEN** an admitted job and an interrupted attempt
- **WHEN** the browser closes or the backend restarts
- **THEN** eligible work resumes under bounded recovery with new lease authority and no duplicate success

#### Scenario: Disable or shut down execution
- **GIVEN** queued and running work
- **WHEN** AI is disabled or application shutdown begins
- **THEN** new dispatch stops, running work receives cancellation, workers are joined and late results are fenced

#### Scenario: Upgrade without inventing consent
- **GIVEN** an existing database with internal jobs and immutable results
- **WHEN** the integration schema and worker are installed
- **THEN** history remains intact and only jobs with valid production admission metadata can execute

## MODIFIED Requirements

### Requirement: Freeze bounded private submissions atomically

Submission SHALL persist one job and its exact deduplicated asset membership atomically, with owner, current installation, exact private provider revision/model, Visual mode, normalized languages, selection digest, explicit consent version and finite call limits. It SHALL accept at most 500 distinct assets, require enabled AI and provider authority and reject malformed or foreign references. Public submission SHALL additionally enforce the production admission contract and retain its versioned provenance; internal fixtures cannot grant production consent.

#### Scenario: Valid submission survives reopening
- **GIVEN** a valid bounded private submission
- **WHEN** a valid synthetic submission is stored and the database is reopened
- **THEN** its exact configuration and deduplicated membership remain queued without provider or Immich calls

#### Scenario: Invalid or foreign submission is rejected atomically
- **GIVEN** a new submission that lacks required authority or bounds
- **WHEN** submission has invalid bounds, missing consent binding, disabled AI/provider, a foreign revision or stale installation
- **THEN** no job or partial membership is published

#### Scenario: Insertion failure rolls back membership
- **GIVEN** an admission transaction that has not committed
- **WHEN** persistence fails after a job header is inserted
- **THEN** neither its header nor any item survives

### Requirement: Recover and cancel without late publication

Expired leases SHALL be reconciled in bounded batches: canceled work becomes canceled, exhausted work becomes failed and otherwise interrupted work becomes retry_wait. Recovered work SHALL use a new token. Cancellation SHALL stop queued/retry/blocked work immediately and request cancellation of running work, preserving completed results. Cancellation, shutdown and lost authority SHALL prevent subsequent reservations; stale workers SHALL never finalize another lease. The executor SHALL honor context cancellation and heartbeat failures without abandoned work goroutines.

#### Scenario: Restart recovers expired work and fences old worker
- **GIVEN** a running item with a retained lease
- **WHEN** a running lease expires across reopen and recovery runs
- **THEN** eligible work becomes retryable and an old token cannot reserve, renew or publish over its replacement

#### Scenario: Repeated interruptions reach a terminal bound
- **GIVEN** an item with finite execution-attempt allowance
- **WHEN** work repeatedly loses leases or fails before dispatch
- **THEN** the execution-attempt cap eventually fails it without unbounded retries or invented call counts

#### Scenario: Cancellation rejects late completion
- **GIVEN** an owned mixed-state job
- **WHEN** a job with queued, running and completed items is canceled
- **THEN** no new dispatch is admitted, a late response cannot create a result, and completed history is preserved

#### Scenario: Worker observes cancellation and heartbeat failure
- **GIVEN** a worker executing under a live lease and monitored context
- **WHEN** the execution context is canceled or a heartbeat cannot retain authority
- **THEN** the executor receives cancellation and the worker waits for it to return without publishing success

### Requirement: Reuse installation identity and provide controlled lifecycle cleanup

Installation rotation through the existing endpoint/epoch binding SHALL invalidate all unfinished work from the previous identity and fence its leases atomically while retaining terminal history. Account deletion SHALL cascade private job/item/result data. Explicit cleanup SHALL delete only a bounded set of terminal jobs older than a caller-supplied cutoff, scoped to one owner, and SHALL not contact Immich. Automatic retention and public cleanup SHALL remain unavailable until a separately approved lifecycle policy is implemented. Job submission, progress, cancellation and authorized execution SHALL NOT imply permission to enable cleanup or Immich writes.

#### Scenario: Installation rotation invalidates unfinished work
- **GIVEN** unfinished work and completed history bound to the prior installation
- **WHEN** the existing controlled installation binding rotates
- **THEN** old queued/running work is canceled, old tokens cannot publish and completed analyses remain retained

#### Scenario: Account deletion removes private lifecycle data
- **GIVEN** multiple accounts with private job and result records
- **WHEN** an account is deleted through the existing database boundary
- **THEN** its jobs, items and analyses are removed without affecting another account

#### Scenario: Explicit cleanup is bounded and terminal only
- **GIVEN** retained active, recent and old terminal jobs for multiple owners
- **WHEN** an owner requests cleanup with a valid cutoff and batch limit
- **THEN** only that owner's eligible terminal history is removed, active/recent work remains and no external mutation occurs

#### Scenario: Existing database upgrades safely
- **GIVEN** a populated database using the preceding schema
- **WHEN** a database at the previous migration is upgraded and reopened
- **THEN** its catalog, provider and selection data remain intact and new lifecycle constraints are enforced on pooled connections
