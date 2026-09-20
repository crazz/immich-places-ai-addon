-- +goose Up
CREATE TABLE ai_job_context (
 userID TEXT NOT NULL,
 jobID TEXT NOT NULL,
 itemID TEXT NOT NULL,
 bundleJSON TEXT NOT NULL CHECK(length(CAST(bundleJSON AS BLOB))<=32768),
 createdAt INTEGER NOT NULL,
 PRIMARY KEY(userID,jobID,itemID),
 FOREIGN KEY(userID,jobID,itemID) REFERENCES ai_job_items(userID,jobID,id) ON DELETE CASCADE
);
-- +goose StatementBegin
CREATE TRIGGER ai_job_context_immutable BEFORE UPDATE ON ai_job_context
BEGIN
 SELECT RAISE(ABORT,'AI context is immutable');
END;
-- +goose StatementEnd

-- +goose Down
DROP TABLE ai_job_context;
