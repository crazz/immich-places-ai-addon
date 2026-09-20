# AI batch launch and progress

CH13 connects gallery selection, explicit disclosure and durable Visual or Context-assisted jobs. Open **AI Locate** beside Settings. This workflow creates analysis proposals and never writes Immich metadata or populates manual pending coordinates.

## Preview and authorize

Choose selected assets, the current page or all matching assets. The server freezes exact eligible membership and reports counts, exclusions and expiry; at most 500 eligible assets may enter a job. Changing page, selection or filters clears the preview and confirmation. An active GPX preview/status filter, an unopened album/folder listing or an unsettled catalog page blocks a new preview with an explanation.

The form defaults to Visual, strict schema output and editable English language choices. It identifies the exact provider/model/revision and current capability/execution-policy readiness. Calls, total tokens and per-call output tokens are finite; a smaller allowance may finish only part of a batch. Estimated cost caps use the displayed currency and require an attested tariff. Unknown reported usage or cost remains unknown.

Image disclosure starts unchecked. Context-assisted mode additionally requires an explicit choice for each of capture time, the concrete selected album, user hint and nearby known locations. Hints are limited to 2,000 UTF-8 bytes and are sent only when their class is selected. Nearby context includes at most six metadata records within six hours, with no neighbor images. Empty context remains a valid Context-assisted choice. Visual sends no context envelope. Configuration changes reset confirmation.

An execution policy must explicitly contain `"context": true` before Context-assisted admission is ready. This attests that its existing request-size and token allowances cover the bounded context envelope. Existing Visual policies retain their identity when this optional field is absent. Provider tests never start automatically from launch.

## Durable execution and observation

The first Context attempt stores the exact minimized CH10 bundle before reserving a provider call. Migration 024 adds owner/job/item-bound `ai_job_context` rows with a 32 KiB bound, immutable updates and item-cascade deletion. The stored projection, digest, source bindings, omissions and consent travel together. No image bytes are persisted. Retry and restart reuse the same bundle after freshness checks; changed evidence fails that item and requires an explicit new analysis. Completion validates against the stored mode and source set, never a model-supplied authority set.

An ambiguous submission exposes **Reconcile submission**, which repeats the identical admission and idempotency key. It does not retry automatically. **Start a different run** clears confirmation and obtains a new key only on the next explicit submit.

Saved jobs use bounded server pagination. Opening a job records its ID in `aiJob` URL state so reload resumes observation without another submission. Active jobs refresh every two seconds while visible and online. Hidden/offline observation pauses; returning resumes immediately. Failed reads back off to 30 seconds, and terminal/blocked jobs stop polling. Manual refresh remains available. Owner/job changes abort requests and hide superseded private responses.

Progress distinguishes execution states from successful `unknown` or `ambiguous` proposals. A polite status region announces state/count updates without moving focus. Closing the window stops observation only. Cancellation preserves completed results and cannot recall transmitted data or usage already incurred; repeat cancellation is safe. AI disablement prevents new launch while keeping existing read/cancel APIs available.

## Explicit reruns

**Retry failed** includes only terminal failed members of the owned parent. Reanalysis includes explicitly selected parent members, including successful, canceled or blocked items. Both return to fresh eligibility preview and unchecked disclosure, use a new run ID and persist the parent/kind in the confirmation-bound admission. The server checks every member against the owned parent inside admission. Neither action resets the parent or overwrites a result.

Detailed persistent result browsing belongs to CH14; the read-only proposal map belongs to CH15. See the [verification record](engineering/ai-batch-workflow-verification.md) and [production job contract](ai-production-jobs.md).
