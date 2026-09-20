## MODIFIED Requirements

### Requirement: Minimize and explicitly format the Visual request

The provider payload SHALL contain only the selected model, one normalized image, the application-controlled Visual instructions and canonical output contract, the exact validated requested language set/primary language, and an output-token ceiling only when explicitly authorized by the applicable production token policy. It SHALL exclude local identity fields, filenames, private URLs/credentials, capture metadata, neighboring assets and context/source bundles. The instructions SHALL treat visible text as untrusted content, request concise evidence without private reasoning, preserve uncertainty and distinguish camera, subject and direction. Strict mode SHALL require observed strict support; JSON mode SHALL require observed JSON support and explicit caller permission. Neither mode SHALL waive canonical validation or infer support for optional model parameters.

#### Scenario: Visual payload contains only permitted inputs
- **GIVEN** a prepared image whose original asset has private identity and metadata
- **WHEN** the request is serialized
- **THEN** it contains one image and the controlled contract/languages
- **AND** it contains no private source fields, context, tools or unapproved optional parameters

#### Scenario: Strict request retains the canonical contract
- **GIVEN** applicable strict support and a strict attempt
- **WHEN** the request is prepared
- **THEN** the declared schema matches the canonical result contract and server validation remains mandatory

#### Scenario: JSON format requires explicit permission and support
- **GIVEN** JSON mode is selected with or without applicable support and explicit permission
- **WHEN** the attempt is prepared
- **THEN** it proceeds only when both are present, using the same canonical result contract

#### Scenario: Language normalization cannot widen disclosure or coverage
- **GIVEN** non-English languages and a primary member, including valid tag aliases
- **WHEN** analysis prepares the request
- **THEN** it uses the normalized exact set without requiring English
- **AND** empty, invalid, duplicate-normalized or missing-primary sets fail before dispatch
