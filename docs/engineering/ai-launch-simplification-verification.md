# AI launch correction verification

20 September 2026. This correction follows [ADR-08](decisions/ADR-08-simple-ai-launch.md): repair responsive launch layout, remove the redundant image checkbox, and use application defaults instead of requiring manual token-policy setup.

## Behavioral evidence

| Contract | Automated evidence |
|---|---|
| Desktop and phone fields fit their dialog/columns | `tests/e2e/providers.spec.ts` geometry scenario at 1440 and 390 pixels |
| Start is the explicit authorization of current settings; editing does not dispatch | `LaunchForm.test.tsx` one-click Visual scenario |
| Context classes remain selected independently; uncertain submission retains its exact key | `LaunchForm.test.tsx` reconciliation scenario and `tests/e2e/workflows.spec.ts` |
| Advanced controls are optional and provider changes refresh supported format and allowances | `LaunchForm.test.tsx` default-settings scenario |
| Readiness, admission, worker and immutable result work without configured policies | `TestAIProductionRunsWithApplicationDefaultsWithoutOperatorPolicy`, using real SQLite and synthetic image/provider services |
| Actual usage above requested defaults is retained without a false attestation violation | `TestAIProductionDefaultsRetainUsageAboveRequestedTokensWithoutBlocking` |
| Explicit operator violations still block | Existing `TestAIProductionOverAllowanceInvalidatesPolicyEvenForInvalidOutput` |
| Default identity remains exact; explicit restrictions cannot silently become defaults | `TestExecutionDefaultsKeepExactIdentityAndExplicitRestrictions` and existing readiness/revision tests |
| Malformed DTOs remain rejected; application defaults are accepted | `executionReadiness.test.ts` |
| Disabled, untested and explicitly restricted providers have actionable errors | `launchAdmission.test.ts` |

The layout regression first failed with 308 pixels of provider overflow in a 448-pixel dialog, then passed after the AI dialog used its existing wider size and constrained field columns. The one-click scenario first failed because the checkbox remained present. Default production launch first failed with `policy_required`; the above-estimate usage scenario first failed because it blocked the job. DTO and optional-controls scenarios also failed before their implementation. Each cycle was completed separately, including a refactor assessment. The new binding characterization passed immediately against the already implemented resolver, consistent with the project's characterization rule. Detailed local cycle evidence is retained under ignored `out/checks/launch-simplification/`.

## Scope and review

No dependency, schema migration, public route or Immich mutation was introduced. Default policy identity includes its version and exact authority binding. The existing admission and reconciliation formats remain unchanged; Start captures the configuration in that record. Context classes, source checks, calls, attempts, payload limits, deadlines and explicit operator restrictions remain enforced.

GitNexus was refreshed against the current checkout before change analysis. The new resolver/defaults and existing admission, worker accounting, readiness parsing and launch functions were included. The report identifies critical impact because these are shared execution and DTO boundaries. Source inspection and the full regression harness cover those paths, including legacy manual/GPX journeys. The graph's process enumeration reports bounded/cross-language gaps; it is not proof of complete runtime coverage or a substitute for tests. No subagents were used.

The deployed proxy adapter was inspected read-only and does not forward maximum-token fields. Consequently default reservations and output-token requests are explicitly estimates, not claimed enforcement or billing guarantees. Ordinary tests use synthetic data. No private photograph was submitted to a provider during this correction.

## Completed checks

- `bun run check --base 3fa8510`: 12 gates passed, including race-enabled Go tests, dependency/size checks, lint/types and production builds. Backend AI statement coverage: 88.01%. Frontend: 88 tests in 44 files; AI lines 94.01%, branches 85.29%. Three documented inherited lint warnings remain.
- The first smoke invocation could not start Chromium inside the macOS sandbox (`MachPortRendezvousServer: Permission denied`). `bun run check --base 3fa8510 --gate smoke` then ran outside that sandbox: 3 legacy and 11 AI browser journeys passed. No application tests were skipped to resolve the environment failure. All 13 gates have passing evidence.
- `openspec validate --all --strict`: all 6 maintained specifications passed.
- GitNexus `detect_changes(scope: all)` included 73 changed indexed symbols and 32 affected processes after refreshing the index; no partial/truncated change-report flag was returned. The cycle check returned `clean`, `enumeration: complete`, zero cycles. The process-index limits described above still apply.
- Both Dockerfiles built successfully for `linux/amd64`. Desktop and phone regression screenshots were inspected locally; controls stay in their columns, with vertical scrolling when needed.

Live rollout uses the existing isolated preview stack and preserves its database and encryption key. The deployment helper backs up SQLite before changing only the preview image revision, then verifies health and unchanged production/proxy container identities. A successful synthetic run is not evidence of live model quality or provider token enforcement.
