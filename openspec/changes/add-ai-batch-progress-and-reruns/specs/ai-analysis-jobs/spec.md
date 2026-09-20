## ADDED Requirements

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

Launch SHALL show exact provider revision/readiness, mode, format, languages, limits and disclosure before submission. Image consent and each Context-assisted class SHALL require explicit choice. Visual SHALL reject context payloads; Context-assisted SHALL enforce CH10 bounds, including explicit empty context. Changes to confirmed inputs SHALL invalidate confirmation. Missing compatibility, policy or consent SHALL disable launch without automatic provider tests.

#### Scenario: Launch Visual without hidden context
- **GIVEN** a ready provider, valid preview and chosen languages/limits
- **WHEN** the user confirms image disclosure in Visual mode
- **THEN** the exact batch is submitted once with no context fields and no implicit fallback or extra provider test

#### Scenario: Choose bounded Context-assisted inputs
- **GIVEN** Context-assisted mode and independently selected context classes
- **WHEN** the user confirms disclosure
- **THEN** only those classes and their bounded values are admitted, with image consent still required and absent usable context made explicit

#### Scenario: Reset confirmation after changes
- **GIVEN** a confirmed launch form
- **WHEN** provider revision, selection, mode, hint/classes, languages, format or limits changes
- **THEN** submission requires renewed confirmation for the exact new input

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

## MODIFIED Requirements

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
