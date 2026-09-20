# AI analysis jobs Specification

## Purpose

Keep AI analysis work and validated history durable, owner-scoped and bounded across cancellation, competing workers and backend restarts without granting Immich write authority.

## Requirements

### Requirement: Freeze bounded private submissions atomically

Submission SHALL persist one job and its exact deduplicated asset membership atomically, with owner, current installation, exact private provider revision/model, Visual or Context-assisted mode, normalized languages, selection digest, explicit consent version and finite call limits. It SHALL accept at most 500 distinct assets, require enabled AI and provider authority and reject malformed or foreign references. Public submission SHALL additionally enforce the production admission contract and retain its versioned mode-specific provenance; internal fixtures cannot grant production consent.

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

### Requirement: Admit exact consented production selections

Authenticated production submission SHALL bind a current eligible frozen selection of 1–500 assets, owner, installation, active exact provider revision/model, explicit image consent and applicable context-class consent, Visual or Context-assisted mode, output format, normalized languages and finite limits. It SHALL atomically retain the exact selection/filter and consent provenance without inference or upstream image work during admission. Stale selection facts SHALL reject the entire new submission. No submitted identity or replacement membership SHALL override server authority.

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

Every provider call SHALL atomically consume a durable call reservation and token scheduling allowance under the current lease before dispatch. Concurrent attempts and retries SHALL not exceed admitted item/job limits. Reservations SHALL remain consumed when actual usage is smaller, unavailable or delivery is uncertain. For application defaults, token allowances SHALL be labeled planning estimates and SHALL NOT claim provider-enforced token ceilings. Reported usage and monetary estimates SHALL remain separate from enforced call and reservation limits; unknown values SHALL not become zero or exact billing claims.

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

#### Scenario: Retain usage above default token estimates
- **GIVEN** a run using application defaults and a provider reporting more tokens than requested
- **WHEN** its otherwise valid result completes
- **THEN** actual usage and the result are retained without fabricating an operator-policy violation, and cost remains unknown

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

### Requirement: Preview the user's exact batch intention

Launch SHALL distinguish selected assets, current page and all matching assets through the existing server snapshot contract. It SHALL show authoritative scope, counts, exclusions and expiry before confirmation, reject empty/oversized selections and never substitute loaded page IDs for all-matching membership. Unsupported filters and changed preview inputs SHALL require correction or a fresh preview.

#### Scenario: Distinguish page from all matching
- **GIVEN** an eligible filter spanning multiple pages
- **WHEN** the user previews the current page or all matching assets
- **THEN** the first preview contains only that page and the second contains the full eligible frozen query membership, with accurate counts

#### Scenario: Explain exclusions and limits
- **GIVEN** inaccessible, unsupported or excluded assets, an empty result or an over-limit selection
- **WHEN** launch previews the batch
- **THEN** it shows safe exclusion/limit information and prevents an invalid submission without silently changing the scope

#### Scenario: Invalidate changed preview inputs
- **GIVEN** a displayed preview
- **WHEN** selection, page, filters or supported scope changes
- **THEN** the old preview and confirmation cannot authorize the changed batch

### Requirement: Confirm explicit mode and data disclosure

Launch SHALL identify the provider, selection, mode and languages, with working settings and optional advanced format/limit controls. Clicking Start analysis SHALL authorize the displayed images and exact current configuration without a second consent checkbox. Context-assisted classes SHALL remain individually selected. Visual SHALL reject context payloads; Context-assisted SHALL enforce CH10 bounds, including explicit empty context. Merely previewing or changing fields SHALL NOT submit a job. Missing compatibility or an unsatisfied explicit restriction SHALL disable launch with an actionable explanation, without automatic provider tests. Ordinary launch SHALL NOT require a manually authored policy.

#### Scenario: Launch Visual without hidden context
- **GIVEN** a ready provider, valid preview and chosen languages/limits
- **WHEN** the user clicks Start analysis in Visual mode
- **THEN** the exact batch is submitted once with no context fields and no implicit fallback or extra provider test

#### Scenario: Choose bounded Context-assisted inputs
- **GIVEN** Context-assisted mode and independently selected context classes
- **WHEN** the user confirms disclosure
- **THEN** only those classes and their bounded values are admitted, with image consent still required and absent usable context made explicit

#### Scenario: Bind the current choices to the Start action
- **GIVEN** a displayed launch form
- **WHEN** the user changes providers or other choices and then clicks Start analysis
- **THEN** provider defaults are refreshed and the submitted authorization binds exactly the new choices, with no dispatch caused by editing

### Requirement: Persist and validate the exact contextual attempt

Context-assisted durable execution SHALL freeze CH10's authorized evidence bundle and provenance before dispatch, apply CH12's current authority and budget controls, and validate completion against that stored mode/source set. Retries SHALL reuse only still-current frozen evidence. Changed evidence SHALL require explicit reanalysis, never silent replacement. Visual history SHALL remain valid and immutable.

#### Scenario: Persist contextual completion across reopen
- **GIVEN** an authorized Context-assisted attempt with its frozen evidence
- **WHEN** valid output completes and storage reopens
- **THEN** the immutable result retains exact mode, sources, consent/policy versions, digest and omissions and can be revalidated against them

#### Scenario: Reject invented or changed evidence
- **GIVEN** output citing unsupplied sources or frozen sources that changed before retry/publication
- **WHEN** the worker validates authority and completion
- **THEN** it publishes no success and never replaces evidence or promotes unsupported precision

#### Scenario: Preserve Visual isolation during upgrade
- **GIVEN** existing Visual results and new Context-assisted support
- **WHEN** migration, reads and either mode's execution occur
- **THEN** old results remain intact, Visual retains empty context, and Context requires its own complete consent and token policy

### Requirement: Follow durable progress across navigation

Progress SHALL use server job state, distinguish execution states from successful proposal outcomes and display safe failures and honest usage status. It SHALL survive reload/browser closure without resubmitting work, ignore superseded responses, clear private state on owner change and label offline/stale data. Polling SHALL be bounded and stop when it cannot provide useful active progress.

#### Scenario: Resume observing a mixed job
- **GIVEN** a job with active, failed and successful unknown/ambiguous items
- **WHEN** the owner reloads or returns after browser closure
- **THEN** its durable progress is restored without a new submission and successful unknown/ambiguous outcomes are not shown as technical failures

#### Scenario: Lose connectivity or switch identity
- **GIVEN** an outstanding progress request
- **WHEN** connectivity fails or the user switches job/account
- **THEN** stale status is explicit and old responses cannot overwrite the current private view

#### Scenario: Reconcile ambiguous submission
- **GIVEN** a confirmed POST whose acknowledgement is lost
- **WHEN** the user explicitly retries reconciliation
- **THEN** the same idempotency key recovers the admitted job without creating another run

### Requirement: Cancel and rerun without replacing history

Cancellation SHALL be explicit and preserve completed results. Retry-failed SHALL select only an owned parent job's terminal failed items; reanalysis SHALL use explicitly chosen parent members. Both SHALL obtain a fresh eligible snapshot, new confirmation and a new run identity with validated lineage. Neither SHALL reset the parent or overwrite its results, reviews or future drafts.

#### Scenario: Cancel mixed work
- **GIVEN** a job with pending and completed items
- **WHEN** the user cancels, including repeated cancellation
- **THEN** future dispatch stops, completed results remain and the UI explains that transmitted data/usage cannot be recalled

#### Scenario: Retry failed members as a new run
- **GIVEN** an owned parent with failed, succeeded, canceled and blocked items
- **WHEN** the user previews and confirms retry-failed
- **THEN** only currently eligible failed members enter a new linked run and the parent remains unchanged

#### Scenario: Reanalyze with new choices
- **GIVEN** selected members of an owned prior run, including unknown or ambiguous results
- **WHEN** the user chooses a new hint/provider and confirms the fresh eligible preview
- **THEN** a new run is created with fresh consent while earlier results remain inspectable

#### Scenario: Reject invalid lineage or changed eligibility
- **GIVEN** a foreign parent, nonmember asset or asset that became ineligible
- **WHEN** rerun admission is attempted
- **THEN** foreign/nonmember lineage is rejected and changed eligible membership requires a fresh confirmed preview without expanding targets

### Requirement: Keep launch and progress accessible without write authority

Launch/progress controls SHALL be keyboard-operable, named and described with text rather than color alone. Progress updates SHALL not repeatedly steal focus. AI disablement SHALL prevent launch while preserving read/cancel access and existing catalog/manual/GPX behavior. No action in this workflow SHALL accept a draft, populate manual pending coordinates or write Immich metadata.

#### Scenario: Complete the workflow using a keyboard
- **GIVEN** a keyboard user
- **WHEN** they preview, configure, confirm, inspect progress and cancel
- **THEN** controls and error/progress announcements are operable and understandable with predictable focus

#### Scenario: Preserve the write boundary and disabled behavior
- **GIVEN** launched or completed AI work, including after execution disablement
- **WHEN** the user inspects progress or uses existing manual/GPX features
- **THEN** AI performs zero Immich mutations and cannot leak proposals into the manual pending-save path
