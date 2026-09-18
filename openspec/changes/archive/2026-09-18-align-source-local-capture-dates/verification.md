# Capture-date verification

Implemented and reviewed inline on 18 September 2026 against `5a35077`. No subagents, live services, private photos or manufactured characterization failures were used.

## Scenario traceability

| Spec scenario | Automated evidence |
|---|---|
| Offset timestamps cross UTC midnight | `TestCaptureDayCountsPreserveOffsetCalendarDate`, `TestCatalogDayCountsReturnSourceLocalDays` |
| Offsetless and daylight-saving capture times | `TestCaptureDateRepresentationsAndOrdering` |
| Inclusive end and one-sided bounds | `TestCaptureDateRepresentationsAndOrdering`, `TestCaptureDatesAcrossCatalogConsumers` |
| Missing and invalid calendar dates | `TestCaptureDatesWithoutCalendarDay` |
| Scoped assets and counts agree | `TestCaptureDateScopesRemainIsolated` |
| Folder, album and marker consumers retain date boundaries | `TestCaptureDatesAcrossCatalogConsumers`, `TestCaptureDateScopesRemainIsolated` |
| Gallery sort remains independent | `TestCaptureDateRepresentationsAndOrdering` |
| Invalid ranges across catalog endpoints | `TestCatalogDateRangesRejectReversedBounds`, `TestCatalogDateRangeHTTPContracts` |
| Valid and missing required bounds | `TestCatalogDateRangeHTTPContracts` |
| Authentication is preserved | `TestCatalogDateRangeHTTPContracts` |

New-behavior tests first demonstrated UTC-shifted day buckets, invalid-date leakage in bounded assets/albums/markers, and HTTP 200 for reversed ranges at all seven catalog endpoints. Minimal shared-policy/validation changes made them pass. Existing raw metadata, leap-day/offsetless/DST formats, sorting, ownership/filtering and HTTP contracts were characterized directly. Refactor assessment retained the small shared helpers; no further abstraction was needed.

## Reviews and checks

Spec-compliance review preceded code-quality review, followed by cumulative integration review, all performed here. No critical or important findings remained. Production scope is one new 27-line policy file and the existing asset, day-count, album, marker and range-validation consumers. The inherited database/handler ceilings were reduced to 982/926 lines. No schema, dependency, frontend implementation or Docker context changed.

Focused real SQLite and HTTP tests pass. `bun run check --base 5a35077` exited zero with all thirteen gates passing: 66 checker regressions, size/dependency/format/lint/type checks, Go vet and race suite, seven frontend tests, both production builds and three browser journeys. The race suite completed in 77.965 seconds and browser journeys in 7.3 seconds. The three inherited lint warnings remain visible; no new warning was suppressed.

OpenSpec change and maintained-spec strict validation pass. The maintained spec was compared with every delta requirement and scenario. Whitespace and local document links were checked.

GitNexus was refreshed with current source before review. Its complete structured change analysis reports 81 changed symbols, 19 files and 12 affected processes, with high risk and no partial/truncated flags. Shared catalog and marker relationships were checked against source and exercised by scoped regressions and the full suite. The indirect auth and Dawarich paths were confirmed in source: registration and refreshed user responses count markers without date bounds. The shared helper returns no clause for those unbounded calls; the complete backend suite and browser auth journey cover existing behavior. The complete cycle enumeration reports zero cycles. The unavailable MCP connection and CLI display cap were bypassed using the same installed local backend for the full structured result, not treated as clean empty results.

Detailed local run logs and inline review notes are under `/tmp/immich-ai-validation-20260918/ch04-*`; they are transient evidence, not repository artifacts. Container verification is not applicable because this change adds no package/build-context change. Remote CI and live compatibility were not run.

## Limits

These tests prove the existing catalog query contract, not a new undated-group UI, frozen AI selections, live Immich compatibility or reference-NAS performance. No timestamp is migrated or assigned an inferred timezone. Later FR-01 work remains visible in the roadmap. Backend AI code is absent and its coverage remains unmeasured.
