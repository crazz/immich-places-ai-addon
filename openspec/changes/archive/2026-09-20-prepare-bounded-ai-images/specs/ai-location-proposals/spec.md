## ADDED Requirements

### Requirement: Prepare only the exact currently authorized image

Image preparation SHALL accept one exact asset bound to the requesting application user and current installation. It SHALL require enabled AI, current local catalog eligibility and a usable personal Immich credential, then verify current upstream access and image eligibility before retaining a prepared result. Cached membership alone SHALL NOT establish current access. Hidden, unavailable, trashed, unsupported-type and suppressed stack-child assets SHALL fail without substituting another asset. Caller-supplied source URLs or provider destinations SHALL NOT choose the image source. Preparation SHALL convey no analysis consent or write authority.

#### Scenario: Eligible authorized image is prepared
- **GIVEN** an enabled installation and one locally eligible image accessible through the requesting user's Immich credential
- **WHEN** that exact asset is prepared
- **THEN** the result contains only a prepared copy of that asset
- **AND** the original asset remains unchanged

#### Scenario: Another user's local asset cannot be fetched
- **GIVEN** the asset is absent from the requesting user's eligible catalog or belongs only to another application user
- **WHEN** preparation is requested
- **THEN** preparation fails without an upstream image request or disclosure of the other user's asset details

#### Scenario: Disabled or stale installation cannot prepare an image
- **GIVEN** AI is disabled, the installation binding has changed, the account is absent or its credential is missing
- **WHEN** preparation is requested
- **THEN** no prepared image is returned and no image is fetched

#### Scenario: Current upstream access or eligibility is lost
- **GIVEN** a locally eligible asset whose current upstream state is inaccessible, hidden, locked, trashed, non-image or a suppressed stack child
- **WHEN** preparation verifies that state
- **THEN** it rejects the asset without fetching an alternative image or retaining a prepared result

#### Scenario: Invalid identity cannot choose a network path
- **GIVEN** a missing or malformed asset identity, an attempted source URL or a different installation identity
- **WHEN** preparation is requested
- **THEN** it fails without using that input as a network destination or path outside the configured asset operation

### Requirement: Bound and isolate authenticated image retrieval

Retrieval SHALL use only the configured server-side Immich connection and the requesting user's current credential. It SHALL enforce finite metadata/image byte limits and an overall deadline, honor cancellation, reject redirects and perform no hidden retries. Credentials SHALL remain confined to the configured Immich destination. Declared length SHALL NOT replace actual streamed-byte enforcement. All response bodies SHALL be released on success and failure; retrieval SHALL NOT expose a partial image as success.

#### Scenario: Exact bounded preview retrieval succeeds
- **GIVEN** a currently authorized image whose metadata and preview fit the configured limits
- **WHEN** retrieval completes within its deadline
- **THEN** only the configured exact-asset read operations occur
- **AND** the preview body is fully bounded and released

#### Scenario: Redirect cannot forward credentials
- **GIVEN** a metadata or image response redirects to the same host or another destination
- **WHEN** retrieval receives the redirect
- **THEN** preparation fails without following it or forwarding the user's credential

#### Scenario: Declared and streamed sizes are independently bounded
- **GIVEN** a response advertises an excessive length or exceeds the limit while streaming with a missing or misleading length
- **WHEN** retrieval processes the response
- **THEN** the complete operation fails at the byte boundary without retaining a partial image
- **AND** a response exactly at the limit remains eligible for subsequent validation

#### Scenario: Cancellation and timeout stop retrieval
- **GIVEN** preparation is canceled or its deadline expires before or during a response
- **WHEN** retrieval observes the interruption
- **THEN** it releases resources and returns no prepared image
- **AND** no retry or later image request begins

#### Scenario: Failed upstream response is not retried or exposed
- **GIVEN** an upstream authorization failure, unavailable asset, rate limit, server error, malformed metadata or interrupted stream
- **WHEN** retrieval fails
- **THEN** it returns a bounded safe failure with no partial prepared result
- **AND** no response body, credential or private URL is included in the failure

### Requirement: Normalize one bounded metadata-free raster copy

Preparation SHALL decode only documented supported static image inputs, enforce decoded dimensions and pixel count before allocating the full raster, normalize supported orientation, preserve aspect ratio without upscaling, and produce one explicitly typed image within the configured long-edge and encoded request-image limits. The result SHALL contain newly encoded pixels without source EXIF, GPS, comments, filenames or other source metadata. Unsupported orientation-bearing formats, animation, malformed images and exceeded limits SHALL fail the entire preparation; the system SHALL NOT silently choose a frame, ignore a required orientation transform, crop the scene, change the source asset or retry with an original file. Transparent pixels SHALL have a documented deterministic background when the output format cannot retain transparency.

#### Scenario: Orientation and aspect ratio survive normalization
- **GIVEN** a supported image with a valid orientation transform and dimensions above the output edge limit
- **WHEN** preparation succeeds
- **THEN** the visible image has the correct orientation and is proportionally scaled within the edge limit
- **AND** no scene content is cropped or mirrored incorrectly

#### Scenario: Small image is not enlarged
- **GIVEN** a supported image already within the configured dimensions
- **WHEN** preparation succeeds
- **THEN** its dimensions remain unchanged except for the required orientation transform

#### Scenario: Private source metadata is absent from the copy
- **GIVEN** an otherwise valid image containing location, camera, comment or other source metadata
- **WHEN** preparation succeeds
- **THEN** the copy retains the intended visible pixels without those metadata fields
- **AND** the original bytes and upstream asset remain unchanged

#### Scenario: Decode budgets reject oversized dimensions before full decode
- **GIVEN** a small encoded file declaring an excessive dimension, pixel count or invalid dimensions
- **WHEN** preparation inspects the input
- **THEN** it fails before full raster allocation and returns no prepared image
- **AND** valid exact-boundary dimensions may proceed

#### Scenario: Unsupported or malformed images fail explicitly
- **GIVEN** an unsupported format, animation, truncated image, conflicting format information or unsupported orientation-bearing metadata
- **WHEN** preparation validates the input
- **THEN** it fails without guessing a format, selecting a frame or publishing an incorrectly oriented image

#### Scenario: Encoded request-image limit includes transmission expansion
- **GIVEN** a normalized raster whose encoded image or transmission representation exceeds the configured image budget
- **WHEN** preparation attempts to publish it
- **THEN** the operation fails without returning a partial image or silently lowering quality through repeated attempts

#### Scenario: Transparency has deterministic output
- **GIVEN** a supported transparent image and an output format without transparency
- **WHEN** preparation succeeds
- **THEN** transparent pixels use the documented background and visible content remains correctly positioned

### Requirement: Discard preparation when its authority or source becomes stale

Preparation SHALL verify that the local account, credential, installation and asset eligibility remain valid before publishing the prepared copy. A detected upstream source revision or eligibility change during retrieval SHALL invalidate the complete preparation. Cancellation observed during processing SHALL prevent success even if decoding or resizing finishes. A prepared copy SHALL be bound to the source and preparation policy for later consumers, but SHALL NOT replace their current authorization and consent checks.

#### Scenario: Account or credential changes during retrieval
- **GIVEN** preparation has started and the account is deleted, its credential changes, the asset loses local eligibility or the installation changes
- **WHEN** preparation reaches a subsequent authorization check
- **THEN** the complete result is discarded without provider dispatch or publication

#### Scenario: Source changes while bytes are retrieved
- **GIVEN** the upstream image revision or eligibility differs between the preparation checks
- **WHEN** preparation detects that difference
- **THEN** it returns no prepared copy and does not mix revisions or retry against a different source

#### Scenario: Cancellation during processing prevents success
- **GIVEN** retrieval finished but processing is canceled before publication
- **WHEN** the processing boundary observes cancellation
- **THEN** no prepared copy is returned even if a bounded transformation already completed

### Requirement: Keep prepared images transient and without side effects

A successful prepared image SHALL own its bytes and expose only independent copies plus bounded preparation metadata needed by the next consumer. The zero or incomplete value SHALL NOT masquerade as a prepared image. Image bytes and credentials SHALL NOT enter ordinary logs, errors, default serialized application objects or persistent job records. Preparation SHALL make no provider calls, database writes or Immich mutations. Temporary resources, if used, SHALL be cleaned on all outcomes; process restart SHALL NOT leave retained temporary images requiring manual cleanup.

#### Scenario: Returned data is independent
- **GIVEN** a successful prepared image
- **WHEN** its caller mutates the original input or returned byte slices
- **THEN** the retained image, dimensions, identity and digest remain unchanged
- **AND** an uninitialized value cannot be used as a prepared payload

#### Scenario: Failure and serialization do not disclose private image data
- **GIVEN** an input or upstream failure containing a credential, private URL, filename, location metadata or image bytes
- **WHEN** preparation reports failure or an application serializes its result wrapper
- **THEN** only bounded safe status or preparation metadata is exposed
- **AND** private bytes or credentials are not logged or serialized implicitly

#### Scenario: Preparation has no persistent or mutating effects
- **GIVEN** successful, failed, canceled and concurrent preparations of synthetic images
- **WHEN** each operation ends
- **THEN** provider, persistence-write and Immich-mutation counters remain zero
- **AND** all transient resources are released without changing legacy preview, manual placement or GPX behavior
