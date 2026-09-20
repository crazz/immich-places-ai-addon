# CH09 Visual analysis verification

Implemented on top of CH08 commit `85078722824a2a7df7225b3c0de823681b604acb`. Work follows sequential inline OpenSpec Plus planning/apply/TDD and self-review without subagents. The [caller guide](../ai-visual-analysis.md) defines the internal handoff and production-admission limits. No public route, startup worker, migration or private-image live call is introduced.

## Scenario traceability

All five delta requirements and 19 scenarios map to automated tests. Workflow tests use real codecs and the canonical validator; root tests use real file-backed SQLite and synthetic HTTP services.

| Scenario | Evidence |
|---|---|
| Bound authorized input can be analyzed | `TestAIVisualAnalysisUsesExactAuthorizedImageAndProvider`: exact source, personal/provider credentials, model, one dispatch, valid unknown |
| Invalid or foreign binding prevents dispatch | `TestVisualRejectsInvalidBindingAndMissingAuthorityBeforeDispatch`: owner/installation/asset mismatch, released/nil image, guards, profile/revision/format/languages; root denial fixture |
| Inapplicable capability or disabled authority prevents dispatch | `TestAIVisualDenialsNeverDispatchPrivateImage`: enabled/access/profile/model/protocol/policy/lifecycle/image/format observations, zero provider calls |
| Historical revision remains exact | `TestAIVisualExplicitJSONAndHistoricalRevisionRemainExact`: a newer active revision does not redirect old work; disablement applies across revisions |
| Source or authority changes before transmission or publication | `TestAIVisualRejectsAuthorityAndSourceChangesAtBothBoundaries`: profile, policy, hidden asset, source, credential, installation and caller changes after resolution or response |
| Visual payload contains only permitted inputs | `TestVisualEncodingKeepsOneImageAndCanonicalFormat` and real HTTP spy: one JPEG, controlled schema/prompt, no source IDs/URL/key/context/tools or optional parameters |
| Strict request retains the canonical contract | Encoder fixture plus `TestCanonicalSchemaHandoffOwnsItsBytes`: same embedded canonical schema, no altered capability codec |
| JSON format requires explicit permission and support | Root JSON/historical fixture, invalid-request cases and read-only capability projection |
| Language normalization cannot widen disclosure or coverage | `TestLanguageHandoffUsesCanonicalContextRules`: non-English normalized exact set, invalid/duplicate tags and absent primary rejected |
| Exactly one reserved dispatch occurs | `TestVisualReservationIsSingleUseAndFailureIsSafe`, root counters and existing CH02 single-transmission transport |
| Missing or exhausted budget prevents transmission | Invalid request/reservation fixtures and `TestVisualCannotPublishWithoutSuccessfulReservation` |
| Unsupported format is reported without automatic downgrade | `TestVisualFormatFailureClassificationIsExplicit`: only 400/422 allowlisted format signal; safe workflow failure fixture; no follow-up call |
| Bounds and cancellation reject partial success | Encoding budgets, bounded parser/transport, `TestVisualDeadlineAndCancellationPreventLatePublication`, `TestAIVisualCancellationDoesNotPublishOrLeakAuthorityFailure` |
| Located ambiguous and unknown outcomes remain distinct | `TestVisualAttemptReturnsValidatedUnknownWithoutPrivateInput`, `TestVisualAlwaysValidatesOutcomeGeometryProvenanceAndLanguages`: real canonical validation |
| Incomplete or ambiguous response framing is rejected | `TestVisualResponseRequiresOneOrdinaryCompleteAssistant`: finish state, role, refusal, tool/function data, choice count, duplicate/case ambiguity and trailing JSON |
| Strict transport cannot bypass semantic validation | Outcome/geometry/provenance/language integration fixture rejects invalid coordinates, invented sources, missing language and unsupported direction |
| Caller mutation cannot alter accepted evidence | `TestVisualResultOwnsProposalAndMetadata`: parser input, language slices, nested proposal and usage copies remain independent |
| Usage is bounded optional observation | `TestVisualUsageRemainsOptionalBoundedAndIndependent`: valid counts retained, absent/malformed/negative/excessive/duplicate counts remain unknown |
| Failures and side effects remain isolated | `TestAIVisualConcurrentAttemptsNeverWriteDatabaseOrImmich`: eight concurrent attempts per outcome under query_only and same-connection total_changes; safe protocol/default-format fixtures and shared legacy/capability gates |

Additional regression: `TestAIVisualSourceFailureRemainsTransientAndNeverDispatches` preserves a safe transient category for unavailable source reads without spending a provider reservation. Default result formatting is explicitly redacted for value/pointer and Go diagnostic formats.

## Verification results

All thirteen shared gates passed against CH08 base `8507872`: checker tests, source-size ratchet, dependencies, Go formatting, lint, type generation, TypeScript, Go vet, race-instrumented Go tests/coverage, Go build, frontend unit tests, frontend build and browser smoke. Backend AI statement coverage was **89.31% (2288/2562)** against the 80% floor. All 28 frontend tests and seven browser journeys passed. Lint reported the same three inherited warnings and no errors.

The final admission-limit classification correction additionally passed the complete Visual adapter tests, workflow/adapter race tests and vet. The backend Docker image was rebuilt successfully after that correction. No dependency manifest changed. The initial dependency gate correctly rejected a direct analysis-to-provider import; the final implementation owns dispatch contracts in the core and translates provider types/errors in adapters, with the unchanged dependency gate passing.

Fresh staged GitNexus analysis returned all **166/166 changed symbols across 46 files and 25 affected flows**, with no partial or truncated result. Its aggregate risk was **CRITICAL**. Reviewed flows cover runner/validator construction, dispatch version loading and DNS pinning, canonical language/context validation, provider policy/credential checks and image eligibility/access. Source confirms the constructor has only synthetic test callers and no startup/public-route registration. Current race, authorization, framing and zero-write tests cover these integration boundaries. The cycle check found no circular imports. Index inference still has documented dynamic/receiver and flow-enumeration gaps; this is not a claim of an exhaustive runtime graph.

## Boundaries and remaining evidence

Tests use synthetic images, local providers and real temporary SQLite with production migrations. The capability projection is read-only even for expired running reports. No original Immich mutation, provider writeback, durable job or result persistence is connected. Production dispatch still requires CH11/CH12 durable admission, consent and reservation callbacks.

Local tests establish request framing and validation behavior, not full-schema support by a real model, geographic truth, direction accuracy, reference-NAS performance or exactly-once billing. No live model quality or full-schema compatibility run was performed. The [evaluation guidance](../ai-visual-analysis.md#evaluation-and-release-evidence) defines what to measure separately.
