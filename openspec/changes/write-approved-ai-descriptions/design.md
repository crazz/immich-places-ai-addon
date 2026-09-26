## Context

See [proposal](proposal.md). At `0c5bc0c`, `writepreview.Plan` and `writeback.DecodePlan` accept only `gps-preview-v1`, one target, one GPS field and 16 KiB of canonical plan bytes. Migration 028 stores an immutable approval and one target state; `immichwrite.Transport.Send` serializes only a GPS pair. Draft descriptions are local and bounded to 16 KiB. Source baselines currently parse only GPS. These are deliberate CH19 boundaries, not generic field support.

GitNexus identified `Build`, `aiWriteStore.confirm`, `approvalPlan`, `aiWriteAttempt.Send`, the draft editor and operation view. Source and handler checks confirmed the unresolved receiver call from `aiWriteHTTP.go` to `store.confirm`; graph impact remains a lower bound. Preserve the recovery regressions recorded in CH19. Follow the adopted [architecture](../../../docs/engineering/architecture.md), [testing](../../../docs/engineering/testing.md), [coding standards](../../../docs/engineering/coding-standards.md), ADR-07, ADR-09 and REC-01/REC-05.

The pinned [Immich v3.2.2 controller](https://raw.githubusercontent.com/immich-app/immich/v3.2.2/server/src/controllers/asset.controller.ts), [DTO](https://raw.githubusercontent.com/immich-app/immich/v3.2.2/server/src/dtos/asset.dto.ts) and [service](https://raw.githubusercontent.com/immich-app/immich/v3.2.2/server/src/services/asset.service.ts) support single-asset PATCH with optional description and paired GPS. PATCH is excluded from generated OpenAPI; absence there is not absence of the route. This source inspection is not live write evidence.

## Goals / Non-Goals

**Goals:** extend the existing confirmed writer with exact standard-field selection, safe owned append blocks, source-consistent readback and durable field outcomes while preserving old GPS approvals byte-for-byte.

**Non-Goals:** a generic arbitrary-field writer, implicit description selection, schema edits to applied migrations, a second mutation runtime or a promise of remote compare-and-swap/sidecar completion.

## Decisions

### D1. Use an explicit versioned plan extension

Keep the v1 canonical decoder and recovery path for retained CH19 plans. New standard-field plans use a separately tagged v2 envelope with one target, selected GPS and/or description, exact per-field before/intended values, text comparison policy, language, description policy and append lineage. Canonical bytes and digest include every choice. Unknown fields/versions fail closed; never reinterpret or re-sign existing approval as v2.

Extend `internal/ai/drafts`, `writepreview` and `writeback` at their existing domain boundaries, with focused description policy code. Root `aiDraft*`, `aiWritePreview*` and `aiWrite*` adapters own new persistence and read translations; `internal/aiadapters/immichwrite` retains the sole non-retrying mutation transport. No new library is needed: installed JSON, hashing, UTF-8 and SQLite facilities suffice for a domain-specific owned-block format.

### D2. Model independent field readiness and baselines

The draft records a chosen primary language and preserve/replace/managed_append policy. Preserve is the default and omits description from the payload; a selected description must be nonempty, complete and current through generation/adoption or explicit review. Description-only staging does not require a camera point. GPS-only staging ignores unavailable/stale text and heading. Selecting neither writable field is invalid.

Extend fresh observations with an explicit description presence/value and separate read-availability status. The v3.2.2 adapter treats absent/null description as empty for text comparison only after valid asset/EXIF decoding, retains raw presence for audit, and rejects malformed types or unavailable reads. Compare exact Unicode string content without trim, whitespace/newline rewriting or Unicode normalization. Source identity remains ID/owner/type/checksum, not metadata timestamps. Conflict checks cover selected fields only; unrelated GPS cannot block description-only saving.

### D3. Own one append block conservatively

Allocate a stable random block identifier per owner/installation/asset append lineage and retain it in private audit. Serialize visible plain-text boundary lines with that identifier and a version, surrounding the chosen text. Store the complete prior owned block and its hash after verified readback. Preserve all text outside that block exactly; first append adds only the displayed separator and block. Later append replaces the same block, including a changed primary language, rather than adding another.

Recognize ownership through retained verified audit, never through a marker in arbitrary upstream text alone. Missing, edited, duplicate, nested or malformed expected boundaries produce a conflict; no repair, marker adoption or extra append is automatic. Marker-looking content without matching local ownership is preserved and forces explicit replacement or manual resolution. Include full before/after text in the preview. Enforce a 64 KiB UTF-8 cap on both observed and final descriptions and the existing 16 KiB per-language draft cap; reject overflow without truncating user text. Policy changes or edits invalidate old previews.

### D4. Preserve one request and conservative reconciliation

Send one PATCH per target standard step containing exactly selected GPS components and/or description. Retain the 20-second single-attempt transport deadline, no redirects, no alternate method and no hidden retries. Fresh authority/source/selected-field comparison occurs before the durable reservation and is fenced again before send. A no-op requires every selected value to equal its before value and current readback.

Read back each selected field independently: GPS tolerance remains `1e-7` degrees, description uses exact string comparison under the stated absent/empty rule. All intended fields must match to mark the standard step verified. Mixed baseline/intended values after a combined call are recorded as partial, preserve evidence of observed successes, and do not allow replay of the combined payload. After the prior sender is known complete, the user can review a new plan for remaining fields. All-baseline readback permits CH19's bounded explicit second attempt only with valid unchanged authority and established sender completion. Third values conflict; unavailable reads remain unresolved. Idempotent confirmation/retry recovery happens locally before new eligibility reads.

Verified GPS alone may refresh the catalog pair under the existing sync fence even when another field failed, but the UI cannot call the entire operation successful. Description results live in the private operation projection; do not invent a second catalog or change gallery GPS membership for a description-only operation.

### D5. Preserve immutable v1 storage and recovery

Use additive versioned preview/operation/field-state tables rather than widening migration 027/028 CHECK constraints or rewriting their immutable payloads. Bound each new canonical plan to 1 MiB, new private text observations to the caps above, and audit pagination/events per retained step. New history readers merge owner-scoped v1/v2 records with stable ordering. Both formats share the existing installation/asset exclusion table so different versions/accounts cannot write the same photo concurrently. Legacy ambiguous guards survive upgrade and account deletion unchanged.

Apply the maintained preview expiry and capacity limits across all supported plan versions together. Adding versioned tables must not multiply an owner's temporary-preview quota or permit cleanup to delete retained approval/audit evidence. Extend P11–P13 regression fixtures with mixed-version records at the capacity boundary and after reopen.

Draft mutation guards must consult both operation formats atomically. Only an exact reviewed revision may be confirmed; active/possibly sent work freezes edits as in CH19. New storage has composite owner/installation foreign keys, retained preview/draft protection, account deletion fencing and installation invalidation. Additive migration allocation is serialized after CH17; test fresh and 028-to-current upgrade/reopen with queued, verified and ambiguous v1 records.

### D6. UI, rollout and verification

Extend typed draft/preview/operation DTOs and AI-only controls. Show independent checkboxes, primary language and policy, full text comparisons, per-field status and revision/expiry. Confirmation identity recovery, late-history fencing, HTTP-origin support and manual pending conflict resolution remain intact. A GPS overlap requires resolution; a description-only operation does not discard or send manual GPS.

See [verification plan](verification-plan.md): pure policy tests, real SQLite migration/concurrency tests, local socket payload/count/readback tests, RTL and a built keyboard browser journey cover selected-field behavior and recovery. Apply uses Plus TDD, GitNexus impact and the full thirteen gates with the adopted coverage floors. No new harness is assumed. Add a backend operator capability allowlist alongside the existing write-enable/profile controls: existing GPS activation does not enable descriptions, stack GPS or metadata. New capabilities default off; operator activation attests the recorded authorized evidence for the actual installation/profile. Freeze its policy identity in new plans and recheck it before dispatch. Keep description writes disabled until description-specific GATE-02 evidence; CH19 GPS evidence does not suffice.

For rollout, back up SQLite and preserve the encryption key; deploy a binary that reads both versions before admitting v2 plans. For rollback, stop new admissions/sends, retain unresolved guards and use a compatible binary capable of both formats. Never restore an older database over retained approvals or run destructive down migrations as recovery.

## Risks / Trade-offs

- External edits can occur after the last read → conflict detection narrows but cannot remove the race; do not claim remote atomicity.
- A combined request may yield mixed observations → require a fresh plan for remaining fields instead of resending successful fields.
- Conservative append ownership can block a manually edited block → show conflict and full text, protecting existing user content over automatic repair.
- Versioned tables add read-adapter work → preserve immutable historical bytes and unresolved authority without an unsafe migration rewrite.
