# CH15 proposal review verification

20 September 2026; base `ea91bd3`; sequential inline OpenSpec Plus apply, TDD and self-review. No subagent or independent reviewer is claimed. See the [user contract](../ai-proposal-review.md) and [change](../../openspec/changes/archive/2026-09-20-review-ai-camera-and-subject-proposals/proposal.md).

## Scenario evidence

| Canonical scenario | Executed evidence |
|---|---|
| Show different camera and subject points | `reviewParser.test.ts`, `CandidateFacts.test.tsx` and the located browser journey preserve distinct camera/subject coordinates. |
| Inspect an ambiguous alternative | `ProposalReview.test.tsx` and the ambiguous browser journey change only inspection focus and compare the original API payload after navigation. |
| Preserve missing and zero coordinates | `reviewParser.test.ts`, `reviewValidation.test.ts` and `CandidateFacts.test.tsx` cover zero coordinates and subject-only unknowns. |
| Show a supplied radius honestly | `CandidateFacts.test.tsx`, `reviewMapLayers.test.ts` and the located browser journey show units, basis and the uncalibrated estimate. |
| Show unknown or zero radius | `CandidateFacts.test.tsx` preserves zero and unknown precision; `reviewMapLayers.test.ts` omits null-radius circles. |
| Inspect projection edge cases | `reviewGeometry.test.ts` retains canonical polar/antimeridian coordinates while choosing a short span or omitting unsupported projected points. |
| Inspect supplied true-north direction | `reviewParser.test.ts`, `CandidateFacts.test.tsx` and real Leaflet layer tests preserve zero true-north heading, method and uncertainty. |
| Keep unknown direction unknown | `CandidateFacts.test.tsx` and `reviewMapLayers.test.ts` leave missing heading unknown and omit unsupported arrows. |
| Inspect source-free Visual evidence | `EvidencePanel.test.tsx` and the located browser journey label source-free model evidence without links or verification claims. |
| Inspect contextual lineage | `reviewParser.test.ts` resolves Context sources separately; `EvidencePanel.test.tsx` distinguishes hints, neighbors and unknown lineage. |
| Handle unsafe text or unresolved references | `reviewParser.test.ts` rejects unresolved references; `EvidencePanel.test.tsx`, `reviewMapLayers.test.ts` and `ResultDetail.test.tsx` render markup as inert text. |
| Switch between complete and unavailable languages | `DescriptionTabs.test.tsx` and the located browser journey cover complete/unavailable states, exact reason, primary label and keyboard switching without provider calls. |
| Inspect another candidate without rewriting text | `DescriptionTabs.test.tsx`, `ProposalReview.test.tsx` and the ambiguous browser journey preserve the original candidate/scene basis. |
| Preserve a non-English language set | `DescriptionTabs.test.tsx` preserves the requested French/Portuguese set after a document-locale change without adding English. |
| Review without map tiles or an image | `ReviewMap.test.tsx` and the located browser journey exercise tile failure/retry and a denied thumbnail while retaining numeric facts, evidence and keyboard tabs. |
| Attempt manual-style map interactions during review | `tests/e2e/review.spec.ts` clicks, pans, drops and opens the context menu with an existing manual preview; review makes zero writes, then manual confirmation writes exactly the original coordinates. |
| Leave or switch private review | `ResultDetailReview.test.tsx`, `useResultRead.test.ts`, `ReviewMap.test.tsx` and browser back/reopen/close cover private cleanup, aborts, canonical focus reset and restored focus; manual/GPX smoke journeys pass. |

## Verification results

- All thirteen shared gates passed: checker tests, size/dependencies, Go formatting, lint, route types/TypeScript, Go vet, race/coverage, both application builds, frontend coverage and browser smoke. Existing lint has three unrelated warnings.
- All **85 frontend tests** passed. AI TypeScript coverage: **93.62% lines (792/846)** and **84.78% branches (1092/1288)**; both exceed the 80% floor.
- Backend race/coverage regression passed. AI Go coverage remains **87.96% (4056/4611)**; no backend application code changed.
- All **13 browser journeys** passed: 3 auth/manual/GPX and 10 AI journeys. New tests use canonical synthetic located/ambiguous provider replies through the real durable workflow, denied tiles/images, keyboard alternatives/languages, immutable response comparison and exact manual write assertions. Desktop and narrow-screen screenshots were visually reviewed.
- Docker image `immich-places-ch15-check` built successfully; it was not deployed or published.
- GitNexus full-index analysis reported **202 changed symbols, 18 affected flows and 56 files**, aggregate risk **critical**, without partial/truncated output or invalid symbol IDs. Shared dialog/base-map consumers received the required regression checks. Import-cycle enumeration completed with zero cycles. Anchored detail/layer taint queries returned no findings; callback/property/implicit flows remain analysis limits, so inert-text tests, source review and browser no-write assertions provide separate evidence.
- Strict OpenSpec validation passed all six maintained capabilities. The synchronized review capability retains CH14 and CH15: 11 requirements and 32 scenarios. All 141 checked local links in changed documents resolve.

Browser verification caught a real entry-point overlap: legacy map controls intercepted clicks on AI Results. The button now sits above those controls, and the same unforced pointer journey passes with pending manual edits. Early test runs also exposed incorrect UI selectors and narrow fixture/RTL option types; these were corrected and checks rerun. Detailed TDD, review and gate logs remain ignored under `out/checks/ch15-apply/`.

## Limits

All fixtures are local and synthetic. No live provider, private photograph, external map/geocoder, production Immich mutation or NAS measurement was used. Tests establish presentation, isolation and lifecycle behavior, not geographic accuracy or live model quality. Large/unknown direction uncertainty uses numeric text without a precise arrow. Projection limits leave original numeric data available. Draft editing, translation regeneration and confirmed writing remain later changes.
