## ADDED Requirements

### Requirement: Require explicit production token compatibility policy

Production analysis SHALL require a complete administrator-controlled token policy bound to the exact owner, installation, profile revision/model and destination policy. It SHALL identify an approved output-limit parameter and conservative input/output bounds for the enforced request envelope. Operator-attested policy and observed capability results SHALL remain distinct. Unknown, stale, malformed or violated policy SHALL block production dispatch without guessing parameters or probing private input. Safe readiness MAY be exposed to the profile owner without secrets or evidence documents.

#### Scenario: Use an explicitly supported output limit
- **GIVEN** a current complete policy and applicable image/format observations
- **WHEN** a bounded production request is authorized
- **THEN** it uses only the approved output-limit field and reserves the policy's full permitted input plus chosen output allowance

#### Scenario: Refuse unknown or stale token support
- **GIVEN** absent, conflicting, incomplete or outdated token policy, or a request outside its bounded envelope
- **WHEN** production admission or dispatch is attempted
- **THEN** it fails safely without guessing a field, inferring support from another observation or sending a private probe

#### Scenario: Read readiness without changing evidence
- **GIVEN** a saved profile and observed capability results
- **WHEN** the owner reads readiness or edits the profile
- **THEN** the response distinguishes attested policy from observed support, a changed revision needs its own policy, and no automatic provider call occurs
