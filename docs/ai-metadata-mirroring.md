# Optional reviewed metadata mirroring

CH22 implements an optional metadata step after verified standard fields. Implementation, synthetic verification and maintained-spec synchronization are complete. Live metadata compatibility remains unverified; the capability is disabled by default.

## Review and visibility

Select at least one standard field, then deliberately choose the metadata to share for the analyzed photo. Choices include reviewed direction and its method/uncertainty, precision, the chosen place, up to eight current reviewed description languages, and minimal provenance. Model identity requires provenance selection. Unknown direction stays unknown. Changing facts, text or selection invalidates an earlier preview.

The preview shows the complete namespace before/value comparison and asks the owner to acknowledge that authorized asset readers may see the exported content. The schema contains no private hints, neighboring-photo locations, credentials, raw prompts/responses or research URLs. The canonical structured value is limited to 64 KiB; the writer never silently truncates local text.

Metadata is optional. Unsupported or unverified installations retain local review and standard-only writing. Deliberately remove the mirror choice to prepare a standard-only plan. Inspection and preview make no Immich mutation or provider request.

## Exact namespace and independent results

The metadata step writes one `immich-places-ai-addon` key for the analyzed photo. Stack siblings receive only their independently approved GPS. Other custom metadata keys are preserved. A complete authorized list read distinguishes an absent namespace from unavailable or malformed data. Existing values require recognized locally verified ownership and an unchanged fresh baseline; a matching application label alone does not grant ownership.

The analyzed photo's standard fields must first verify and have settled sender completion. A standard no-op can satisfy that prerequisite. The metadata step then rechecks current authority, image identity, standard fields and namespace before its own reserved request. It sends one exact `PUT /api/assets/{asset}/metadata` request with one item. Redirects and transport failures do not trigger hidden retries or alternate routes.

A successful HTTP response is insufficient. Source-consistent readback compares the approved value semantically: object key order does not matter, while exact strings, numeric values and array order do. Metadata failure leaves verified standard fields, their local refresh and all local reviewed content intact. Combined history reports incomplete or partial work honestly.

Every metadata step has its own generation, attempts, sender evidence and audit. At most two mutation attempts are possible; the second requires explicit eligible retry. Recovery of metadata never resends successful standard fields. Repeating an accepted retry generation returns stored private state locally, including after restart, expiry or disablement. Unknown sender completion remains unsettled even when the intended value has been observed; reconciliation cannot turn that uncertainty into retry permission.

Manual pending GPS remains separate. The comparison and recovery controls work through text and keyboard without a map. Account/result changes fence delayed private responses.

## API and durable state

- Draft edits carry an optional `mirror` selection, or `null` to remove it. `reviewPrecision` explicitly reviews existing precision evidence.
- Selected previews require `mirrorDisclosure: "asset-readers-v1"`. They use `mirror-preview-v4`, an immutable standard target manifest and one analyzed-photo namespace plan. Old v1–v3 plans retain their original scope and bytes.
- Confirmation uses the existing protected preview ID, digest and stable idempotency key. One atomic approval creates the standard targets and a blocked metadata step.
- Operation reads expose standard `targets` and a separate `mirror` outcome. `verified` retains evidence that the approved value was observed; current status and observation distinguish later conflicts. `settled` additionally requires known sender completion.
- `POST /ai/write-operations/{operation}/targets/{asset}/metadata/retry` and `/metadata/reconcile` require the inspected `{ "generation": number }`. They remain owner/installation scoped and cannot select another step or payload.

Migrations 043–045 extend preview version storage, add private metadata steps/events and stable random record IDs, and preserve metadata-aware deletion guards and history. Populated old approvals receive no inferred mirror rows. Retained mirror state blocks destructive downgrade paths. Account deletion removes owned private records while preserving an opaque exclusion for a possible sender; it never deletes upstream metadata. Installation rotation cannot reuse new authority for old work.

## Capability and rollout boundary

The installation/profile policy must explicitly include `metadata`, with evidence for the configured `immich-v3.2.2` adapter and permissions. Description and stack capabilities remain independent. Read-only capability checks never issue trial writes. Policy removal or replacement prevents dispatch of an old plan.

Synthetic tests do not establish deployed compatibility, native Immich display, file/sidecar propagation or remote compare-and-swap protection. The pinned API supports list reads and item updates, but live mutation/readback and unrelated-key preservation still need separately authorized disposable-fixture evidence. CH19 GATE-02 and other field/scope live gates remain separate. The existing selected live photo has not been used for metadata mutation tests.

See [description writes](ai-description-writes.md), [explicit stack targets](ai-stack-writes.md) and the [CH22 verification plan](../openspec/changes/mirror-supported-ai-metadata/verification-plan.md).
