# CH16 verification plan

Status: original verification plan; executed scenario evidence and actual gate results are in [implementation verification](implementation-verification.md). No live calls are claimed. Apply uses OpenSpec Plus TDD for changed behavior; known behavior characterization may pass immediately under the adopted testing standard.

| Scenarios | Tasks | Required automated evidence |
|---|---|---|
| D01–D04 | 1.1–1.2 | Real SQLite accept/reaccept/reopen/immutable-result tests; real handler requests; built-browser accept/reload/reject flow; provider and Immich mutation counters stay zero. |
| D05–D07 | 2.1 | Pure draft validation plus protected HTTP tampering tests for camera/candidate/target/fields; component tests for coarse estimates, unknowns, zero and invalid input. |
| D08–D09 | 3.1–3.2, 1.2 | Barrier-controlled competing SQLite revisions, lost-response handler fixture, two-tab browser conflict with preserved unsaved values, stage/edit invalidation. |
| D10–D11 | 2.2 | Pure factual-revision transitions and component tests across candidate changes, scene-only text, each language, nullable direction and stale radius. |
| D12–D14 | 4.1–4.2 | Real read-adapter/fake-Immich integration, offline and denied access, exact baseline values, observation expiry, revision races and renewed source review; assert zero upstream mutations. |
| D15–D17 | 1.3, 4.2 | Real SQLite 025 upgrade/fresh/repeated migrations, reopen, account-qualified cascades, catalog reset, cleanup/accept race, installation invalidation and late-publication rejection. |
| D18–D20 | 2.1, 2.3, 3.2 | RTL and built Playwright keyboard/narrow-screen/tile-failure journeys, unsaved navigation, private response fencing and unchanged manual pending/GPX behavior. |

During apply, record exact test names and actual outcomes for every scenario in `implementation-verification.md`; grouped rows here do not replace that mapping. Include body bounds, safe errors, session/origin protection and history pagination regressions at the HTTP/SQLite layers. Use synthetic proposals and images, real temporary SQLite and deterministic local services; no private-photo or public-tile dependency.

Required closure checks: GitNexus impact before existing-symbol edits and complete pre-commit change analysis; strict OpenSpec validation; `bun run check --base <recorded-pre-change-commit>` for size/dependencies/format/lint/types, Go vet/race/coverage, frontend coverage, both builds and affected built-browser journeys. Respect 500-line limits and current AI coverage floors. Use the installed harness; no missing base infrastructure has been identified. New fixtures and journeys belong to their owning slices.

Verify migration rollback on failure and application rollback with draft tables retained. Update the maintained capability purpose at CH16 synchronization to include explicit local draft actions while preserving side-effect-free inspection and no Immich-write authority. Keep docs and roadmap status accurate. Live Immich writes are out of scope; retained model output from any supported mode suffices, so codex-proxy repair is not a prerequisite.
