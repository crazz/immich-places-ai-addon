-- +goose Up
CREATE TABLE ai_provider_profiles (
    userID TEXT NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
    id TEXT NOT NULL,
    activeRevision INTEGER NOT NULL CHECK (activeRevision > 0),
    enabled INTEGER NOT NULL CHECK (enabled IN (0, 1)),
    PRIMARY KEY (userID, id)
);

CREATE TABLE ai_provider_versions (
    userID TEXT NOT NULL,
    profileID TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK (revision > 0),
    name TEXT NOT NULL,
    baseURL TEXT NOT NULL,
    model TEXT NOT NULL,
    secretCiphertext TEXT,
    PRIMARY KEY (userID, profileID, revision),
    FOREIGN KEY (userID, profileID) REFERENCES ai_provider_profiles(userID, id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE ai_provider_versions;
DROP TABLE ai_provider_profiles;
