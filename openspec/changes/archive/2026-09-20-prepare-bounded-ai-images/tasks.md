## 1. Produce a bounded metadata-free analysis copy

- [x] 1.1 Normalize supported static images with correct orientation, proportional downscaling, deterministic transparency and no source metadata, using the approved pure-Go dependencies.
- [x] 1.2 Enforce source, decoded-raster and encoded transmission limits; reject malformed, animated and unsupported orientation-bearing inputs without partial output or fallback attempts.
- [x] 1.3 Expose an independently owned, explicitly releasable prepared value with source/policy binding, safe serialization and cancellation-aware publication.

## 2. Prepare only the exact currently authorized Immich image

- [x] 2.1 Bind preparation to the current enabled installation, application account, personal credential and eligible owner-scoped catalog asset, rejecting invalid identities before network access.
- [x] 2.2 Retrieve current metadata and the exact preview through a bounded, non-retrying read-only transport that rejects redirects, closes response bodies and reports only safe errors.

## 3. Discard stale preparations and preserve existing workflows

- [x] 3.1 Reject preparations whose account, credential, installation, local eligibility or upstream source changes before publication, including cancellation during processing.
- [x] 3.2 Preserve zero provider calls, persistence writes and Immich mutations across successful, failed and concurrent preparation, and document the internal handoff and verified limits without changing legacy image/manual/GPX behavior.
