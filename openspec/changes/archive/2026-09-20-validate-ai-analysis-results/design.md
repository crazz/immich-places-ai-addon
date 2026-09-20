## Context

See [proposal.md](proposal.md) for intent and the [capability delta](specs/ai-location-proposals/spec.md) for behavior. At planning base `86b44ac2a33a9ece62afc7551581ca8dacc02287`, the canonical v1 schema and two synthetic examples exist as documents; no complete analysis-result validator or result persistence exists. CH07 has no preceding product dependency and can be implemented independently of selection. CH09 will supply provider output and CH11 will own durable result envelopes later.

GitNexus query/context and source verification identify `providerhttp.ParseAssistantText` as the current synthetic capability response boundary, called by the capability protocol and its tests. It rejects refusal/truncation/tool responses but is bounded and coupled to the capability probe contract; it does not validate analysis results. Leave that path intact. The backend Dockerfile already copies `internal/`, so a schema embedded inside a focused AI package can reach the binary without runtime filesystem access. The installed dependency checker has general AI-core rules and an analysis-to-writer rule; extend its protected package coverage to the actual new result package.

Follow the adopted [architecture](../../../../docs/engineering/architecture.md), [testing](../../../../docs/engineering/testing.md), [coding standards](../../../../docs/engineering/coding-standards.md) and [ADR-07](../../../../docs/engineering/decisions/ADR-07-ai-internal-packages.md). The owner approved the v6.0.3 validator recommendation recorded in [library-evaluation.md](library-evaluation.md). No architectural exception or runtime upgrade is needed.

## Goals / Non-Goals

**Goals:** deterministic validation with one canonical structural contract, explicit semantic policy, bounded diagnostics and a narrow result type that later consumers can use without treating model data as authority.

**Non-Goals:** schema generation from model output, JSON repair, provider response-envelope parsing, factual location verification, a public validation endpoint, database records, worker lifecycle, image processing, review UI or write-plan construction. Calling the validator alone cannot prove session validity or current upstream asset access.

## Decisions

### One focused package and a narrow internal input

Implement `backend/internal/ai/results/` with files split by parsing/schema, typed result representation, outcome/references, geometry/provenance and languages. Keep source/tests within the 500-line standard; do not add empty future analysis/jobs packages or a general shared-model hierarchy. The package imports the standard library, the approved JSON Schema library and the existing x/text language support. It has no handler, SQL, provider transport, Immich client, writer, environment lookup or runtime file dependency.

The validation entry point receives model bytes plus a separate server-created context containing mode, requested languages, primary language, completion state and bounded authorized evidence descriptors. Completion states distinguish `complete`, `refused`, `truncated` and `tool_response`; zero/unknown state cannot default to complete. Missing/invalid mode, language context or duplicate/invalid evidence descriptors produce an invalid-context error. Non-complete transport state produces an incomplete-response error before parsing. CH09 is responsible for mapping its actual transport status to this context; CH07 does not import capability-specific transport errors or widen the existing probe parser.

Evidence descriptors carry only an opaque source ID and server-attested support flags for context extent, source-reported radius and known-viewpoint alignment. They contain no credential, URL to fetch, image bytes or new external lookup. Callers must construct them from the exact authorized input bundle after access and consent checks. CH07 checks membership and internal consistency, not external authorization or truth. Visual input has an empty bundle; Context-assisted input can carry descriptors for later CH10 consumers without implementing evidence discovery here. Output IDs cannot supply or override this context.

### One canonical schema, compiled locally

Use `github.com/santhosh-tekuri/jsonschema/v6` pinned to v6.0.3 for Draft 2020-12 structural validation. The documented comparison and owner decision live in the evaluation record; the selected library fits Go 1.25 and avoids reimplementing generic schema semantics. Keep `encoding/json` for parsing and the already-installed `golang.org/x/text` version for language handling; declare x/text as direct when used without an unrelated upgrade. Commit module/checksum changes only during implementation, using a reviewed dependency update rather than an incidental install rewrite.

Move the single canonical JSON file from `docs/ai-locate/contracts/ai-analysis-result.v1.schema.json` to `backend/internal/ai/results/ai-analysis-result.v1.schema.json` during apply. Preserve its bytes, `$id`, version and closed-object/nullability contract. Embed that asset into the result package and update current repository links to its new location, including these planning artifacts and example guidance. Original Word exports remain historical copies. A short document at the former directory can point to the canonical asset, but there must not be a second handwritten schema or generated approximation serving as another authority.

Construct and compile a validator once from the trusted embedded resource, explicitly using its 2020-12 dialect and a loader that rejects all unregistered lookups. Local `$defs` references resolve from that resource; built-in dialect metadata does not authorize network or arbitrary file access. Never accept a schema location, dialect or loader from a response. Do not expose mutable schema bytes or compiler state. Construction failure returns a sanitized unavailable error; later composition must keep analysis unavailable rather than weaken validation or disable unrelated manual workflows. Concurrent validations share only the immutable compiled schema and own per-call state; verify that usage with race tests.

### Bounded parse and validation pipeline

Apply the following stages in order: validate/copy server context and completion state; enforce raw byte/Unicode/structural budgets; decode one JSON value with exact numeric tokens; validate against the compiled schema; decode the typed wire result; enforce semantic invariants; publish an independent validated value. Each failed stage returns no validated result. Accept surrounding JSON whitespace but no markdown fences, extracted substring, trailing value, unknown property or duplicated object key. Use a narrow lexical guard around the standard decoder for duplicate keys, depth and Unicode integrity; it is not a custom general JSON parser. Reject invalid UTF-8 and unpaired escaped surrogates instead of relying on replacement-character repair.

| Budget | Initial fixed ceiling |
|---|---:|
| Model result bytes | 1 MiB |
| Nesting depth | 32 containers |
| Total decoded value/container nodes | 20,000 |
| One decoded string | 8 KiB UTF-8 |
| Identifier or language tag | 128 bytes |
| Observations / candidates | 100 / 20 |
| Requested languages / description entries | 10 / 10 |
| Authorized source descriptors | 100 |
| Evidence refs / source refs per candidate | 100 / 100 |
| Uncertainty notes per candidate / top-level warnings | 20 / 50 |
| Returned findings | 20 findings, at most 4 KiB total |

These are application safety ceilings, not provider token limits or performance claims. Context and result arrays have their own bounds before expensive semantic work. Length/count tests cover each limit exactly and one over. Before schema validation, bound lexical number tokens to 128 bytes and explicit exponent magnitude to 324, reject float overflow/non-finite conversion and nonzero values that underflow to zero, then retain accepted tokens as `json.Number` for exact structural range comparisons. This prevents a short extreme exponent from causing unbounded arbitrary-precision work and prevents rounding an out-of-range value into an accepted boundary. No coercion from strings, booleans or omitted fields is allowed.

The static schema, bounded tree and collection caps bound per-call work without a validation goroutine or hidden retries. Domain loops build request-local ID sets and traverse bounded collections. A future worker owns its surrounding deadline/cancellation and attempt budget; CH07 performs no queue scheduling. Record local representative and maximum-bound timing/allocations to catch an unacceptable implementation, without inventing a NAS service-level guarantee.

### Semantic policies and evidence limits

Separate outcome, reference, geometry and language checks so each has stable findings and independently testable boundaries. All returned candidates are proposals; a selected candidate is never an approval. Unknown output remains successful with no automatic selection; an ambiguous result retains alternatives in provider order. Do not promote a subject point or use `0,0` as a null sentinel. Structurally valid but semantically inconsistent output fails the complete result; there is no repair/downgrade path that silently produces a different accepted object.

IDs are opaque exact-match identifiers: reject blank or duplicated identifiers/references rather than trim or renumber them. Check required nonblank observation/place/support text and optional ISO alpha-2 country values using the existing language/region data plus explicit country-membership rules; reject unknown, macro-region, reserved and user-assigned codes rather than infer a country from coordinates. Normalization applies only to language tags, not prose or model IDs. Candidate references can point only into the corresponding validated collection or the caller's authorized source descriptor map. A provided-context observation requires Context-assisted mode and a nonempty bundle; a candidate citing it must cite an authorized source. These reference checks cannot prove that untrusted prose faithfully describes that source.

Keep camera and subject coordinate fields distinct through decoding, validation and output copying. Pair/range/finiteness checks include zero and exact geographic boundaries. Radius nullability is bidirectional: null pairs with `unknown`, numeric pairs with a supported basis. `point` can represent zero radius; area/site require a positive radius when numeric; city/region use null and unknown basis in this initial policy. This conservative rule implements the documented misleading-precision concern without inventing minimum-meter accuracy thresholds. A visual radius needs referenced visual observations, while context/source bases need an actually referenced descriptor carrying that support flag. No radius is labeled calibrated or externally verified by this package.

Direction remains candidate-local and nullable. A camera-bearing candidate is a reviewable viewpoint even when unselected; unknown/ambiguous outcome rules still prevent automatic selection. Visual direction requires a referenced visual observation. Known-viewpoint alignment requires a referenced authorized descriptor with alignment support; two coordinate pairs or model-written bearing prose alone cannot supply that flag. Check the canonical true-north, azimuth and uncertainty bounds without modulo-normalizing 360 into zero. Do not derive a heading or parse free text to guess whether a geographic assertion is true. The checks establish the declared evidence connection; optical-axis accuracy remains a later quality/review concern.

### Exact language/status handling

Normalize and validate requested/output BCP 47 tags with the existing x/text language package, rejecting parser errors instead of accepting a best-effort partial tag. Detect duplicates after canonical normalization, including aliases/case variants; do not use a language matcher that silently substitutes a nearby language. Validate the primary tag against the same normalized request set. One to ten requested languages is valid; English is optional. Output order can be retained, but set equality and exactly-one coverage are authoritative.

Model entries retain `complete` or `unavailable`; these differ from a failed attempt. Enforce the existing text/reason and candidate/scene basis rules, with whitespace-only text considered blank and original accepted prose retained. Missing entries are not synthesized as unavailable, and inconsistent unavailable entries are not repaired. CH11 later persists statuses and CH17 handles explicit retranslation; this package neither retries nor calls another model.

### Immutable success and sanitized failure

Return an opaque validated proposal with private owned state and copying accessors/serialization, plus the validator-policy version. Callers cannot create a nonzero valid proposal merely by filling public fields or mutate it through shared slices, maps, pointers, schema buffers or context descriptors. Downstream methods must reject the zero/uninitialized value. Retain normalized language tags and the original accepted field values; never add application IDs, verification labels, approval or write-plan state to the model object. The later server-owned envelope remains a separate type owned by its consumer.

Failures have a stable category such as invalid context, incomplete response, invalid JSON, limit exceeded, unsupported schema, schema violation or semantic violation, with bounded subcodes and safe structural paths. Stop after the first failing stage; within the semantic stage sort findings deterministically by safe path and code before applying the result budget. Paths contain only known schema field names and bounded numeric indices. For unknown properties use the known parent path; never echo an injected property name. Do not serialize library error strings, offending values, raw JSON, excerpts, precise locations, source IDs or credentials. No raw input logging or partial accepted result is produced. Accepted warnings/descriptions remain untrusted text for later safe rendering, not commands or verified facts.

### Verification and downstream handoff

Use Plus TDD for new behavior and the established characterization rule for the two existing canonical fixtures. Keep the original fixture examples with their exact matching language/mode contexts; add small independent fixtures for ambiguity, subject-only results, authorized context, languages and each invalid boundary. Deriving negative inputs from a valid fixture is acceptable when assertions prove rejection of a specific behavior, not the implementation's shape. Map every spec scenario to the outcome task in [tasks.md](tasks.md).

Use pure Go tests for parsing, semantic checks, context isolation, deterministic errors, constructor failure and copying/immutability. Test duplicate keys at several depths, number overflow and precise range violations before float rounding, invalid Unicode, closed nullable fields, every budget and refusal/truncation/tool state. Fuzz the bounded parser/validator for panics, unexpected acceptance and leaked values; retain useful regression seeds. Test concurrent calls under the race detector. A denying loader fixture and import/dependency assertions prove local schema access and no provider/Immich/SQL/writer reachability; a fake later consumer receives either the full typed proposal or failure without a persistence implementation.

Run the 13 installed shared checks with the actual implementation base, AI Go statement coverage at least 80%, relevant frontend coverage if touched, all backend race tests and both production builds. Extend dependency-checker fixtures for `results/` and check updated schema/document links. Verify packaging by compiling the result package's self-contained test binary in the backend container build context, then running its embedded-schema fixture check in a runtime environment without the source/docs tree. The ordinary backend binary need not link an unused package before CH09 introduces a consumer; do not add a public test route or unused startup call merely to force linkage. Run a pinned-dependency vulnerability check and report actual findings/tool availability; no such audit has already passed during planning. There is no CH07 SQL migration or persistence claim requiring a new SQLite suite, and no new UI requiring a new browser journey; existing manual/GPX/provider smoke regression remains part of shared verification. GitNexus impact precedes any existing symbol/shared-contract edit and complete change analysis precedes a commit.

CH09 must pass authoritative completion state and the exact requested languages/evidence bundle, use the canonical backend validator even if its transport schema is simplified, and classify a valid unknown separately from failure. CH11 must wrap only successfully validated data with its own user/installation/job/asset identity and persist immutable results. Both consumers retain current authorization and cannot treat validation as analysis consent or write permission. CH15 and later review still own safe text rendering and honest provenance display.

## Risks / Trade-offs

- Schema-valid output can still contain false geography or misleading prose → claim only machine-checkable consistency and evidence membership; preserve review and separate quality evaluation.
- Third-party schema compilation can load resources by default → embed one trusted resource, install a rejecting loader and test construction/validation with all external access denied.
- Canonical-schema relocation can break documentation or packaging → preserve bytes/version, update references and verify repository fixtures plus a standalone package test binary built in the backend container context before completion.
- Fail-closed semantics reject a complete response for one bad language/reference → preserve explicit unavailable statuses when valid; leave retries and review workflows to their owners rather than silently salvage a different result.
- Conservative city/region radius and alignment-evidence rules reject some plausible model claims → keep uncertainty explicit; any later relaxation needs measured rationale and a capability delta, not a hidden validator bypass.
- The source descriptor context is trusted input → validate its internal shape, require per-request construction at the authorized consumer and test that model output cannot extend it; CH07 alone does not establish live tenant access.

### Migration and rollout

No database or application-state migration is required. During implementation, add the approved pinned dependency, relocate/embed the single schema and reconcile links plus semantic guidance with the capability. Preserve version 1.0 structural bytes; record semantic policy version separately so later stored envelopes can identify the applied rules. Verify module/checksum and Go 1.25/container compatibility without changing the runtime baseline. CH07 introduces no public route or production dispatch; CH09/CH11 integrate the validator in their own changes. Binary rollback before those consumers has no persisted CH07 data to convert. Future consumer rollback must preserve their own versioned result records rather than treating an old validator as authority to rewrite history.
