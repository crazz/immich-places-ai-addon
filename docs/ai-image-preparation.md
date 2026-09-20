# Internal AI image preparation

CH08 provides an internal read-only preparation boundary for later Visual analysis. It has no public endpoint, provider call, durable image storage or automatic startup consumer. CH09 consumes the prepared copy; CH12 will connect authorized analysis to durable jobs. Existing browsing, manual placement and GPX routes are unchanged.

## Caller contract

The root `aiImagePreparer` is constructed with the existing database, installation/selection state and server-configured Immich endpoint. Treat these construction settings as immutable. Pass an application owner, current installation ID and one exact asset ID to `prepare`; there is no source-URL input. Asset and installation IDs must be canonical UUIDs. Application owner IDs are local opaque identifiers and are not compared with Immich owner IDs.

Preparation checks current AI enablement, persisted installation binding, account credential and owner-scoped catalog eligibility before any request. Hidden libraries, hidden assets, non-images and suppressed stack children are excluded. Current personal-token access to Immich supplies an additional upstream permission check; cached catalog presence is insufficient.

Successful processing reads exact asset metadata, its `size=preview` image and metadata again. Local authority is rechecked before each subsequent network operation and publication. Changes to the credential, installation, relevant catalog facts or upstream source fingerprint discard the entire copy. Separate remote reads cannot guarantee an atomic source snapshot; later consumers must reauthorize and enforce current consent before provider dispatch.

## Raster policy and limits

| Boundary | Default / hard ceiling |
|---|---|
| Total preparation deadline | 30 seconds |
| Each metadata response | 1 MiB |
| Encoded preview source | 20 MiB |
| Decoded pixels | 40,000,000 |
| Single input dimension | 40,000 pixels |
| Output long edge | 2,048 pixels |
| Request-image representation | 10 MiB, including data-URL prefix and Base64 expansion |
| Expanded PNG text metadata | 64 KiB across all text chunks |
| Output | JPEG, quality 85, white transparency background |

Raster limits may be lowered through the internal policy. Invalid or above-ceiling values fail. There are no new environment variables or UI settings before an actual admission consumer exists.

Supported inputs are static JPEG, PNG and WebP with matching declared MIME. JPEG EXIF orientation is applied before proportional downscaling. Small images are not enlarged and no crop is introduced. PNG/WebP EXIF or XMP/text orientation unsupported by the chosen normalizer is rejected, as are animation, malformed containers, conflicting MIME and excessive dimensions or output. PNG compressed text is inspected through a bounded decoder. The resulting JPEG contains newly encoded pixels without original EXIF, GPS, comments or source application metadata. Color profiles are not retained; visual-quality evaluation remains a release concern.

Retrieval uses a dedicated client with no redirects, environment proxy, automatic decompression or reused-connection retries. Declared lengths and streamed bytes are independently limited. Errors and non-success response bodies are not forwarded to callers. No original-file fallback, alternate asset, implicit quality reduction or additional resolution attempt occurs.

## Ownership and failures

`images.Prepared` owns its encoded bytes behind a private synchronized state. `Bytes` returns an independent copy; `Info` returns explicit bounded identity, source/policy and output metadata. A zero, nil or released value cannot provide a payload. Call `Release` when the consumer finishes, and release or clear consumer-owned copies when appropriate. Copying the handle shares its release lifecycle; copying returned bytes does not.

Default JSON/Go formatting does not disclose retained bytes or source identities. Do not log payloads or serialize explicit byte copies into job records. Processing uses memory only. Release clears retained encoded bytes, but garbage collection and image libraries do not promise cryptographic erasure of every temporary allocation.

Failures return no prepared value: invalid/unsupported input, exceeded limit, unavailable/changed authority, unavailable upstream or context cancellation/deadline. Raw HTTP, SQL and decoder errors are not exposed. CPU transformations are bounded but cannot be interrupted mid-library call; stage-boundary and final checks prevent publication after observed cancellation.

## Verification boundary

Synthetic image, HTTP and real SQLite tests cover the maintained OpenSpec scenarios. The [verification record](engineering/ai-image-preparation-verification.md) records actual checks and remaining limits. No private images were retrieved or sent to a provider for this implementation. Compatibility with the pinned Immich v3.2.2 source contract is distinct from deployed-server image compatibility; GATE-02 remains unverified until a separately authorized live check.
