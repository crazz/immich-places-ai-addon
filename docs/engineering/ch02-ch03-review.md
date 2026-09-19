# CH02 / CH03 completion review — 19 September 2026

## Scope

Reviewed `dbc2a52dfae82b111b107dff702eb94c6d67cb31..eac9f753eebec4c2e858ef05c76187059c41cc47` and the follow-up corrections in this document. The review covers provider egress and capability checks; it does not claim a fresh line-by-line review of all earlier foundation changes in PR #1.

CH02's five requirements were synchronized verbatim into the maintained provider specification and its completed change was archived. CH03's six requirements were already synchronized; archive-relative links and the current implementation status were reconciled. All task checkboxes are complete. Historical planning and verification sections remain historical.

## Confirmed defects and corrections

| Trigger | Before | Corrected behavior and regression evidence |
|---|---|---|
| Session revoked while DNS is pending | A capability probe could still transmit | `TestCapabilitySessionRevocationDuringDNSPreventsTransmission` verifies zero requests; dispatch rechecks authority after DNS |
| Profile edited during the final provider response | The old revision could publish applicable proof | `TestCapabilityEditDuringFinalResponseCannotPublishCurrentProof` verifies final authority revalidation |
| JSON followed by a closing delimiter | A malformed object could count as supported | `TestSyntheticSampleRejectsTrailingClosingDelimiter` requires complete decoder EOF |
| Uppercase values outside the requested enum | Values were repaired before comparison | `TestSyntheticJSONDoesNotRepairSchemaEnumValues` verifies exact schema values |
| Dispatch or direct adapter entry without a caller deadline | DNS could run outside the total timeout | Dispatch and adapter deadline tests cover bounded DNS and retained earlier deadlines |
| A public rule resolves to special-purpose addresses | Global-unicast classification admitted reserved, shared and transition ranges | Positive/negative address fixtures verify explicit exclusions against the IANA IPv4/IPv6 registries |
| IPv6, invalid ports or noncanonical numeric hosts in policy | Canonical identity could be invalid or ambiguous; ordinary dotted paths were rejected | `TestCanonicalDestinationRoundTripsAndRejectsAmbiguity` covers matching and round trips |
| HTTP uses port 443, HTTPS uses port 80, or Host is IPv6 | Host formatting could lose the approved authority | `TestHostHeaderPreservesApprovedAuthority` preserves the host and effective port |
| Stored proof was produced by the permissive protocol | Old results remained applicable after parser changes | Protocol v2 and `TestLegacyPermissiveProtocolCannotRemainApplicable` invalidate v1 without replay |
| Provider returns explicit numeric usage | Usage was discarded and consumption disclosure vanished on reload | Usage aggregation, malformed/absent metadata, migration from 019 and independent SQLite reopen tests cover persistence and unknown values |

Each defect was reproduced with a failing behavioral test before its production correction. Existing-behavior characterization tests were allowed to pass immediately. The authority formatter was extracted without behavioral changes and verified before its new failing cases. Refactor assessment retained focused provider, adapter and capability owners; no new dependency or architecture departure was needed. Migration 019 was preserved; migration 020 adds capability metadata without rewriting earlier evidence.

## Verification

The first full local check passed every non-browser gate; Chromium launch was blocked by the sandbox. The next full shared check, run outside that restriction, passed all thirteen gates and five browser journeys after the authorization, JSON, URL, address and timeout corrections. Numeric usage was corrected afterward and therefore requires the final check recorded below.

The usage follow-up full run passed every gate except Go tests: the migration bootstrap fixture still expected version 019, and `TestHandleUpdateLibrarySuccess` exposed asynchronous enrichment outliving its database cleanup. The fixture now expects 020, and the library handler test helper joins its existing service wait group before cleanup. No production library behavior changed. Both affected fixtures then passed a fixed 20-repeat race run. The full backend race/coverage gate passed afterward (96.123 seconds for the main package), with AI statement coverage **87.93% (969/1102)**. The successful frontend/build/browser gates from the preceding run remain valid because only backend test fixtures changed. All thirteen required gates are satisfied; 28 frontend tests and five browser journeys passed. Size and formatting checks were repeated successfully after the fixture edits.

Pinned tools: Node.js 22.23.2, Go 1.25.14 and Bun 1.4.2. Local attempt logs are retained under `out/checks/ch02-*` and `out/checks/ch03-*` (ignored execution output). The original HEAD also had green engineering CI and both Docker builds; those results do not substitute for verification of these follow-up edits.

## Review limits and live gate

GitNexus was bound to this checkout and refreshed with PDG data. The initial comparison identified 909 changed symbols and 81 affected flows, with aggregate critical risk. Upstream impact checks covered the changed production functions, methods and interfaces, and direct dependents were checked against source/tests. Shared canonicalization and HTTP transport reported HIGH risk before edits. Security explanation queries returned no findings for the 33 changed production files, but process enumeration, callbacks, property dispatch and cross-language links have documented gaps; this is not proof of absence of vulnerabilities.

Deterministic tests establish policy, ownership, transport bounds, persistence and UI behavior. They do not establish live NAS model availability, geolocation quality, access to private photographs, or authorization to mutate Immich. The operator's synthetic-only NAS procedure remains a separate gate. A final refreshed GitNexus change analysis was run before committing; aggregate risk remains critical for this broad foundational change, with no partial or truncated result flag. The graph limitations above still apply. Later analysis, selection, jobs and writeback remain planned.
