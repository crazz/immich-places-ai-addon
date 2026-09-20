## Context

See [proposal.md](proposal.md) for motivation and [the delta](specs/ai-location-proposals/spec.md) for behavior. Existing Immich preview handlers stream bytes through a retrying client with ordinary redirect behavior. CH05 supplies a persistent installation binding and owner-scoped catalog eligibility. Application user IDs are independent of Immich owner IDs. There is no current image normalization pipeline or analysis dispatch consumer.

The adopted [architecture](../../../../docs/engineering/architecture.md), [testing](../../../../docs/engineering/testing.md) and [coding standards](../../../../docs/engineering/coding-standards.md) govern this change. The owner selected the pure-Go imaging/x/image approach; exact versions, alternatives and maintenance trade-offs are recorded in [library evaluation](library-evaluation.md).

## Goals / Non-Goals

**Goals:** Keep deterministic raster transformation separate from authenticated I/O; make a prepared image an explicitly transient, independently owned value; preserve the single Go backend and existing installation identity; establish a testable internal handoff for CH09.

**Non-Goals:** No provider dispatch, durable image storage, migration, public analysis route, new service, original-file fallback, high-resolution retry or changes to legacy preview/manual/GPX paths. This internal result does not grant consent or write authority.

## Decisions

### Architecture and ownership

`backend/internal/ai/images/` owns deterministic format policy, limits, raster preparation and an opaque prepared value. It uses the approved imaging library and standard codecs, with x/image for WebP. It does not import HTTP, SQL, root services, provider or writer packages. Root `aiImage*.go` files own narrow Immich read operations and local authorization using existing database services. This retains ADR-07's dependency direction without a service boundary or a generic image framework.

The root preparer receives the database, existing selection installation state, configured Immich endpoint and a dedicated bounded HTTP client. It accepts application owner, installation and exact asset identifiers, never a source URL. It remains an internal consumer contract until durable admission is connected by CH12. No application startup or public endpoint dispatches preparation in CH08.

### Authorization and source binding

Validate canonical UUID asset and installation identities before path construction; application owner IDs remain bounded opaque local identifiers. A short local check reads enabled state, the persisted installation ID, the current account credential and owner-scoped catalog asset. Reuse selection eligibility for unavailable/non-image/hidden/stack-child policy; hidden libraries remain excluded by the catalog query. Never compare an upstream owner ID with the unrelated application user ID. Successful current personal-token access to the exact upstream asset, combined with local ownership, establishes the read boundary.

Read current upstream metadata before the preview and again after transformation. Require the exact asset ID, IMAGE type, timeline/archive visibility, non-trashed state and either no stack or primary-asset identity. Strictly require the documented revision-bearing fields; reject incomplete or malformed metadata. Bind the source to a digest of ID, upstream owner, checksum, updatedAt and stack membership facts, without retaining raw metadata. A detected change invalidates the complete preparation.

Recheck local authority before every subsequent upstream read and immediately before publication, comparing the credential and relevant local eligibility facts with the first check. Database calls finish before network calls; no transaction spans I/O. Later consumers must repeat current authority and consent checks; a successful preparation cannot eliminate the remote race after the final read.

### Authenticated retrieval and resource budgets

Use a dedicated non-retrying client with every redirect rejected, including same-origin redirects. The only operations are exact `GET /api/assets/{id}` and `GET /api/assets/{id}/thumbnail?size=preview` relative to the configured endpoint. The credential is an `x-api-key` header confined to that destination. Reject configured URL userinfo, query and fragment rather than constructing ambiguous credential-bearing requests. Do not reuse the legacy retrying client or forward an upstream redirect URL.

Initial internal policy uses a 30-second overall deadline, 1 MiB metadata limit and 20 MiB preview-source limit. Declared lengths and actual streamed bytes are checked independently with limit-plus-one reads; all bodies close on every outcome. Compressed responses are not accepted, so wire byte bounds cannot be bypassed by automatic decompression. Non-success responses are discarded without returning their bodies or URLs. The transport has finite header/connection timing and no environment proxy inheritance that could redirect the user's credential.

Raster defaults follow the technical design: 2048-pixel long edge, 40 million decoded pixels and 10 MiB maximum data-URL representation including prefix and Base64 expansion. A 40000-pixel single-dimension ceiling also bounds extreme aspect ratios. Limits are explicit internal policy values with validated hard ceilings; operators can lower them when this consumer is exposed. No environment/configuration switches are added before that consumer exists.

### Raster policy and transient value

Allow only static JPEG, PNG and WebP with matching declared MIME. Inspect container structure and dimensions before full decoding. Reject animated PNG/WebP, unsupported formats, malformed/truncated structures and orientation-bearing PNG/WebP metadata that imaging cannot normalize. PNG text is inspected with a shared 64 KiB expanded-text budget across all chunks; text/XMP orientation unsupported by the normalizer also fails. JPEG EXIF orientation uses imaging's established transform; malformed or unsupported orientation data fails instead of silently producing the wrong scene. Newly encoded JPEG contains only transformed pixels, with transparency composited over white and fixed quality 85. Preserve aspect ratio, do not upscale or crop, and perform one encoding attempt.

Check context before and after bounded CPU stages. Library transforms cannot be preempted mid-call; do not spawn abandoned transform goroutines. The decoded pixel ceiling bounds their work, and cancellation always prevents publication. Encode through a bounded writer, then account for the exact transmission expansion before retaining bytes.

The prepared value keeps bytes and binding fields private, exposes independent copies and bounded metadata, and is unusable in its zero value. Bind it to owner, installation, asset, source digest and a versioned preparation policy. Its digest identifies the encoded output. Default JSON and string/error formatting must not reveal bytes or private identities. Explicit release clears retained bytes and invalidates further payload access; concurrent access/release is synchronized. Only memory is used, so no filesystem cleanup or startup image recovery is needed. Garbage collection is not a guarantee of cryptographic erasure of all decoder allocations or caller-owned copies.

### Failure handling and observability

Expose a small bounded error vocabulary for invalid input, denied/stale authority, unsupported image, exceeded limits, unavailable upstream and interruption. Never wrap raw HTTP, decoder, SQL or credential errors into user-visible failures. All failures return no prepared value. There are no hidden retries, alternate assets, fallback originals, provider calls or persistence writes. Callers can record safe error categories and bounded dimensions/durations; the preparer itself logs no payloads.

### Verification and rollout

Pure Go tests exercise orientation, aspect ratio, metadata removal, transparency, byte/pixel/edge boundaries, format rejection, cancellation and independent ownership. Root integration tests use real temporary SQLite with existing migrations plus synthetic local HTTP servers to verify account/installation scope, credentials, revision changes, redirects, limits, timeouts and zero mutations. Tests distinguish the pinned Immich source contract from live compatibility. Every delta scenario maps to named tests in the verification record.

Follow OpenSpec Plus TDD and inline spec-compliance then quality review, per the owner's no-subagent instruction. Run scoped tests during each slice, then the installed thirteen verification gates, dependency audit and backend container build because modules change. Package an offline image test binary in the backend container to exercise the pure-Go deployment. Record actual coverage and residual audit findings; no passing unit suite claims real-provider quality or deployed Immich compatibility.

Rollout is additive: the module and internal adapter are inert until an authorized later consumer uses them. No migration or existing route change is needed. Reverting this change restores the dependency set and removes unused internal code without altering user data.

## Risks / Trade-offs

- Imaging's old release and JPEG-only orientation support → pin approved versions, audit reachable dependencies, exercise all eight EXIF transforms, and reject unsupported orientation-bearing containers.
- Hostile compressed rasters and CPU latency → bounded source, dimensions, pixels, output and deadline; cancellation is checked at stage boundaries, not claimed to preempt library CPU loops.
- Source/access changes during reads → before/after metadata and local authority checks; later dispatch still reauthorizes because separate remote requests cannot provide an atomic asset snapshot.
- Metadata removal may discard color profiles → deterministic JPEG output is an analysis copy, with visual quality evaluated separately before release.
- Pinned Immich DTOs differ from deployment → fail closed on missing eligibility/revision fields and retain GATE-02 live image compatibility as unverified until separately tested.
