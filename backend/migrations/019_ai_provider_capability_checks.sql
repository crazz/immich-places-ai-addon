-- +goose Up
CREATE TABLE ai_provider_capability_checks (
    userID TEXT NOT NULL,
    profileID TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK (revision > 0),
    attemptID TEXT NOT NULL,
    protocolVersion TEXT NOT NULL,
    policyFingerprint TEXT NOT NULL,
    lifecycle TEXT NOT NULL CHECK (lifecycle IN ('running', 'completed', 'failed', 'canceled', 'interrupted', 'superseded')),
    startedAt TEXT NOT NULL,
    deadlineAt TEXT NOT NULL,
    completedAt TEXT,
    requestedModel TEXT NOT NULL,
    reportedModel TEXT,
    observationsJSON TEXT NOT NULL,
    compatibility TEXT NOT NULL,
    PRIMARY KEY (userID, profileID, revision),
    UNIQUE (attemptID),
    FOREIGN KEY (userID, profileID, revision) REFERENCES ai_provider_versions(userID, profileID, revision) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE ai_provider_capability_checks;
