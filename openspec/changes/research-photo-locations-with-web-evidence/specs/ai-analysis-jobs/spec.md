## MODIFIED Requirements

### Requirement: Freeze bounded private submissions atomically

Submission SHALL persist one job and its exact deduplicated asset membership atomically, with owner, current installation, exact private provider revision/model, Visual, Context-assisted or Research mode, normalized languages, selection digest, explicit consent version and finite call limits. It SHALL accept at most 500 distinct assets, require enabled AI and provider authority and reject malformed or foreign references. Public submission SHALL additionally enforce the production admission contract and retain its versioned mode-specific provenance; internal fixtures cannot grant production consent.

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

#### Scenario: JB01 Freeze Research inputs and deduplicate submission
- **GIVEN** a valid selection, hint, optional displayed context and provider configuration
- **WHEN** Research is submitted and storage reopens
- **THEN** those exact inputs remain frozen and a repeated identical key resolves to the same job
- **AND** reuse of that key with a changed hint or mode conflicts

### Requirement: Confirm explicit mode and data disclosure

Launch SHALL identify the provider, selection, mode and languages, with working settings and optional advanced format/limit controls. Clicking Start analysis SHALL authorize the displayed images and exact current configuration without a second consent checkbox. Context-assisted classes SHALL remain individually selected. Research SHALL show an editable hint, optional selected-album label and optional capture time in the normal launch form. Research SHALL be the initial mode for new launches unless the user explicitly selects another mode; reruns SHALL retain the explicitly displayed mode. Missing search metadata or a research-specific test SHALL NOT disable launch. Visual SHALL reject context payloads; Context-assisted SHALL enforce CH10 bounds, including explicit empty context. Research SHALL reuse those bounds for its displayed input classes without adding neighboring metadata or photographs. Merely previewing or changing fields SHALL NOT submit a job. Missing compatibility or an unsatisfied explicit restriction SHALL disable launch with an actionable explanation, without automatic provider tests. Ordinary launch SHALL NOT require a manually authored policy.

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

#### Scenario: JB02 Start Research with ordinary inputs
- **GIVEN** a provider with current image/output support and a valid selection
- **WHEN** the user enters a hint and clicks Start analysis in Research mode
- **THEN** the exact displayed inputs are submitted once without another checkbox, search-metadata test or manually authored policy

#### Scenario: JB03 Keep existing mode choices usable
- **GIVEN** the user explicitly chooses Visual or Context-assisted instead of Research
- **WHEN** the launch or rerun is submitted
- **THEN** the selected mode and its existing context rules are preserved without silent substitution

## ADDED Requirements

### Requirement: Run Research with bounded execution and current authority

Research SHALL reuse the existing private admission, exclusive leases, reservations, idempotency, cancellation and recovery controls. Its provider-work deadline SHALL be finite and configurable independently of Visual so a supported deployment can allow longer investigations. Renewing a lease SHALL not extend that deadline. Every actual provider call SHALL consume a durable reservation; no search-action accounting, tool-event persistence, automatic verification call or source-metadata API SHALL be required. Cancellation, lost authority and failed renewal SHALL prevent later dispatch or publication. Ordinary defaults SHALL allow one provider dispatch per item, so an interrupted dispatched item cannot automatically retransmit its photograph under those defaults. An explicit rerun SHALL create separate history with freshly checked inputs.

#### Scenario: JB04 Support a longer investigation without losing ownership
- **GIVEN** a deployment supports the configured finite Research deadline and an item holds a renewable lease
- **WHEN** Research continues past the normal Visual timeout within that deadline
- **THEN** the same attempt remains observable across UI closure/reload and its active lease is renewed
- **AND** Visual retains its existing timeout and lease renewal does not extend the Research deadline

#### Scenario: JB05 Fence canceled or stale completion
- **GIVEN** an active Research attempt
- **WHEN** it is canceled, loses authority, exceeds its deadline or cannot renew its lease
- **THEN** no further dispatch or successful late publication occurs, while consumed usage remains recorded

#### Scenario: JB06 Recover without replay under ordinary defaults
- **GIVEN** the single allowed Research dispatch occurred before the process stopped without committing a result
- **WHEN** storage reopens and recovery reconciles the item
- **THEN** its consumed reservation prevents automatic retransmission and an explicit rerun creates a different job

### Requirement: Retain estimates and answer references as immutable private history

Research completion SHALL atomically persist the validated result, including proposed coordinates, nullable estimated error, alternatives, explanation and bounded answer-provided references, with its normal owner/item/input/provider/prompt/schema provenance. Coarse error, missing links and missing search metadata SHALL NOT prevent successful completion. Results SHALL remain immutable and readable after restart, disabled execution or loss of the source photograph. Foreign reads SHALL disclose nothing; account/history deletion SHALL remove the corresponding answer references, and a late worker SHALL not recreate them. Storage or authority failure SHALL not publish partial success.

#### Scenario: JB07 Reopen a coarse result with answer references
- **GIVEN** an accepted estimate with a radius of 500 meters or more and answer-provided references
- **WHEN** storage reopens or execution/source-photo access becomes unavailable
- **THEN** the same private coordinates, radius and safe reference presentation remain available without research or an Immich write

#### Scenario: JB08 Preserve atomicity and ownership
- **GIVEN** completion encounters a storage failure, stale authority or another owner's result identity
- **WHEN** it attempts to persist or read the answer
- **THEN** no partial success or foreign evidence disclosure occurs

#### Scenario: JB09 Clean up answer references with history
- **GIVEN** retained Research history and an in-flight old worker
- **WHEN** the owning account or applicable history is deleted
- **THEN** its retained answer references are removed and late completion cannot recreate them
