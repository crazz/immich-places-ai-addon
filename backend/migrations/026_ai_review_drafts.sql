-- +goose Up
CREATE TABLE ai_drafts (
 userID TEXT NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
 installationID TEXT NOT NULL,
 id TEXT NOT NULL,
 analysisID TEXT NOT NULL,
 revision INTEGER NOT NULL CHECK(revision > 0),
 state TEXT NOT NULL CHECK(state IN ('draft','staged','rejected')),
 PRIMARY KEY(userID,installationID,id),
 UNIQUE(userID,installationID,analysisID),
 FOREIGN KEY(userID,analysisID) REFERENCES ai_analyses(userID,id) DEFERRABLE INITIALLY DEFERRED
);
CREATE TABLE ai_draft_revisions (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 draftID TEXT NOT NULL,
 revision INTEGER NOT NULL CHECK(revision > 0),
 content TEXT NOT NULL CHECK(length(CAST(content AS BLOB)) <= 131072),
 PRIMARY KEY(userID,installationID,draftID,revision),
 FOREIGN KEY(userID,installationID,draftID) REFERENCES ai_drafts(userID,installationID,id) ON DELETE CASCADE
);
-- +goose StatementBegin
CREATE TRIGGER ai_draft_revisions_immutable BEFORE UPDATE ON ai_draft_revisions
BEGIN
 SELECT RAISE(ABORT,'Draft revisions are immutable');
END;
-- +goose StatementEnd

CREATE TABLE ai_draft_baseline_observations (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 draftID TEXT NOT NULL,
 revision INTEGER NOT NULL,
 id TEXT NOT NULL,
 expiresAt INTEGER NOT NULL,
 content TEXT NOT NULL CHECK(length(CAST(content AS BLOB)) <= 8192),
 PRIMARY KEY(userID,installationID,draftID,id),
 FOREIGN KEY(userID,installationID,draftID,revision) REFERENCES ai_draft_revisions(userID,installationID,draftID,revision) ON DELETE CASCADE
);
CREATE INDEX ai_draft_observation_expiry ON ai_draft_baseline_observations(expiresAt);

-- +goose Down
DROP TABLE ai_draft_baseline_observations;
DROP TABLE ai_draft_revisions;
DROP TABLE ai_drafts;
