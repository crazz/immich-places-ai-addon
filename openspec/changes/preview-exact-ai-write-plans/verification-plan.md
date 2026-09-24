# CH18 verification plan

Status: planned, not executed. CH16 must first be implemented, verified and synchronized. Use its actual public draft contract during apply; reconcile any prerequisite drift before coding.

| Scenarios | Tasks | Required automated evidence |
|---|---|---|
| P01–P03, P06 | 1.1 | Workflow and real HTTP-adapter tests for exact one-photo GPS scope, missing/partial/zero metadata, malformed input, authority loss and bounded failure; real protected handler plus zero provider/mutation counters. |
| P04–P05 | 2.1 | Baseline drift and material-image identity fixtures; metadata-only timestamp changes do not masquerade as image replacement; browser conflict and explicit renewed-review path. |
| P07–P09 | 2.2, 3.1 | Real SQLite plan/digest reopen, independent-owner fixtures and barrier-controlled draft/credential/installation changes during read; canonical bytes/digest and tampered scope tests. |
| P10, P14–P16 | 1.2, 3.3 | RTL and built-browser numeric diff, unchanged pair, keyboard/narrow-screen/tile failure, same-asset manual pending conflict and private late-response fencing. |
| P11–P13 | 3.2 | Injected-clock expiry, stale revision, bounded capacity/cleanup and preservation of referenced drafts/history on real SQLite. |
| P17–P18 | 2.2, 3.3 | Disabled-execution reads, account deletion/late publication, installation rotation and foreign/nonexistent error equivalence through real auth and persistence. |

Apply records exact scenario-to-test names and executed results in `implementation-verification.md`. Include fresh/CH16-upgrade/repeated migration runs, foreign-key behavior on pooled connections, storage failure before publication, safe errors, no redirects, oversized responses and zero hidden mutation paths. Characterize inherited manual behavior without manufacturing test failures. New behavior follows OpenSpec Plus TDD.

Use existing synthetic HTTP services, real temporary SQLite, Vitest/RTL and the built-application Playwright harness. No missing base harness is known; add preview-specific fixtures within the owning slices. No private photo, live provider or live Immich mutation is needed.

Closure: strict OpenSpec validation; GitNexus impact before existing-symbol edits and complete pre-commit change analysis; full `bun run check --base <recorded-pre-change-commit>` including size/import rules, format/lint/types, Go vet/race/AI coverage, frontend AI coverage, builds and affected browser journeys. The apply record must name actual pass/fail/unavailable results. Back up before rollout and verify application rollback retains additive preview tables. Document that passing preview tests does not close live write compatibility GATE-02.
