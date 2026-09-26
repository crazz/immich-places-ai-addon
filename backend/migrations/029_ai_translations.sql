-- +goose Up
CREATE TABLE ai_translation_runs (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 id TEXT NOT NULL,
 draftID TEXT NOT NULL,
 revision INTEGER NOT NULL,
 idempotencyKey TEXT NOT NULL,
 requestDigest TEXT NOT NULL,
 requestJSON TEXT NOT NULL CHECK(length(CAST(requestJSON AS BLOB)) <= 131072),
 policyID TEXT NOT NULL,
 policyJSON TEXT NOT NULL CHECK(length(CAST(policyJSON AS BLOB)) <= 8192),
 createdAt INTEGER NOT NULL,
 PRIMARY KEY(userID,installationID,id),
 UNIQUE(userID,installationID,idempotencyKey),
 FOREIGN KEY(userID,installationID,draftID,revision) REFERENCES ai_draft_revisions(userID,installationID,draftID,revision) ON DELETE CASCADE
);
CREATE INDEX ai_translation_draft ON ai_translation_runs(userID,installationID,draftID,createdAt,id);
-- +goose StatementBegin
CREATE TRIGGER ai_translation_runs_immutable BEFORE UPDATE ON ai_translation_runs
BEGIN
 SELECT RAISE(ABORT,'Translation inputs are immutable');
END;
-- +goose StatementEnd
CREATE TABLE ai_translation_items (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 runID TEXT NOT NULL,
 language TEXT NOT NULL,
 state TEXT NOT NULL CHECK(state IN ('queued','reserved','complete','unavailable','failed','canceled','interrupted')),
 text TEXT CHECK(text IS NULL OR length(CAST(text AS BLOB)) <= 16384),
 failure TEXT NOT NULL DEFAULT '',
 cancelRequested INTEGER NOT NULL DEFAULT 0 CHECK(cancelRequested IN (0,1)),
 PRIMARY KEY(userID,installationID,runID,language),
 FOREIGN KEY(userID,installationID,runID) REFERENCES ai_translation_runs(userID,installationID,id) ON DELETE CASCADE
);
CREATE INDEX ai_translation_pending ON ai_translation_items(state,userID,installationID,runID);

CREATE TABLE ai_translation_usage (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 runID TEXT NOT NULL,
 language TEXT NOT NULL,
 inputReserved INTEGER NOT NULL CHECK(inputReserved>0),
 outputReserved INTEGER NOT NULL CHECK(outputReserved>0),
 estimatedMicros INTEGER CHECK(estimatedMicros IS NULL OR estimatedMicros>=0),
 inputReported INTEGER,
 outputReported INTEGER,
 totalReported INTEGER,
 PRIMARY KEY(userID,installationID,runID,language),
 FOREIGN KEY(userID,installationID,runID,language) REFERENCES ai_translation_items(userID,installationID,runID,language) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE ai_translation_usage;
DROP TABLE ai_translation_items;
DROP TABLE ai_translation_runs;
