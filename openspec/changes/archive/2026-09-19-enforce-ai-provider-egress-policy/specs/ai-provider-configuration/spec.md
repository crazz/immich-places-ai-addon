## ADDED Requirements

### Requirement: Require administrator approval for provider destinations

Every provider dispatch SHALL require an installation policy approving the exact canonical API base URL, including scheme, host, effective port and base path. An absent or empty policy SHALL deny all provider dispatches. Invalid policy configuration SHALL prevent an AI-enabled backend from starting. User profile configuration SHALL NOT grant destination approval. HTTPS SHALL be required unless the administrator explicitly approves HTTP for a specific local endpoint with bounded allowed network ranges. Wildcard hosts, credential-bearing URLs, queries, fragments and ambiguous paths SHALL NOT be accepted as approval rules.

#### Scenario: Use the existing approved NAS proxy
- **GIVEN** an enabled owned profile references the existing codex-proxy endpoint
- **AND** the administrator has explicitly approved its exact local base URL and network range
- **WHEN** an authorized internal operation dispatches through that profile
- **THEN** only the approved endpoint can receive the request
- **AND** no additional proxy service or global proxy model change is required

#### Scenario: Saved settings do not approve a destination
- **GIVEN** no destination policy or a policy approving a different base URL
- **WHEN** a user saves a syntactically valid provider profile and an internal operation later attempts dispatch
- **THEN** saving succeeds without a network call
- **AND** dispatch fails with a safe policy error before contacting the unapproved destination

#### Scenario: Reject mismatched or ambiguous destinations
- **GIVEN** a rule approving one scheme, host, effective port and base path
- **WHEN** dispatch targets a different value or uses credentials, query, fragment, encoded separators or dot segments to alter the destination
- **THEN** it fails without contacting that destination or forwarding credentials

#### Scenario: Invalid administrator policy fails closed
- **GIVEN** AI is enabled and the administrator supplies malformed rules, wildcard hosts or an unbounded local-network exception
- **WHEN** the backend starts
- **THEN** it rejects the invalid configuration with a redacted actionable error
- **AND** an empty valid policy remains usable for offline profile management

### Requirement: Validate the actual provider network destination

Provider dispatch SHALL validate resolved addresses against the approved endpoint's address policy at connection time and SHALL connect only to a validated address. Public endpoints SHALL resolve only to permitted globally routable addresses. Local endpoints SHALL resolve only within explicitly approved private or loopback ranges. Empty or mixed allowed/disallowed answers SHALL fail closed. Unspecified, multicast, link-local and cloud metadata destinations SHALL remain prohibited even under a local approval. Environment proxy settings SHALL NOT bypass this boundary. HTTPS certificate and hostname validation SHALL remain enabled.

#### Scenario: Resolve an approved Docker hostname
- **GIVEN** the approved codex-proxy hostname resolves entirely within its permitted Docker network range
- **WHEN** a dispatch establishes a connection
- **THEN** it connects to an address from that validated answer for the approved hostname and port

#### Scenario: Reject changed, mixed or absent DNS answers
- **GIVEN** an approved hostname previously resolved to an allowed address
- **WHEN** a later dispatch resolves an unapproved address, a mixed allowed/disallowed set or no address
- **THEN** that dispatch fails without sending an HTTP request or credential to any resolved destination
- **AND** a second uncontrolled resolution cannot substitute an address after validation

#### Scenario: A local exception cannot authorize metadata access
- **GIVEN** the administrator has approved a local provider
- **WHEN** its destination is or resolves to an unspecified, multicast, link-local or cloud metadata address, including an equivalent IPv4-mapped IPv6 form
- **THEN** dispatch is rejected without transmitting an HTTP request or credential

#### Scenario: Preserve direct routing and TLS verification
- **GIVEN** an approved HTTPS endpoint and environment proxy settings
- **WHEN** dispatch runs or the endpoint presents a certificate invalid for its hostname
- **THEN** environment proxy settings receive no provider traffic
- **AND** invalid TLS fails without sending the HTTP payload or authorization header

### Requirement: Confine credentials to a single approved request

Provider credentials SHALL be loaded only for the authorized owner and exact immutable revision and sent only to that revision's approved destination. Plaintext or corrupt stored credentials SHALL fail closed. Provider requests SHALL NOT forward browser cookies, application session credentials, Immich credentials or caller-supplied headers. All redirects SHALL be rejected, including same-origin redirects, without replaying the body or authorization. No automatic destination, model or credential fallback SHALL occur.

#### Scenario: Send only the revision-bound provider credential
- **GIVEN** an authorized profile revision with an encrypted provider credential
- **WHEN** its approved destination receives a request
- **THEN** it receives only that revision's provider authorization and explicitly permitted transport headers
- **AND** it receives no browser or Immich credential

#### Scenario: Stop every redirect
- **GIVEN** an approved provider returns a redirect to the same origin, another origin or an internal service
- **WHEN** the response is processed
- **THEN** the operation reports a redirect failure
- **AND** the redirect target receives no follow-up request, body or credential

#### Scenario: Corrupt or removed credentials cannot be reused
- **GIVEN** a stored credential is corrupt or plaintext, or credential removal has erased it across revisions
- **WHEN** an internal operation attempts dispatch
- **THEN** corrupt or plaintext credentials produce a safe failure before a provider request
- **AND** credential material removed before dispatch admission is never sent, including material obtained before the removal
- **AND** a deliberately credential-free current configuration can still use an approved endpoint that needs no secret

### Requirement: Recheck profile authority at provider dispatch

Dispatch SHALL carry a server-established owner identity, profile ID and exact immutable revision; it SHALL NOT substitute the active revision. The backend SHALL verify ownership, revision existence, current installation/profile enablement and current credential state before resolution and again immediately before admitting the outbound request. Disablement SHALL apply across historical revisions. Rejection SHALL reveal no other user's configuration and SHALL produce no provider request. Changes after dispatch admission SHALL prevent subsequent calls; already admitted requests are in flight and cannot be recalled reliably.

#### Scenario: Dispatch the exact authorized revision
- **GIVEN** an enabled owned profile has multiple immutable revisions
- **WHEN** an authorized internal operation names one existing revision
- **THEN** the operation uses that revision's destination, model configuration and current credential state
- **AND** a newer revision is not substituted silently

#### Scenario: Deny missing, foreign or disabled profiles
- **GIVEN** a missing revision, another user's profile, a deleted owner, disabled AI or a disabled logical profile
- **WHEN** dispatch is requested
- **THEN** the request fails before DNS resolution or provider contact
- **AND** missing and foreign profiles have indistinguishable unavailable responses

#### Scenario: Revocation while resolution is pending
- **GIVEN** a request has passed its initial authority check but has not been admitted for transmission
- **WHEN** profile disablement, account deletion or credential removal completes before the final authority check
- **THEN** the pending operation is denied or uses the now credential-free state as applicable
- **AND** no removed credential or request for a disabled/deleted profile is transmitted

### Requirement: Bound provider transport and preserve offline settings

Every provider operation SHALL have finite request, response, header and time limits with cancellation propagation. The transport SHALL make at most one application request per explicit dispatch, with no automatic retries, redirect replay or mode fallback. Oversized requests SHALL fail before transmission; oversized responses, cancellation and deadline expiry SHALL stop processing and produce safe structured errors. Raw provider bodies, credentials and image data SHALL NOT appear in ordinary errors or logs. CH02 SHALL expose no public dispatch endpoint and SHALL preserve zero provider/Immich calls from profile creation, listing, editing and disablement, and zero provider calls from existing AI-disabled workflows.

#### Scenario: Complete one bounded provider operation
- **GIVEN** an authorized request within all configured transport limits
- **WHEN** the approved endpoint returns a response within those limits
- **THEN** the caller receives its bounded response from one request
- **AND** the transport does not invoke Immich or select another provider or model

#### Scenario: Reject payloads and stop slow responses
- **GIVEN** a request exceeding its limit, an oversized response or headers, an expired deadline or caller cancellation
- **WHEN** dispatch or response processing reaches the violated bound
- **THEN** the operation fails safely and stops the affected I/O
- **AND** an oversized request is never transmitted
- **AND** no partial response is reported as success

#### Scenario: Errors do not cause hidden retries or leaks
- **GIVEN** a provider returns an authentication error, rate limit, server error or unsafe error body, or the network response is lost
- **WHEN** the transport reports the failure
- **THEN** it reports a stable redacted category and request identifier
- **AND** no automatic retry, model substitution or credential-bearing diagnostic is emitted

#### Scenario: Settings and existing workflows stay offline from providers
- **GIVEN** CH02 is installed with AI enabled or disabled
- **WHEN** users perform permitted profile operations or existing browsing, manual placement and GPX operations
- **THEN** those operations preserve their existing behavior and send no provider requests
- **AND** no public generic provider-dispatch route is available
