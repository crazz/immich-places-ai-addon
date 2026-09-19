-- +goose Up
CREATE TABLE ai_installation_identity (
 singleton INTEGER PRIMARY KEY CHECK (singleton=1),
 id TEXT NOT NULL,
 fingerprint TEXT NOT NULL
);
CREATE TABLE ai_selection_snapshots (
    userID TEXT NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
    id TEXT NOT NULL,
    expiresAt INTEGER NOT NULL,
    manifest TEXT NOT NULL,
    digest TEXT NOT NULL,
    PRIMARY KEY (userID, id)
);
CREATE INDEX ai_selection_expiry ON ai_selection_snapshots(expiresAt);
CREATE TABLE ai_selection_items (
    userID TEXT NOT NULL,
    snapshotID TEXT NOT NULL,
    position INTEGER NOT NULL CHECK (position >= 0),
    assetID TEXT NOT NULL,
    facts TEXT NOT NULL,
    PRIMARY KEY (userID, snapshotID, position),
    UNIQUE (userID, snapshotID, assetID),
    FOREIGN KEY (userID, snapshotID) REFERENCES ai_selection_snapshots(userID, id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE ai_selection_items;
DROP TABLE ai_selection_snapshots;
DROP TABLE ai_installation_identity;
