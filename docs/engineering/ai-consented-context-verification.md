# Consented context verification

CH10 implements `add-consented-ai-context` against base `23c6ca6`. The implementation and reviews ran inline at the user's request, without subagents. New behavior followed one-test RED/GREEN/refactor cycles; characterization tests exercised existing shared behavior without manufacturing failures. Attempt records are in ignored `out/checks/ch10-apply/`.

## Scenario traceability

| OpenSpec scenario | Automated evidence |
|---|---|
| Only authorized classes are disclosed | `TestOnlyConsentedClassesAreDisclosed`, `TestContextAttemptSendsOnlyExactBundleOnce`, `TestAIContextReadsCurrentTargetWithoutUnconsentedSources` |
| Missing or mismatched consent prevents transmission | `TestMissingOrMismatchedConsentPreventsPreparation`, `TestContextBindingMismatchFailsBeforeEncoding`, `TestAIContextRejectsObservedChangesAtEveryAttemptBoundary` |
| Visual does not inherit a previous context choice | `TestVisualRejectsContextInsteadOfInheritingIt` |
| Offset and missing-time meanings survive | `TestCaptureTimePreservesOffsetAndMissingSemantics`, `TestEmptyAuthorizedContextRemainsExplicit`, neighbor eligibility tests |
| Context limits are deterministic | `TestNeighborLimitAndOrderingAreDeterministic`, `TestInvalidContextBoundsFailWithoutPartialBundle` |
| Invalid bounds do not silently alter user input | `TestInvalidContextBoundsFailWithoutPartialBundle`, `TestAIContextMetadataReadRejectsMalformedAndUnboundedBodies` |
| Eligible sources retain their provenance | `TestAIContextUsesOnlyCurrentEligibleMetadataNeighbors`, `TestNeighborProvenanceStaysInPrivateEnvelope` |
| Ineligible or unrelated sources are excluded | `TestIneligibleAndAIOriginNeighborsAreExcluded`, `TestAIContextUsesOnlyCurrentEligibleMetadataNeighbors`, both optional-neighbor omission tests |
| Album context cannot broaden the authorized scope | `TestSelectedAlbumCannotBroadenScope`, `TestAIContextSelectedAlbumUsesExactCurrentMembership`, `TestAIContextAlbumLossNeverSubstitutesAnotherAlbum` |
| Former album membership omits optional context | `TestAIContextAlbumLossNeverSubstitutesAnotherAlbum` |
| Empty authorized context remains explicit | `TestEmptyAuthorizedContextRemainsExplicit`, `TestEmptyContextKeepsRequestedModeThroughResult` |
| Frozen-source or authority change discards the attempt | `TestAIContextRejectsFrozenCaptureChangesWithoutReplacingBundle`, `TestAIContextRejectsObservedChangesAtEveryAttemptBoundary` (18 boundary cases) |
| Caller mutation cannot widen disclosure | `TestFrozenBundleAndProvenanceOwnTheirData`, `TestContextResultOwnsExactPrivateProvenance` |
| Valid contextual output retains separate evidence kinds | `TestContextCanonicalEvidenceCannotInventSourceOrGeometryAuthority/valid` |
| Invented evidence or unsupported geometry fails | Remaining six cases in `TestContextCanonicalEvidenceCannotInventSourceOrGeometryAuthority` |
| Resource failure or cancellation cannot create partial success | `TestContextResourceAndReservationFailuresNeverPublish` (eight cases), concurrent canceled/failed integration |
| Provenance survives the transient handoff | `TestContextResultOwnsExactPrivateProvenance`, `TestAIContextAnalyzesThroughCurrentProfileAndSourceAuthority` |
| Failures and execution remain private and read-only | `TestAIContextConcurrentAttemptsNeverWriteDatabaseOrImmich`, projection/default-format privacy tests |

Additional adapter tests reject ambiguous capture and album-membership JSON. Real file-backed SQLite, query-only connections, upstream operation counters and the actual provider codec/dispatcher exercise the integration boundary. Root fixtures use synthetic images/accounts and authenticated local HTTP services. No private image, production provider or live library is used.

## Verification status

Focused contextual/analysis and root Context/Visual suites pass. The concurrent integration suite also passes with the race detector. The backend Docker build includes the new package and passes. All thirteen shared gates pass against `23c6ca6`. Backend AI statement coverage is **88.70% (2990/3371)** against the 80% floor; all 28 frontend tests and seven browser journeys pass. Lint retains three inherited warnings. The initial full run overlapped the final neighbor-eligibility RED cycle; the complete Go race/coverage gate was rerun after GREEN. Chromium initially failed to create its macOS Mach port inside the sandbox; the complete smoke gate passed outside it. Neither failed attempt is counted as passing evidence.

GitNexus was bound to this checkout and base revision. Pre-edit impact for result metadata was LOW; unresolved runner/adapter receiver calls returned UNKNOWN and were checked against source and existing Visual callers/tests. The refreshed PDG index reports complete cycle enumeration with zero cycles. Taint queries do not model all closures, callbacks or fields, so empty findings are not proof of safety. Staged change analysis returned all **404/404 changed symbols across 71 files and 11 affected flows**, with no partial or truncated result and aggregate **HIGH** risk. Reviewed flows cover neighbor owner credentials/current eligibility and the shared attempt, language, metadata and canonical evidence validation. Existing Visual and new Context authority, framing, race and zero-write tests cover these boundaries. Source confirms no public/startup consumer or writer dependency. Documentation links and archived requirement preservation were checked independently.

## Synthetic comparison and limits

| Dimension | Visual fixture evidence | Context-assisted fixture evidence |
|---|---|---|
| Outcome coverage | Existing located/ambiguous/unknown cases remain accepted | Valid located and unknown cases accepted, including explicitly empty context |
| Camera coordinate error | Not measured: returned coordinates are scripted | Not measured: returned coordinates are scripted |
| Direction quality | Canonical direction rules/regressions pass; azimuth accuracy unmeasured | Visual direction remains subject to canonical rules; context cannot grant alignment authority |
| Misleading context | Context input rejected by Visual entry point | Invented sources and unsupported radius/alignment claims rejected; unknown lineage remains unverified |
| Calls and side effects | Existing one-call and zero-write tests retained | One reservation/call; failed/canceled/concurrent attempts do not write SQLite or Immich |

This paired synthetic contract comparison does not estimate model quality or establish that a misleading hint will be ignored by a real model. Geographic coverage, independent camera error, direction accuracy, latency and misleading-context behavior still require an authorized evaluation dataset and recorded endpoint/model configuration. Full-schema provider compatibility, production admission and reference-NAS throughput remain outside this internal change.
