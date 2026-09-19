## 1. Owners receive a bounded explicit-image preview

- [x] 1.1 Deliver “Protect selection resources by current authority” through the protected preview/read API and normal frontend proxy, with automated coverage of every owner, session, Origin, disabled-AI and caching scenario.
- [x] 1.2 Deliver “Bound and deduplicate explicit input”, including exact count relationships, deterministic ID normalization/order and complete rejection of malformed, ambiguous or over-limit requests, with all specified input scenarios automated.
- [x] 1.3 Deliver “Explain eligibility without leaking inaccessible assets”, including deterministic exclusion precedence, zero-eligible previews and no implicit stack expansion, with pure policy and real-catalog isolation coverage for every scenario.

## 2. Frozen selections retain the user's actual catalog scope

- [x] 2.1 Deliver “Preserve the declared catalog scope” for all-catalog, album and recursive-folder views with tag/GPS/visibility/date filters, owner-safe scope validation and real SQLite coverage of every scoped, folder, date, zero-coordinate and invalid-input scenario; preserve CH04 and current catalog behavior.

## 3. Published snapshots survive restart without changing targets

- [x] 3.1 Deliver “Publish an exact atomic snapshot”, including immutable counts/provenance/digest, consistent concurrent resolution and rollback on failed publication, with real SQLite atomicity, cancellation and no-expansion scenarios.
- [x] 3.2 Deliver “Expire and revalidate retained membership”, including the protected read and internal later-consumer contract, stale whole-snapshot rejection, installation/policy binding and non-sliding expiry, with every scenario covered by real database reopen and controlled-clock tests.

## 4. Snapshot metadata has a bounded lifecycle

- [x] 4.1 Deliver “Bound snapshot retention and ownership cleanup”, including valid operator limits, owner/global quotas, periodic/startup cleanup while AI is disabled, sanitized cleanup failures and account-deletion isolation, with every lifecycle scenario automated.
- [x] 4.2 Make snapshot and minimal installation-identity persistence usable after fresh install or upgrade from 020, with additive migration, epoch/connection invalidation, foreign-key integrity and safe binary-rollback guidance verified through real migration/reopen fixtures; document the proposed limits and same-URL replacement procedure.

## 5. Selection remains isolated from external work and legacy workflows

- [x] 5.1 Deliver “Keep selection separate from analysis and writes” across production handler/proxy integration, with zero provider/image-fetch/write counters, retained manual/GPX/AI-disabled browser journeys and no incomplete launch UI; establish all installed size/dependency, format/lint/type, race/AI-coverage and production-build gates for the completed change.
- [x] 5.2 Make bounded selection behavior observable under a synthetic 100,000-asset catalog and 500 selected IDs, recording latency/memory and failure at configured limits without claiming NAS performance or upstream-permission verification; reconcile the operator guide and CH06 handoff with the implemented snapshot contract.
