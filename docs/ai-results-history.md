# Private AI result history

CH14 adds **AI Results** to the authenticated application shell. History does not depend on catalog synchronization, Missing location membership, a current Immich key or enabled AI execution. It contains terminal succeeded, failed and canceled items from the current installation, scoped to the local account. Existing runs retain their original model, mode and proposal when later runs or profile edits occur.

The list separates execution, proposal, review and write states. A successful `unknown` outcome means analysis completed without establishing a location. Failed and canceled items have no invented proposal. Review is `unreviewed` and write is `not_requested`; browsing does not create drafts, change manual pending coordinates, call a provider or mutate Immich.

## Browse and inspect

Open **AI Results**, expand **Filter results** and choose terminal execution/outcome, asset or job ID, capture-date bounds or the selected album ID at launch. Album filtering uses the historical selected scope, not current membership. Dates preserve the recorded source-local calendar day. Runs without retained date/album facts remain explicitly unknown; there is no upload-time or current-catalog fallback. **Unknown only** ignores date bounds.

**Older results** follows an opaque continuation. **Refresh results** starts a new snapshot of terminal history. Opening a result and returning preserves filters, page, scroll and row focus. Reload restores navigation from IDs/filter values in the URL; private result text is not stored there. Errors offer explicit retry, and a failed refresh marks retained information as potentially stale.

Current photo previews have separate authorization. Every image read checks current installation, local account/key, visibility and upstream source access before and after preparation. Hidden, removed or inaccessible sources display **Current image unavailable** while their retained local history remains readable. Images are transient stripped JPEGs; account changes abort requests and release object URLs.

## Read API and persistence

All routes use the existing local session and `Cache-Control: no-store`:

- `GET /ai/results`: default 30, maximum 100 items; filters `state`, `outcome`, `assetId`, `jobId`, `startDate`, `endDate`, `albumId`, `undated=true`, and opaque `cursor`.
- `GET /ai/results/{analysisID}`: exact immutable successful detail.
- `GET /ai/jobs/{jobID}/items/{itemID}/result`: exact successful, failed or canceled detail.
- `GET /ai/jobs/{jobID}/items/{itemID}/thumbnail`: separately authorized current image.

A versioned encrypted cursor binds owner, installation, query and terminal-history watermark. Descending terminal time/job/item ordering plus the monotonic sequence excludes later completions even with tied or backward timestamps. Refresh includes them. Lists use indexed summaries, not full analysis payloads. Detail revalidates stored schema, semantic contract and original provenance; unsupported/corrupt records return a safe unavailable response without rewriting history.

Migration 025 adds an immutable terminal read projection with transactional completion/cancellation publication and provenance-only backfill. Existing job/item/analysis tables remain authoritative. No retention schedule or deletion policy is activated. Rollback should hide the new surface and restore a compatible backup when necessary; do not destructively downgrade retained history.

See [verification](engineering/ai-results-history-verification.md). [CH15 proposal inspection](ai-proposal-review.md) adds detailed geometry, evidence and language inspection. Draft acceptance and confirmed writing remain separate changes.
