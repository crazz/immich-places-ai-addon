# Explicit stack GPS targets

CH21 adds an optional, explicitly reviewed target list to the confirmed writer. The analyzed photo remains the default. Implementation and synthetic verification are complete; live stack compatibility remains unverified and ordinary writes remain disabled by default.

## Review and confirmation

Stage a draft with GPS selected and a valid camera pair, then choose **Review stack members**. Add members individually, read their current baselines, and acknowledge the source and GPS for every selected photo. A stack can contain different viewpoints; membership is not evidence that the proposed camera location fits every image.

The comparison shows each selected photo ID and exact before/intended GPS. The analyzed photo can also carry its independently selected description and exact language/policy comparison. Siblings receive GPS only. Confirmation submits the immutable preview identity, digest and idempotency key; it cannot substitute coordinates, fields or members.

Selection includes the analyzed photo and at most 50 distinct IDs. Missing, partial and zero GPS remain distinct. Hidden, trashed, inaccessible, non-image or foreign-stack selections cannot produce a usable partial substitute. Additional unselected members never expand a saved plan. A selected sibling that leaves the stack conflicts before its send.

Review and preview have separate five-minute validity periods. A published immutable plan keeps its own validity; an expired source review cannot authorize a new preview. A fresh preview rechecks every selected baseline and source. Both review/preview use bounded reads and make no mutation or provider call.

Manual pending GPS overlapping any selected target must be resolved explicitly before continuing. Unrelated manual edits remain pending. Map failure does not hide the keyboard-accessible numeric comparison. Account, result, draft-revision and selection changes fence delayed private responses.

## Independent outcomes

One confirmation reserves all target exclusions atomically or none. Execution uses sequential single-asset requests through the existing bounded writer runtime. This is not remote batch atomicity: some targets may finish while others conflict, expire, fail or remain unresolved.

Each target retains its own attempts, generation, sender-completion evidence, fields and audit. Every standard step permits at most two mutation attempts, with an explicit eligible retry required for the second. Status checks only reconcile. Repeating an accepted retry generation returns current private state locally, even after restart, expiry or disablement; it never allocates another attempt or resends successful siblings.

Already-matching approved values are verified without a mutation attempt. A successful HTTP response is insufficient: source-consistent readback must verify the selected fields. Partial description/GPS evidence is retained independently and blocks replay of the combined payload. Only verified GPS refreshes local coordinates. An unavailable local catalog row remains missing and exposes pending local refresh; recovering local storage does not resend the mutation.

Expiry blocks unstarted attempts while preserving earlier outcomes and read-only reconciliation. Any possibly acting target protects the approved draft revision. Disablement/shutdown prevent new sends without claiming to undo prior requests. Account deletion removes owned private records but can retain an opaque unresolved exclusion. Installation changes cannot reuse new authority for old operations.

## API and storage

- `POST /ai/stack-reviews`: staged draft ID/revision and optional exact `targetIds`. Omission selects only the analyzed photo. Returns a private bounded observation; candidate IDs are not write authority.
- `POST /ai/write-previews`: optional `stackReviewId` refers to that observation. Two or more targets produce canonical `stack-preview-v3` with an exact target/field matrix; analyzed-only plans retain v1/v2 formats.
- Existing confirmation, lookup-by-key, operation GET and draft history routes retain their protected owner/installation scope. A v3 operation contains `targets`; aggregate status grants no attempt or retry authority.
- `POST /ai/write-operations/{operation}/targets/{asset}/retry` and `/reconcile` require the observed target generation. Legacy aggregate retry/reconcile rejects a v3 operation with `TARGET_REQUIRED`.

Migrations 036–042 add private reviews/previews, immutable v3 approvals, independent target/event/field rows, deletion guard cleanup, typed append lineages and a history projection. Existing v1/v2 payload bytes, counters, audit and unresolved exclusions remain intact. Shared installation/asset guards prevent overlapping approvals across versions and accounts.

Back up the database and encryption key before upgrade. Preserve every unresolved guard and audit during recovery. A rollback requires a binary that understands v3; an older GPS-only binary cannot safely reconcile it. Do not remove the new tables or restore an older database over newer operations.

## Capability and verification boundary

The configured write policy must explicitly include `stack_gps`, bound to the installation, `immich-v3.2.2` profile and evidence identifier. Description selection additionally requires `description`. The immutable plan binds the exact policy identity; policy removal or replacement cannot authorize its dispatch. Enabling single-photo GPS or description alone does not enable stack writing.

Synthetic SQLite, HTTP and browser fixtures verify local contracts only. Live stack capability needs separately authorized, exact-member mutation/readback evidence for the deployed version, rights and adapter. CH19's single-photo live GATE-02 is still separate and pending. Read access, a stack relation and an offline test pass do not establish live write compatibility, sidecar completion or remote compare-and-swap protection.

See the [implementation verification record](../openspec/changes/confirm-explicit-ai-stack-targets/implementation-verification.md), [description writes](ai-description-writes.md) and [confirmed GPS operations](ai-gps-writes.md).
