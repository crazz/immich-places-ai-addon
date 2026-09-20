# CH07 analysis-result validation verification

Implemented on top of CH06 commit `9e3866544ae8a0820e9e1f004248231be4d101cb` on 20 September 2026. Work used sequential inline OpenSpec Plus apply/TDD and self-review, without subagents. The owner authorized completing, archiving and committing each change before the next.

The focused `backend/internal/ai/results` package accepts complete untrusted bytes plus separate server context and returns an opaque proposal or bounded findings. It has no provider, image, SQL, environment or writer dependency. The [caller contract](../ai-result-validation.md) records exact budgets, semantic policy and the CH09/CH11 handoff. This change adds no endpoint, migration, queue or production dispatch.

## Scenario evidence

The eight requirements contain 27 scenarios. All are covered by the following synthetic fixtures; test names without a path belong to `backend/internal/ai/results`.

| Capability scenario | Automated evidence |
|---|---|
| Canonical fixtures remain valid | `TestCanonicalFixturesRemainUnchanged`: located and unknown examples, unchanged values and explicit matching context |
| Structurally invalid or ambiguous output fails closed | `TestAmbiguousOrMalformedJSONCannotBecomeAProposal`, `TestClosedSchemaFieldsAreRequiredAndTyped`, `TestUnsupportedSchemaVersionHasItsOwnFailure`, `TestJSONUnicodeIsNeverRepaired`: required nullable fields, extras, types, duplicate keys, trailing/truncated/fenced input, UTF-8 and surrogate integrity |
| Numeric and resource bounds are enforced | `TestParseBoundsBytesDepthNodesAndStrings`, `TestNumbersAreBoundedWithoutLosingExactRanges`, `TestCollectionsAndIdentifiersHaveIndividualBounds`, `TestSerializedProposalCannotExceedThePayloadBudget`: exact and one-over budgets including serialization expansion; constructor and root HTTP-spy tests deny external resource access |
| Transport failure cannot masquerade as complete | `TestTransportCompletionIsAuthoritative`: refusal, truncation, tool response, absent and unknown states reject otherwise valid bytes |
| Located output selects a camera candidate | Canonical fixture plus `TestTypedCameraAndSubjectCoordinatesStaySeparate`: exact selected camera retained without approval |
| Ambiguous and unknown remain unselected | `TestUnknownSubjectAndAmbiguousAlternativesRemainUnselected`: alternatives retain order, no automatic selection or zero sentinel |
| Inconsistent outcome selection is rejected | `TestContradictoryOutcomesAreRejectedWithoutRepair`: missing camera/selection, dangling selection, one ambiguous candidate and forbidden selections |
| Local references resolve and preserve kinds | `TestSourcesMustBelongToThisAuthorizedBundle`, `TestProvidedContextRetainsAnAuthorizedSourceConnection`: exact local and authorized source membership |
| Dangling and duplicate identities are rejected | `TestInvalidIdentitiesAndLocalReferencesCannotBeInvented`, `TestCountryCodesAreAssignedISOAlpha2Only`: IDs, required prose, references and country membership |
| Foreign or invented sources confer no authority | `TestSourcesMustBelongToThisAuthorizedBundle`: foreign, invented and URL-shaped claims fail against the separate request bundle |
| Visual output cannot claim supplied context | Source and context-provenance fixtures reject provided-context observations and unsupported source references |
| Context observations retain a source connection | `TestProvidedContextRetainsAnAuthorizedSourceConnection`: empty bundle and missing candidate source connection fail |
| Landmark does not become photographer position | `TestUnknownSubjectAndAmbiguousAlternativesRemainUnselected`: subject-only unknown stays separate with null camera/direction |
| Boundary and zero coordinates remain meaningful | `TestCoordinateBoundariesAreExactForBothPairs`, typed-coordinate and number tests: both pairs, zero, exact ±90/±180 and precise beyond-boundary decimals |
| Supported estimate or unknown radius is retained | `TestSupportedRadiusStatesAreRetained`: visual area 250 m, point zero, null/unknown and supported context/source bases |
| Contradictory or misleading precision is rejected | `TestRadiusClaimsNeedConsistentGranularityAndBasis`: null/basis mismatch, non-point zero, numeric city/region |
| Radius provenance requires matching evidence | Both radius fixtures remove each required visual/source support connection and require rejection |
| Nullable heading remains valid | `TestHeadingBoundsAndSupportedMethodsRemainExact`: null direction remains absent on a valid camera proposal |
| Supported heading preserves method and uncertainty | Same heading fixture: zero/sub-360, null/bounded uncertainty and referenced alignment evidence |
| Unsupported or out-of-range heading is rejected | `TestDirectionNeedsAViewpointAndMethodSpecificEvidence` and heading-bound fixtures: no viewpoint, missing evidence, subject bearing alone and exact angular violations |
| Non-English and unavailable descriptions are valid | `TestNonEnglishStatusesAndNormalizedTagsArePreserved`: Ukrainian/Portuguese, explicit unavailable status, preserved prose and no English fallback |
| Tags normalize without duplicate coverage | `TestLanguageAliasesCannotDuplicateOutputCoverage`, normalized-tag fixture: aliases/case normalize and duplicate coverage rejects |
| Incomplete or contradictory descriptions reject | `TestDescriptionsCoverOnlyTheRequestedLanguagesAndStatuses`: missing/extra/invalid language, text/reason and candidate/scene basis |
| Invalid server language context fails separately | `TestInvalidServerContextFailsBeforeProviderParsing`, `TestServerContextBoundsAndCopiesAuthorizedDescriptors`, `TestNormalizedLanguageTagsRespectTheByteCeiling`: requested set, primary membership and raw/normalized limits |
| Model envelope injection is rejected | `TestModelCannotInjectEnvelopeOrLeakUnknownKeys`: identity, approval, verification and write-plan extras fail without echoing keys |
| Failures are bounded and redact payloads | `TestFailuresAreBoundedSortedAndNeverEchoPayloads`: deterministic codes/paths, ≤20 findings and ≤4096 JSON bytes, no private prose/IDs/coordinates |
| Validated data is independent and side-effect free | `TestValidatedProposalOwnsAllReturnedData`, context-copy and zero-value fixtures; root `TestCompiledValidatorIsConcurrentAndNeverLoadsModelURLs` shares one compiler across 24 concurrent calls with zero HTTP requests; source-based dependency fixtures reject direct/transitive writer reachability |

The boundary test uses a real loopback HTTP spy in the root adapter test layer. Pure-package tests compile rejecting HTTPS/file-reference schemas, use embedded fixture bytes and exercise result ownership. Structural dependency checks establish absence of image, persistence and writer reachability; the package has no such clients or callbacks to invoke. Existing browser provider/image/write counters remain unchanged. These checks do not establish live provider compatibility, source truth, current tenant authorization or location/direction accuracy.

## Verification and measurements

`bun run check --base 9e3866544ae8a0820e9e1f004248231be4d101cb` completed with exit 0 and **all 13 gates passing in one run**: checker regression, size, dependencies, formatting, lint, generated route types, TypeScript, Go vet, race tests, Go build, frontend tests, frontend production build and smoke. Combined AI Go statement coverage was **89.67% (1727/1926)**, above the installed 80% gate. All seven browser journeys passed. Three inherited lint warnings remain. No frontend implementation changed in CH07; repository-wide frontend coverage is not an AI coverage figure.

The focused fuzz run used `go test -run '^$' -fuzz '^FuzzValidateBoundedResult$' -fuzztime=30s -parallel=2 ./internal/ai/results`: **245177 executions**, 13 additional coverage-interesting inputs, 31.388 seconds total, no failure corpus. Properties check bounded typed failures or serializable, stable revalidated proposals. The final shared race run also executed the seed corpus after the normalized-language and serialization-bound fixes.

`TestRepresentativeAndMaximumBoundValidation` measured the 1903-byte representative fixture at **165.875 µs / 48736 allocated bytes** and the full 1 MiB workload at **12.91125 ms / 9969472 allocated bytes** on this development host. The latter reaches all collection ceilings, including 100 observations/sources, 20 candidates, 10 descriptions, per-candidate reference/note limits and 50 warnings. Allocations are total churn, not retained peak heap; these synthetic measurements are not NAS acceptance or a service-level guarantee.

The ordinary backend-context Docker build passed as `immich-places-ai:ch07-validation`. A separate package-test binary built in the same backend context passed canonical-fixture, rejecting-loader, normalized-language and serialization-bound tests under Alpine with `--network none --read-only`, `/tmp` as tmpfs and no source/docs tree. This proves embedded runtime schema/fixture packaging without adding a public route or an unused startup invocation to force linkage before CH09.

The canonical schema moved byte-for-byte into the package: SHA-256 `a7c1ebaad9e0d92ac3ed556dd2913db1424b7c023edd48a636c47a6006bd4baa`. Both embedded test fixtures also match the original document examples byte-for-byte. There is one canonical schema authority. Current links and semantic guidance point to it; original Word exports remain historical.

Review found and fixed a normalization-budget gap: a valid 128-byte `sh` private-use language tag expands to 133 bytes under canonicalization. A failing regression now requires both raw and normalized tags to fit 128 bytes, while an exact-bound `en` tag remains valid. The dependency regression also proves `results → jobs → writeback` is forbidden. Detailed RED/GREEN notes and attempt logs are retained locally under ignored `out/checks/ch07/`.

Final review also reproduced serialization growth: valid literal HTML characters in bounded prose expand when JSON escapes them. Publication now checks the same 1 MiB ceiling after serialization and returns no proposal on overflow. The regression started from the valid maximum workload and failed before the guard; accepted proposals remain serializable within their validation byte budget.

## Dependency audit

`go mod tidy` and `go mod verify` passed. The only added direct library is owner-approved `github.com/santhosh-tekuri/jsonschema/v6 v6.0.3`; existing `golang.org/x/text v0.32.0` becomes direct without a version change. The two regexp2 checksum entries belong to the schema library's test dependency. Go remains 1.25.

Both `go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 -show verbose ./internal/ai/results` and the whole-backend `./...` scan completed successfully. v1.7.0 was the newest inspected release compatible with Go 1.25; v1.8.0 required Go 1.26. The scans reported **zero called vulnerabilities and zero vulnerabilities in imported packages**. This is reachability evidence, not a guarantee of security.

The scoped result-package scan reported one unused-module finding in existing x/text: [GO-2026-5970](https://pkg.go.dev/vuln/GO-2026-5970), concerning `unicode/norm` iteration and fixed in v0.39.0. The result package uses `language`, not the affected package. No finding was reported for the added schema-validator version.

The whole-backend scan reported **20 findings in unused portions of existing required modules**:

| Existing module | Findings | Reported fixed version |
|---|---|---|
| `golang.org/x/text v0.32.0` | GO-2026-5970 | v0.39.0 |
| `golang.org/x/sys v0.39.0` | GO-2026-5024 | v0.44.0 |
| `go.mongodb.org/mongo-driver v1.11.4` | GO-2026-5327 | v1.17.7 |
| `golang.org/x/crypto v0.46.0` | GO-2026-6355, GO-2026-6354 | v0.56.0 |
| Same x/crypto | GO-2026-6303 | v0.55.0 |
| Same x/crypto | GO-2026-5932 | Not available in this scan |
| Same x/crypto | GO-2026-5033, GO-2026-5023, GO-2026-5021, GO-2026-5020, GO-2026-5019, GO-2026-5018, GO-2026-5017, GO-2026-5016, GO-2026-5015, GO-2026-5014, GO-2026-5013, GO-2026-5006, GO-2026-5005 | v0.52.0 |

These findings are recorded for dependency maintenance; this bounded change does not upgrade unrelated modules or claim that the dependency tree has no advisories. Saved verbose logs distinguish module presence from imported-package/call reachability.

## Inline review and completion

Spec and code-quality self-review covered all 27 scenarios, exact numeric preservation, parser work limits, source authority, copying, nullable geometry, deterministic diagnostics and downstream boundaries. No architectural departure or new persistence claim was introduced. GitNexus was rebuilt for this checkout with PDG retained; source and tests supplement its unresolved receiver/process edges and inconsistent search-name associations.

After synchronization/archive, complete GitNexus change analysis reported **108 changed symbols in 51 files and three affected processes, medium risk**, with matching returned counts and no partial/truncated flags. One process is the dependency checker. The other two associate the new `Source.ID` property with selection preview and capability admission: direct import/caller inspection confirms the new result package has only test consumers, while those production paths are unchanged. The source-based dependency gate and their passing regression suites supplement those imprecise graph associations. Local index metadata matches the CH06 base despite the MCP's cached two-commit staleness hint; the final schema, serialization guard and test symbols are present in the rebuilt graph.

The final cumulative run after the serialization fix passed all 13 gates with exit 0. OpenSpec apply reported **9/9 tasks, all_done** before archival. The maintained proposal spec contains eight requirements and 27 scenarios; strict validation passed all four maintained capabilities, and `openspec list` reports no active changes. All 97 local links in changed Markdown files, canonical/fixture byte comparisons and whitespace checks passed. Inline final review has no unresolved correctness finding. Detailed local logs remain under `out/checks/ch07/`.
