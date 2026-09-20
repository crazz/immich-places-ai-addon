# AI analysis result validation

CH07 provides the internal Go package `backend/internal/ai/results`. It validates complete provider output before CH09/CH11 may treat it as a proposal. There is no validation HTTP endpoint, automatic provider invocation, persistence, job admission or write consent in this change.

The single [canonical schema](../backend/internal/ai/results/ai-analysis-result.v1.schema.json) is embedded in the package. Its v1.0 bytes and ID are unchanged from the planning contract. Construction compiles Draft 2020-12 with a loader that rejects unregistered resources; production validation needs no source tree, schema file path or network access.

## Caller contract

Call `results.New()` once at composition and keep analysis unavailable if construction fails. The compiled validator supports concurrent independent calls. Do not bypass it because a provider advertises strict-schema support.

`Validate(modelBytes, results.Context{...})` accepts separate server-owned inputs:

| Field | Responsibility |
|---|---|
| `Mode` | Explicit `results.Visual` or `results.ContextAssisted` |
| `Completion` | Authoritative transport state: `Complete`, `Refused`, `Truncated` or `ToolResponse`; only complete proceeds |
| `Languages`, `PrimaryLanguage` | The exact requested BCP 47 set and a primary member; canonical aliases cannot duplicate coverage |
| `Sources` | At most 100 already-authorized opaque source IDs with server-attested `ContextExtent`, `SourceReportedRadius` and `ViewpointAlignment` flags |

Visual context has no sources. Context-assisted mode does not itself authorize discovery or transmission; CH10 must construct the exact consented bundle after current access and source-eligibility checks. Model-written IDs, URLs and prose never extend that bundle.

Success returns an opaque `Proposal`. `Data()` decodes an independent typed `Document`; `MarshalJSON()` returns independent JSON bytes; `PolicyVersion()` returns `analysis-result-v1` separately from the model object. Every accessor rejects an uninitialized proposal. Mutating input buffers or returned nested pointers/slices cannot change the retained result. Numbers remain `json.Number`, including values near numeric boundaries; downstream conversion must be explicit. Language tags normalize, while accepted prose/statuses and numeric values retain their meaning.

Failure returns no proposal and a `*results.Failure` with a stable category and bounded `Finding{Code, Path}` entries. Categories include `invalid_context`, `incomplete_response`, `invalid_json`, `limit_exceeded`, `unsupported_schema`, `schema_violation`, `semantic_violation` and `unavailable`. Schema failures use a safe root path; semantic paths use only known fields and bounded indices. No raw library errors, injected keys, source IDs, coordinates or payload excerpts enter diagnostics. There is no repair, fallback language, partial acceptance or hidden retry.

## Enforced policy

- Located output selects an existing camera-bearing candidate; ambiguous output has at least two alternatives and no selection; unknown output stays unselected, including subject-only hypotheses.
- Camera and subject coordinates remain separate, finite and within exact WGS84 bounds. Zero is a value, not an absence marker.
- IDs are nonblank and unique; references are exact and nonduplicated. Country codes are null or assigned uppercase ISO alpha-2 codes, excluding historical, reserved and CLDR-only codes.
- Provided-context observations need Context-assisted mode and a nonempty authorized bundle; a candidate citing them needs an authorized source reference.
- Radius null pairs with unknown basis. Numeric radius needs appropriate referenced evidence; only point granularity permits zero, and city/region use null radius. These are uncalibrated proposals.
- Direction needs a camera viewpoint and referenced visual or server-attested alignment evidence. True-north azimuth is `[0,360)` and nullable uncertainty is `[0,180]`; a subject bearing alone is insufficient.
- The output language set matches the normalized requested set exactly. Complete text is nonblank with null unavailable reason; unavailable text is null with a nonblank reason. Candidate/scene basis must agree with candidate references. English is optional.

## Budgets

| Resource | Ceiling |
|---|---:|
| Raw model bytes / retained serialized proposal | 1 MiB each |
| Containers deep / decoded value and container nodes | 32 / 20000 |
| Decoded string / identifier or language tag (including normalization) | 8192 / 128 UTF-8 bytes |
| Number token / explicit exponent magnitude | 128 bytes / 324 |
| Observations / candidates | 100 / 20 |
| Requested languages / descriptions / authorized sources | 10 / 10 / 100 |
| Evidence refs / source refs per candidate | 100 / 100 |
| Uncertainty notes per candidate / warnings | 20 / 50 |
| Returned findings | 20 and 4096 total JSON bytes |

Duplicate keys, trailing values, invalid UTF-8, unpaired escaped surrogates, numeric overflow and nonzero underflow are rejected before schema validation. Collection and token bounds are enforced while parsing. An individually valid ceiling does not waive the total byte/node budget. Serialization escaping or language normalization cannot publish a proposal over the byte ceiling; that case fails the complete result with `limit_exceeded`.

## Downstream handoff

CH09 maps actual completion state and exact language/context inputs, distinguishes valid unknown from technical failure, and retains the canonical backend check even if transport schemas are simplified. CH11 owns the user/installation/job/asset envelope and immutable result persistence; model fields cannot supply those identities. Neither consumer may treat validation as access, transmission or write approval. CH15 and later review own safe text rendering and factual/provenance review. This package checks consistency and evidence membership, not real-world location accuracy.

See [verification evidence](engineering/ai-result-validation-verification.md) for scenario mapping, fuzz/race/container results, measurements and dependency-audit limits.
