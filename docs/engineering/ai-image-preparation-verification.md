# CH08 image preparation verification

Implemented on top of CH07 commit `4a6869016df9ce533ef7a3af329366453c32713b` on 20 September 2026. Work follows sequential inline OpenSpec Plus planning/apply/TDD and self-review, without subagents. All new repository text is English.

The [caller contract](../ai-image-preparation.md) describes the internal read-only boundary, budgets, supported formats and transient ownership. No public endpoint, migration, provider request or production dispatch is added. CH09 consumes this contract; durable admission is connected later by CH12.

## Scenario traceability

All five delta requirements and 23 scenarios map to the following automated evidence. Raster test names are in `backend/internal/ai/images`; names beginning `TestAIImage` are root backend integration tests with real temporary SQLite and synthetic HTTP services.

| Scenario | Automated evidence |
|---|---|
| Eligible authorized image is prepared | `TestAIImagePreparesExactAuthorizedPreview`: exact owner/installation/asset, current personal credential, source digest and normalized dimensions |
| Another user's local asset cannot be fetched | `TestAIImageLocalDenialsNeverReachImmich`: foreign owner and missing asset yield zero requests |
| Disabled or stale installation cannot prepare an image | Same local-denial test: disabled AI, missing account/key and mismatched/persisted installation |
| Current upstream access or eligibility is lost | `TestAIImageUpstreamEligibilityMustBeCurrentAndComplete`, `TestAIImageFailuresAreSafeAndNeverRetried`: hidden/locked/trashed/non-image/stack child, required revision fields and 401/403/404 |
| Invalid identity cannot choose a network path | Local-denial test: URL/path traversal and malformed installation fail before requests |
| Exact bounded preview retrieval succeeds | Exact-preview fixture and `TestAIImageConcurrentPreparationNeverWritesPersistenceOrUpstream`: only exact metadata/preview GETs |
| Redirect cannot forward credentials | `TestAIImageRejectsEveryRedirectWithoutForwardingCredentials`: metadata/preview, same/cross host, zero redirected requests |
| Declared and streamed sizes are independently bounded | `TestAIImageTransportBoundsDeclaredAndStreamedBytes`: excess declaration, unknown/misleading length, limit+1 and exact boundary; bodies close |
| Cancellation and timeout stop retrieval | `TestAIImageOverallDeadlineAndCancellationReleaseRetrieval`, `TestAIImageCancellationOfFinalMetadataReadPreventsPublication`: before-start, metadata, preview and final metadata interruptions |
| Failed upstream response is not retried or exposed | Failure fixture and `TestAIImageInterruptedStreamClosesBodyWithoutPrivateError`: malformed bodies, 429/500, interrupted read, safe errors and exact attempt counts |
| Orientation and aspect ratio survive normalization | `TestPrepareRasterNormalizesEveryEXIFOrientation`, `TestPrepareRasterDownscalesWithoutCropping`: all eight transforms, visible corner colors and proportional bounds |
| Small image is not enlarged | `TestPrepareRasterRejectsDecodedAndSourceBudgetOverflow`: exact allowed input remains 8x4 |
| Private source metadata is absent from the copy | `TestPrepareRasterDropsPrivateMetadataWithoutChangingInput`: original byte equality; EXIF/private comments/non-JFIF metadata absent |
| Decode budgets reject oversized dimensions before full decode | Decoded/source-budget fixture: source, dimensions, pixels, policy and inclusive boundaries; configuration inspection precedes raster decoding |
| Unsupported or malformed images fail explicitly | `TestPrepareRasterRejectsUnsupportedContainersAndMIME`, `TestPrepareRasterRejectsTextualAndCompressedOrientationMetadata`: static WebP success, MIME mismatch, malformed JPEG, invalid EXIF, APNG, animated/EXIF WebP, PNG EXIF and XMP/text orientation |
| Encoded request-image limit includes transmission expansion | `TestPrepareRasterBudgetIncludesBase64AndPrefix`: one-byte-over rejects and exact full representation succeeds |
| Transparency has deterministic output | `TestPrepareRasterCompositesTransparencyOverWhite`: transparent region white, opaque pixels correctly placed |
| Account or credential changes during retrieval | `TestAIImageRejectsLocalAuthorityChangesBeforePublication`: barrier-controlled account, credential, hidden state, installation and visible-library change; no further request after detected authority loss |
| Source changes while bytes are retrieved | `TestAIImageRejectsChangedSourceRevisionAndEligibility`: revision/checksum/owner/visibility/type/stack differences reject with no retry |
| Cancellation during processing prevents success | `TestPrepareRasterNeverPublishesAfterCancellation`: initial and later processing-boundary cancellation; final root cancellation fixture also rejects publication |
| Returned data is independent | `TestPreparedOwnsCopiesAndReleaseInvalidatesAliases`, `TestPreparedRequiresCompleteBindingAndIdentifiesPolicy`, `TestPreparedConcurrentReadersAndRelease`: copies, zero/nil values, source/policy binding and synchronized release |
| Failure and serialization do not disclose private image data | `TestPreparedDefaultFormattingDoesNotDisclosePayload`, safe HTTP/stream failures: default JSON/Go formatting and errors omit bytes, credentials, URLs and metadata |
| Preparation has no persistent or mutating effects | Eight concurrent operations for success/failure/cancel under SQLite `query_only`, same-connection `total_changes` and exact GET counters; source dependencies expose no provider/writer; legacy regression is covered by the shared gates |

Additional regressions reject compressed HTTP responses and unsafe configured URLs, duplicate/case-variant eligibility fields, final metadata-read failure classification, WebP canvas/bitstream dimension mismatch, and cumulative PNG text decompression beyond one 64 KiB per-image budget. The graph identifies image-only integration; unresolved fixture receiver types are supplemented by source inspection and behavioral tests.

## Verification results

The final `bun run check --base 4a68690` completed successfully after the last production correction: all thirteen gates pass, including race-enabled backend tests, both production builds and all seven browser journeys. AI Go statement coverage is **88.93% (2032/2285)** against the 80% floor. The three inherited lint warnings remain unchanged. Scoped tests, race and vet also passed during the implementation slices.

The backend production Docker image builds successfully. Its current image-package test binary passes in that image with networking disabled, a read-only filesystem and only the synthetic fixture mounted; source and documentation trees are absent. No new image service or native runtime dependency is required.

The final fuzz run completed **1,544,303 executions in 20.600 seconds** after aggregate PNG text-budget hardening, with no failure corpus. Fuzz inputs are bounded to 64 KiB and 4096 decoded pixels; output properties require complete JPEG, dimensions and transmission budgets. It is malformed-input evidence, not an exhaustive decoder audit or a NAS performance measurement.

The refreshed GitNexus PDG index covers this checkout and its uncommitted source. Cycle enumeration is complete with zero cycles. Graph queries identify the image-only paths; unresolved Go fixture receiver edges are supplemented by source inspection and behavioral tests. The final staged change analysis returns all 190 observed symbols across 45 files, six affected image-preparation flows and HIGH aggregate risk, with no partial/truncated response. Those flows reach credential decryption, owner-scoped catalog reads, selection eligibility, source validation and bounded fetch; each boundary was reviewed against source and the passing authorization/freshness/side-effect tests. Existing production symbols are unchanged; source search confirms no public/startup consumer. This does not remove the documented unresolved receiver-edge limitation. All seven tasks are complete; the five added requirements are synchronized into the maintained contract (13 requirements, 50 scenarios) and the change is archived.

## Dependencies and limits

The owner approved `github.com/disintegration/imaging v1.6.2` and `golang.org/x/image v0.45.0`. The selected module graph also upgrades existing x/text to v0.41.0, x/sys to v0.47.0 and x/sync to v0.22.0 (required by x/text). Go remains 1.25.14; the module directive is normalized to 1.25.0. `go mod tidy` and `go mod verify` passed.

The whole-backend `govulncheck v1.7.0 -show verbose ./...` reports zero called vulnerabilities and zero findings in imported packages, with 18 findings in unused portions of existing required modules. The x/text/x/sys upgrades remove the two previously recorded module findings; unrelated x/crypto and mongo-driver findings remain as recorded in [CH07 verification](ai-result-validation-verification.md#dependency-audit). This is reachability evidence, not a claim that the dependency tree has no advisories.

Pure-Go packaging, deployment compatibility and private-data boundaries are assessed separately. The raster test fixture `static.webp` is a locally generated uniform synthetic image made with already-installed Sharp, not a user photo or downloaded asset. No live Immich/provider request or production deployment occurred. GATE-02 deployed image compatibility and reference-NAS performance remain unverified. CPU library operations are not preemptible mid-call; stage/final checks reject canceled results. Separate metadata reads detect observed changes but do not make the upstream source atomic. Memory cleanup does not promise cryptographic erasure of all library allocations.
