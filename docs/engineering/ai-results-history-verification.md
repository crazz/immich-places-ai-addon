# CH14 history verification

20 September 2026; base `187b45d`; sequential inline OpenSpec Plus apply and self-review. No independent reviewer or subagent is claimed. [User contract](../ai-results-history.md) and [change](../../openspec/changes/archive/2026-09-20-browse-persistent-ai-results/proposal.md).

## Scenario evidence

| Canonical scenario | Executed evidence |
|---|---|
| Retain results after GPS or catalog changes | `aiResultRead_test.go` updates GPS, removes the source, reopens SQLite and retains history; `aiResultContext_test.go` preserves Context detail after source deletion. |
| Read with execution disabled | Real SQLite/API reads and the built Visual/Context browser journey with execution disabled. |
| Reject foreign history access | `aiResultIsolation_test.go`, protected HTTP tests and current-image foreign-owner rejection. |
| Page while new work completes | `aiResultPagination_test.go`: tied/backward later completions excluded by the watermark, exact continuation and refresh. |
| Filter retained album and date facts | `aiResultFilters_test.go`, migration backfill tests and unknown-date form contract; source-local day, no current-catalog fallback. |
| Reject invalid query boundaries | Pure query table, protected API and encrypted cursor mismatch tests. |
| Display successful unknown separately from failure | Successful, permanent-failure and canceled SQLite details; `ResultList.test.tsx` and `ResultDetail.test.tsx`. |
| Preserve older runs after reanalysis | `aiResultRuns_test.go` retains original payload/model/revision; actual Visual/Context rerun browser history. |
| Reject corrupt or unsupported detail safely | `aiResultCorruption_test.go`: corrupt payload, schema/validation version, mode, language and installation; unrelated list stays readable. |
| Display an authorized current thumbnail | `aiResultImage_test.go` traverses current authority and real raster preparation; built-browser current images. |
| Lose source access | `aiResultImageAuthority_test.go`: hidden/removed/library/foreign/upstream denied, key/installation changed in flight; UI unavailable placeholder. |
| Change accounts during image loading | `ResultImage.test.tsx` and `useResultRead.test.ts`: abort, late-response suppression, private URL cleanup. |
| Return from detail to the same list | `AIResultsWorkspace.test.tsx`: query/cursor/scroll/focus and reload; `result-history.ts` verifies real browser keyboard navigation. |
| Handle unavailable catalog or history requests | `ResultsShell.test.tsx` without an Immich key/catalog readiness; list explicit retry/filtered-empty and stale read-hook tests. |
| Render untrusted content without write authority | Escaped script-like labels/observations, read-only clients, browser provider-request/write assertions and legacy manual/GPX journeys. |

Additional real SQLite checks cover migration 024→025 with corrupt legacy descriptions, atomic rollback when projection publication fails, and 10,000 terminal entries traversed exactly once through 100 bounded pages. `EXPLAIN QUERY PLAN` uses the terminal-history page index and point joins without scanning full analysis payloads. The local non-race scale test took 1.74 seconds; this is not a reference-NAS benchmark.

## Verification results

- All Go packages passed the race-enabled coverage gate. AI statement coverage: **87.96% (4056/4611)**, above the 80% minimum.
- Go formatting, vet/build, checker tests, size/dependency rules, route types, TypeScript, lint and frontend build passed. Lint retains three pre-existing warnings outside this change.
- All 71 frontend component/unit tests passed with coverage.
- All 11 browser journeys passed: 3 auth/manual/GPX regressions and 8 AI journeys. The AI suite was rerun after fixing the exact-label selector and ensuring execution state resets after each workflow test. The detail screenshot was visually reviewed.
- GitNexus full-index pre-commit analysis reported 243 changed symbols and 47 affected flows across 77 files, aggregate risk critical, with no partial/truncated result or invalid symbol IDs. Import-cycle enumeration was complete with zero cycles. Anchored history/image taint queries returned no findings; this is not proof of safety. Go route/shape extraction returned no routes; real API/browser checks cover that static-analysis gap.
- Docker image `immich-places-ch14-check` built successfully without deployment or publication.

Initial checks caught frontend wire/control mismatches, a test mock changed into an async component by autofix, and strict-label browser selection. These were corrected and the affected checks rerun; no failing check is treated as a pass. Detailed local TDD/check outputs are ignored under `out/checks/ch14-apply/`.

## Limits

Only synthetic local fixtures were used. No live provider/private-photo dispatch, NAS benchmark, production deployment or Immich write was performed by this change. Geographic accuracy and live full-contract model quality remain unverified. No draft acceptance, confirmed writer, retention schedule or deletion policy is introduced. GitNexus findings complement source/behavioral checks; framework dispatch, cross-language fields and sampled execution flows remain static-analysis limits.
