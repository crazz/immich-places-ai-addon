# Verification plan

Status: implemented and verified. The table identifies executed automated evidence; the implementation record contains gate results and the separate synthetic NAS compatibility check. Model geolocation quality on private photos was not measured.

## Scenario mapping

Use the applicable spec scenario as the test case. New behavior follows Plus TDD. Preserved v1 behavior can be characterized against unchanged code without manufacturing a failure. A scenario with several independent branches needs assertions or cases for each branch.

| Capability and scenario | Tasks | Automated verification boundary |
|---|---|---|
| `ai-analysis-jobs` — Valid submission survives reopening | 1.2, 2.1–2.2, 3.1–3.3 | TestAIJobSubmissionSurvivesReopenWithExactMembership |
| `ai-analysis-jobs` — Invalid or foreign submission is rejected atomically | 1.2, 2.1–2.2, 3.1–3.3 | TestAIJobInvalidSubmissionLeavesNoPartialWork; TestAIProductionRejectsIncompleteConsentAndAuthorityAtomically |
| `ai-analysis-jobs` — Insertion failure rolls back membership | 1.2, 2.1–2.2, 3.1–3.3 | TestAIJobInsertionFailureRollsBackHeaderAndItems; TestAIProductionAdmissionStorageFailureRollsBackEveryRecord |
| `ai-analysis-jobs` — JB01 Freeze Research inputs and deduplicate submission | 1.2, 2.1–2.2, 3.1–3.3 | TestAIResearchIdempotencyBindsExactHintWithoutAdditionalWork; TestAIResearchFreezesDisplayedAlbumAndCaptureBeforeCatalogChanges |
| `ai-analysis-jobs` — Launch Visual without hidden context | 1.2, 2.1–2.2, 3.1–3.3 | TestVisualRejectsContextInsteadOfInheritingIt; LaunchForm.test.tsx |
| `ai-analysis-jobs` — Choose bounded Context-assisted inputs | 1.2, 2.1–2.2, 3.1–3.3 | TestContextAttemptSendsOnlyExactBundleOnce; TestAIProductionAdmitsOnlyExplicitContextPolicy |
| `ai-analysis-jobs` — Bind the current choices to the Start action | 1.2, 2.1–2.2, 3.1–3.3 | LaunchForm.test.tsx; ResearchLaunch.test.tsx |
| `ai-analysis-jobs` — JB02 Start Research with ordinary inputs | 1.2, 2.1–2.2, 3.1–3.3 | ResearchLaunch.test.tsx — starts Research by default with the exact hint and no extra confirmation or hidden context; TestAIResearchPublishesCoarsePrivateHistoryWithOrdinaryDefaults |
| `ai-analysis-jobs` — JB03 Keep existing mode choices usable | 1.2, 2.1–2.2, 3.1–3.3 | LaunchForm.test.tsx — starts the displayed Visual run with one click and binds the latest choices; tests/e2e/workflows.spec.ts — launches and reconciles without randomUUID, restores progress and starts Context reanalysis with application defaults |
| `ai-analysis-jobs` — JB04 Support a longer investigation without losing ownership | 1.2, 2.1–2.2, 3.1–3.3 | TestAIResearchCarriesOneTenMinuteDeadlineThroughDispatch; TestAIResearchHonorsShorterInstallationDeadline; TestAIResearchHeartbeatPreservesDeadlineAndCancellationFencesLateOutput |
| `ai-analysis-jobs` — JB05 Fence canceled or stale completion | 1.2, 2.1–2.2, 3.1–3.3 | TestAIResearchHeartbeatPreservesDeadlineAndCancellationFencesLateOutput; TestAIJobHeartbeatRequiresExactUnexpiredLease; TestAIJobInstallationRotationFencesUnfinishedWorkAtomically |
| `ai-analysis-jobs` — JB06 Recover without replay under ordinary defaults | 1.2, 2.1–2.2, 3.1–3.3 | TestAIResearchRestartNeverResendsAfterDefaultReservation |
| `ai-analysis-jobs` — JB07 Reopen a coarse result with answer references | 1.2, 2.1–2.2, 3.1–3.3 | TestAIResearchHistorySurvivesReopenSourceLossAndDisabledExecution; TestAIResearchAndLegacyAnswersSurviveHistoryMigrationAndReopen |
| `ai-analysis-jobs` — JB08 Preserve atomicity and ownership | 1.2, 2.1–2.2, 3.1–3.3 | TestAIResearchPublicationRollsBackAnswerReferencesAndHistoryTogether; TestAIResearchHistorySurvivesReopenSourceLossAndDisabledExecution |
| `ai-analysis-jobs` — JB09 Clean up answer references with history | 1.2, 2.1–2.2, 3.1–3.3 | TestAIResearchHistorySurvivesReopenSourceLossAndDisabledExecution; TestAIJobCleanupIsBoundedPrivateAndTerminalOnly |
| `ai-location-proposals` — Canonical fixtures remain valid | 1.1, 2.1–2.2 | TestCanonicalFixturesRemainUnchanged |
| `ai-location-proposals` — Structurally invalid or ambiguous output fails closed | 1.1, 2.1–2.2 | TestAmbiguousOrMalformedJSONCannotBecomeAProposal; TestClosedSchemaFieldsAreRequiredAndTyped |
| `ai-location-proposals` — Numeric and resource bounds are enforced | 1.1, 2.1–2.2 | TestParseBoundsBytesDepthNodesAndStrings; TestNumbersAreBoundedWithoutLosingExactRanges; TestCollectionsAndIdentifiersHaveIndividualBounds |
| `ai-location-proposals` — Transport failure cannot masquerade as a completed proposal | 1.1, 2.1–2.2 | TestTransportCompletionIsAuthoritative |
| `ai-location-proposals` — LP01 Accept a complete Research result without changing old results | 1.1, 2.1–2.2 | TestResearchPreservesExistingValidationAndUncertainty; TestCanonicalFixturesRemainUnchanged; TestAIResearchAndLegacyAnswersSurviveHistoryMigrationAndReopen |
| `ai-location-proposals` — Local references resolve and preserve their kinds | 1.1, 2.1–2.2 | TestProvidedContextRetainsAnAuthorizedSourceConnection; TestTypedCameraAndSubjectCoordinatesStaySeparate |
| `ai-location-proposals` — Dangling and duplicate identities are rejected | 1.1, 2.1–2.2 | TestInvalidIdentitiesAndLocalReferencesCannotBeInvented |
| `ai-location-proposals` — Foreign or invented source claims confer no authority | 1.1, 2.1–2.2 | TestSourcesMustBelongToThisAuthorizedBundle |
| `ai-location-proposals` — Visual output cannot claim supplied context | 1.1, 2.1–2.2 | TestSourcesMustBelongToThisAuthorizedBundle |
| `ai-location-proposals` — Context observations retain a source connection | 1.1, 2.1–2.2 | TestProvidedContextRetainsAnAuthorizedSourceConnection |
| `ai-location-proposals` — LP02 Accept Research sources from the answer | 1.1, 2.1–2.2 | TestResearchRetainsAnswerSources; TestResearchRejectsDuplicateAnswerSourceIdentity; TestResearchBoundsSourceTextBytes |
| `ai-location-proposals` — LP03 Keep coordinates when a source link is unusable | 1.1, 2.1–2.2 | TestResearchKeepsEstimateWithMissingReference; TestAIResearchDetailKeepsCoordinatesWithoutExposingUnsafeURL |
| `ai-location-proposals` — Supported visual estimate or unknown radius is retained | 1.1, 2.1–2.2 | TestSupportedRadiusStatesAreRetained |
| `ai-location-proposals` — Contradictory or misleading precision is rejected | 1.1, 2.1–2.2 | TestRadiusClaimsNeedConsistentGranularityAndBasis |
| `ai-location-proposals` — Radius provenance requires the matching evidence kind | 1.1, 2.1–2.2 | TestSupportedRadiusStatesAreRetained |
| `ai-location-proposals` — LP04 Retain a 500-meter estimate | 1.1, 2.1–2.2 | TestResearchRetains500MeterEstimate |
| `ai-location-proposals` — LP05 Retain coarse city and region estimates | 1.1, 2.1–2.2 | TestResearchRetainsCoarseCityEstimate; TestResearchPreservesExistingValidationAndUncertainty |
| `ai-location-proposals` — LP06 Distinguish unknown error from invalid numeric data | 1.1, 2.1–2.2 | TestResearchPreservesExistingValidationAndUncertainty |
| `ai-location-proposals` — Only authorized classes are disclosed | 1.1, 2.1–2.2 | TestContextAttemptSendsOnlyExactBundleOnce; TestAIResearchRejectsNeighborDisclosure |
| `ai-location-proposals` — Missing or mismatched consent prevents transmission | 1.1, 2.1–2.2 | TestAIProductionRejectsIncompleteConsentAndAuthorityAtomically; TestAdmissionRequiresExactContextConsent |
| `ai-location-proposals` — Visual does not inherit a previous context choice | 1.1, 2.1–2.2 | TestVisualRejectsContextInsteadOfInheritingIt |
| `ai-location-proposals` — LP07 Use only displayed Research context | 1.1, 2.1–2.2 | TestAIResearchFreezesDisplayedAlbumAndCaptureBeforeCatalogChanges; TestResearchRejectsNeighborAuthorizationBeforeReservation |
| `ai-location-proposals` — LP08 Keep a weak but useful preferred estimate | 1.1, 2.1–2.2 | TestResearchPreservesExistingValidationAndUncertainty |
| `ai-location-proposals` — LP09 Preserve ambiguous candidates or genuine unknown | 1.1, 2.1–2.2 | TestResearchPreservesExistingValidationAndUncertainty |
| `ai-provider-configuration` — PC01 Launch through the existing provider | 1.2, 3.1, 4.2 | TestAIResearchPublishesCoarsePrivateHistoryWithOrdinaryDefaults; live synthetic proxy check (HTTP 200, v2 semantic validation) |
| `ai-provider-configuration` — PC02 Preserve useful answers without telemetry | 1.2, 3.1, 4.2 | TestResearchRetainsAnswerSources; TestResearchKeepsEstimateWithMissingReference; live synthetic proxy check |
| `ai-provider-configuration` — PC03 Report a concrete transport limitation | 1.2, 3.1, 4.2 | TestErrorsDoNotCauseHiddenRetriesOrLeaks; TestVisualResponseRequiresOneOrdinaryCompleteAssistant; inspected proxy timeout configuration (120,000 ms) |
| `ai-provider-configuration` — PC04 Use ordinary bounded defaults | 1.2, 3.1, 4.2 | TestAIResearchPublishesCoarsePrivateHistoryWithOrdinaryDefaults; TestAIResearchTimeoutConfigurationIsFiniteAndDefaultsToTenMinutes |
| `ai-provider-configuration` — PC05 Enforce actual authority and execution limits | 1.2, 3.1, 4.2 | TestAIResearchRejectsExplicitContextRestriction; TestAIProductionRechecksConsentAndPolicyAfterResolution; TestRejectChangedMixedOrAbsentDNSAnswers |
| `ai-results-and-review` — Inspect source-free Visual evidence | 1.3, 2.3, 3.3 | EvidencePanel.test.tsx; tests/e2e/review.spec.ts — inspects stored camera geometry offline without changing pending manual coordinates |
| `ai-results-and-review` — Inspect contextual lineage | 1.3, 2.3, 3.3 | EvidencePanel.test.tsx; TestAIResultContextDetailRetainsFrozenEvidenceWithoutFreshSourceReads |
| `ai-results-and-review` — Handle unsafe text or unresolved references | 1.3, 2.3, 3.3 | EvidencePanel.test.tsx; ResearchReview.test.tsx — keeps coordinates with missing or unsafe references and renders source text inertly |
| `ai-results-and-review` — RV09 Review answer-provided Research references | 1.3, 2.3, 3.3 | ResearchReview.test.tsx — shows answer-provided references as deliberate links beside the proposed coordinates; shows optional answer links even when the candidate omits a local source reference |
| `ai-results-and-review` — RV01 Review a 500-meter estimate | 1.3, 2.3, 3.3 | ResearchReview.test.tsx — shows answer-provided references as deliberate links beside the proposed coordinates |
| `ai-results-and-review` — RV02 Review a city or region estimate | 1.3, 2.3, 3.3 | ResearchReview.test.tsx — shows Research coordinates and estimated error even with a degraded map; keeps enormous estimated errors compact and numeric coordinates visible |
| `ai-results-and-review` — RV03 Inspect unknown error and ambiguous alternatives | 1.3, 2.3, 3.3 | ResearchReview.test.tsx — keeps coordinates with missing or unsafe references and renders source text inertly; tests/e2e/review.spec.ts — reviews Research estimates and references by keyboard on a narrow offline map without writes |
| `ai-results-and-review` — RV04 Use results without maps or pointer input | 1.3, 2.3, 3.3 | tests/e2e/review.spec.ts — reviews Research estimates and references by keyboard on a narrow offline map without writes; reviewMapLayers.test.ts; reviewGeometry.test.ts |
| `ai-results-and-review` — RV05 Follow a source from the answer | 1.3, 2.3, 3.3 | ResearchReview.test.tsx — shows answer-provided references as deliberate links beside the proposed coordinates |
| `ai-results-and-review` — RV06 Preserve coordinates with absent or unsafe links | 1.3, 2.3, 3.3 | ResearchReview.test.tsx — keeps coordinates with missing or unsafe references and renders source text inertly; TestResearchReferencesAreInertAndExcludeUnsafeDestinations |
| `ai-results-and-review` — RV07 Retain history when a reference disappears | 1.3, 2.3, 3.3 | TestAIResearchHistorySurvivesReopenSourceLossAndDisabledExecution |
| `ai-results-and-review` — RV08 Preserve legacy display and account isolation | 1.3, 2.3, 3.3 | TestAIResearchAndLegacyAnswersSurviveHistoryMigrationAndReopen; TestAIResultOwnerAndInstallationIsolationIncludesEveryRead; AuthenticatedAIWorkspace.test.tsx |
| `ai-web-research` — WR01 Return an approximate camera location | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | TestResearchRetains500MeterEstimate; TestResearchSendsOneImageAndHintAndReturnsCoarseAnswer |
| `ai-web-research` — WR02 Keep estimates larger than 500 meters | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | TestResearchRetainsCoarseCityEstimate; TestResearchPreservesExistingValidationAndUncertainty |
| `ai-web-research` — WR03 Challenge an incorrect hint | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | TestResearchSendsOneImageAndHintAndReturnsCoarseAnswer (misleading-hint instruction and accepted returned candidate; actual model accuracy is not asserted) |
| `ai-web-research` — WR04 Preserve uncertainty without withholding useful candidates | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | TestResearchPreservesExistingValidationAndUncertainty |
| `ai-web-research` — WR05 Retain answer-provided links | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | TestResearchRetainsAnswerSources; TestAIResearchPublishesCoarsePrivateHistoryWithOrdinaryDefaults |
| `ai-web-research` — WR06 Keep coordinates when sources are absent | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | TestResearchKeepsEstimateWithMissingReference; ResearchReview.test.tsx — keeps coordinates with missing or unsafe references and renders source text inertly |
| `ai-web-research` — WR07 Keep uncertain source claims reviewable | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | ResearchReview.test.tsx — shows answer-provided references as deliberate links beside the proposed coordinates |
| `ai-web-research` — WR08 Send only displayed inputs | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | TestResearchSendsOneImageAndHintAndReturnsCoarseAnswer; TestAIResearchFreezesDisplayedAlbumAndCaptureBeforeCatalogChanges; TestAIResearchRejectsNeighborDisclosure |
| `ai-web-research` — WR09 Retain read-only behavior | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | TestResearchSendsOneImageAndHintAndReturnsCoarseAnswer; TestAIResearchPublishesCoarsePrivateHistoryWithOrdinaryDefaults; tests/e2e/review.spec.ts — reviews Research estimates and references by keyboard on a narrow offline map without writes |
| `ai-web-research` — WR10 Distinguish an incomplete response from a coarse answer | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | TestTransportCompletionIsAuthoritative; TestVisualResponseRequiresOneOrdinaryCompleteAssistant; TestResearchPreservesExistingValidationAndUncertainty |
| `ai-web-research` — WR11 Compare estimates against known camera locations | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | scripts/ai/research-comparison.test.mjs — measures independently known camera error separately from the model estimate; rejects unmatched or invalid measurements instead of producing a misleading comparison; requires known reference uncertainty for measured error and retains that uncertainty |
| `ai-web-research` — WR12 Report unmeasured accuracy honestly | 1.1–1.3, 2.1–2.3, 3.1–3.2, 4.1 | scripts/ai/research-comparison.test.mjs — keeps matched coarse, unknown and failed outcomes without inventing measured error; writes a reproducible JSON report from an explicit local input file |

## Execution and evidence

- Apply must inspect fresh GitNexus upstream impact before each existing symbol edit and run complete change analysis before committing. Resolve graph identity/truncation gaps with source and tests; no empty graph result proves safety.
- Run affected Go tests under the race detector, real SQLite fresh/upgrade/reopen/ownership tests, and frontend component and built-browser scenarios. Long-deadline tests use controlled time rather than ten-minute sleeps.
- Required completion checks use the installed shared runner against the actual change base: size/ratchet, dependency boundaries, formatting, lint/type generation/types, Go vet/race/build, frontend tests/coverage/build and browser smoke. Verify backend/frontend container builds for the changed application paths and the actual preview build.
- Preserve the adopted 80% AI Go statement and frontend line/branch coverage floors. Coverage alone does not replace the scenario cases.
- Built browser fixtures cover Research launch with a hint, multi-kilometer alternatives, safe links, keyboard/narrow layouts and offline maps alongside retained reload, account and manual/GPX journeys. Component/validator tests separately cover 500-meter and null-error values, missing/unsafe sources and enormous radii. External provider, Immich and map responses remain synthetic/local in ordinary tests.
- Explicit no-write/no-source-fetch counters verify that proposal generation and inspection create no Immich mutation or source-site request.
- Live integration is separate: verify the actual configured provider/proxy can return the new bounded answer and identify effective timeout limitations. No tool-event or search-log capture is required. Private-photo comparison needs the already authorized exact inputs or a new specific authorization; no ordinary test sends private images.
- For the live same-photo comparison, retain approximate results and failures. Report measured camera error only where independent reference coordinates and their uncertainty are known, separately from model-estimated radius. Do not turn either number into an acceptance gate.
- Planning-only verification consists of strict OpenSpec validation, requirement/scenario preservation, English/whitespace/link checks and artifact completeness. It does not establish application correctness or live compatibility.

Total mapped scenarios: 68.
