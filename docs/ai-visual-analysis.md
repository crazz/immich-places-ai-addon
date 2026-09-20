# Internal Visual analysis

CH09 connects [prepared images](ai-image-preparation.md) with the [canonical validator](ai-result-validation.md) through one internal provider attempt. It has no public route or startup worker. CH11 implements [durable jobs and storage](ai-analysis-jobs.md) with synthetic execution; CH12 must connect its lease/reservation guards and add revision-bound consent, current access and token/cost admission before production use. CH10 adds a separate [consented Context-assisted entry point](ai-consented-context.md); Visual rejects supplied context.

## Caller handoff

The root `aiVisualAnalyzer` combines current SQLite/Immich authority with the existing bounded provider dispatcher. Its request names application owner, installation, exact asset, provider profile/revision, requested languages/primary language, explicit `strict` or `json` format and one valid `images.Prepared`. The root loads the model from the exact immutable revision; caller model text cannot redirect it. A newer active revision does not replace an explicitly bound historical revision. Current profile disablement applies to all revisions.

Two callbacks are mandatory: `Guard.Authorize` verifies current caller/consent/lease authority, and `Guard.Reserve` durably reserves a dispatch immediately before transmission. Synthetic callbacks in tests do not implement production admission. The workflow checks authority before encoding, at the transport's after-resolution boundary and before publication. Only one successful reservation is allowed per invocation; even a failed later credential admission may consume that reservation. Reservations and timeouts cannot prove whether the provider received or billed a request.

The adapter rechecks current account/key, installation and local asset eligibility, exact upstream source digest, current profile/policy/capability applicability and prepared-handle validity. Reads finish before network calls. The capability projection is read-only, including for stale running observations. Separate source/authority reads do not provide an atomic upstream snapshot.

The pure `internal/ai/analysis` workflow owns its narrow dispatch types and safe errors; the root adapter translates provider types and failures. It cannot import provider transports, SQL, HTTP clients or writer packages. Construct the validator/workflow at composition and pass only server-authorized requests. The existing capability codecs and their smaller sample budgets are unchanged.

## Payload and result

The versioned `visual-v1` prompt and canonical v1.0 schema request concise visual evidence, uncertainty, distinct camera/subject positions, evidence-backed direction and the exact normalized language set. Visible text is untrusted scene content. English is optional. No arbitrary caller prompt, context bundle, neighboring image, filename, capture metadata, local ID or Immich URL/key enters the provider body. Exactly one JPEG data URL is included; optional token/model parameter guesses, tools and streaming are absent.

Strict mode requires applicable observed strict support. JSON mode additionally requires explicit caller permission and applicable JSON support. A synthetic capability sample does not establish full-schema compatibility. Both modes always run canonical structural and semantic validation; a valid unknown/ambiguous result succeeds without manufacturing a selected camera point.

Only one ordinary assistant choice with explicit `stop` and string content is accepted. Refusal, tool/function data, ambiguous/duplicate relevant envelope fields, invalid text/JSON, truncation and semantic failure produce no proposal. The opaque result exposes an independent proposal and copied metadata through explicit accessors; default JSON/Go formatting does not disclose retained content. Metadata records exact image/source/policy binding, provider revision/model, format, prompt/schema versions, languages and optional bounded usage. Missing or malformed usage remains unknown, not zero or verified cost.

## Limits, failures and ownership

The attempt has a 120-second overall deadline, honoring earlier caller deadlines; request bytes are capped at 15 MiB and response bytes at 1 MiB. CH08 separately limits image transmission representation to 10 MiB. Canonical result limits still apply after envelope parsing. Only bounded nonnegative usage counts up to one billion tokens per reported field are retained as observations; these values are not admission budgets.

There is no automatic retry, repair, downgrade, original-file fallback or alternate asset. An allowlisted unsupported-format signal on HTTP 400/422 is reported distinctly. Authentication/policy denial, rate limits, timeouts and malformed output never implicitly authorize a JSON retry. The durable caller may authorize a later separately budgeted attempt under its own policy.

The caller owns and releases the prepared handle. The workflow clears byte copies it owns and retains only validated output and bounded metadata. Independent caller copies and library/GC allocations are outside that cleanup guarantee. Never log explicit payload accessors. Cancellation or observed authority loss discards late results, but cannot recall transmitted data.

## Evaluation and release evidence

Use synthetic or owner-approved fixtures only. For Visual evaluation, remove ground-truth GPS and filename/album clues from inputs without changing originals; separate trips between evaluation sets. Record located/ambiguous/unknown coverage, independent camera-coordinate error rather than POI-centroid error, subject-only cases, supported/unknown direction and azimuth error, language availability, latency, calls and optional usage. Report the coverage/accuracy tradeoff and dataset composition; no numeric accuracy threshold is implied before an approved baseline.

[Verification](engineering/ai-visual-analysis-verification.md) distinguishes local synthetic behavior from live full-schema provider compatibility, real geographic quality and reference-NAS performance. No private photographs were transmitted and no production deployment was performed by CH09 verification.
