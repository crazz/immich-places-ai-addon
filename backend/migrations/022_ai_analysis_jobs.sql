-- +goose Up
CREATE TABLE ai_jobs (
 userID TEXT NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
 id TEXT NOT NULL,
 installationID TEXT NOT NULL,
 idempotencyKey TEXT NOT NULL,
 requestDigest TEXT NOT NULL,
 requestJSON TEXT NOT NULL,
 model TEXT NOT NULL,
 maxCalls INTEGER NOT NULL CHECK(maxCalls>0),
 calls INTEGER NOT NULL DEFAULT 0 CHECK(calls>=0 AND calls<=maxCalls),
 blocked INTEGER NOT NULL DEFAULT 0 CHECK(blocked IN (0,1)),
 cancelRequested INTEGER NOT NULL DEFAULT 0 CHECK(cancelRequested IN (0,1)),
 createdAt INTEGER NOT NULL,
 PRIMARY KEY(userID,id),
 UNIQUE(userID,installationID,idempotencyKey)
);
CREATE TABLE ai_job_items (
 userID TEXT NOT NULL,
 jobID TEXT NOT NULL,
 id TEXT NOT NULL,
 assetID TEXT NOT NULL,
 position INTEGER NOT NULL CHECK(position>=0),
 state TEXT NOT NULL DEFAULT 'queued' CHECK(state IN ('queued','running','retry_wait','blocked','succeeded','failed','canceled')),
 attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 3),
 calls INTEGER NOT NULL DEFAULT 0 CHECK(calls BETWEEN 0 AND 3),
 nextAttemptAt INTEGER NOT NULL DEFAULT 0,
 leaseCalls INTEGER NOT NULL DEFAULT 0 CHECK(leaseCalls BETWEEN 0 AND 3),
 leaseToken TEXT,
 leaseExpiresAt INTEGER,
 failure TEXT NOT NULL DEFAULT '',
 finishedAt INTEGER,
 PRIMARY KEY(userID,jobID,id),
 UNIQUE(userID,jobID,assetID),
 UNIQUE(userID,jobID,position),
 FOREIGN KEY(userID,jobID) REFERENCES ai_jobs(userID,id) ON DELETE CASCADE
);
CREATE INDEX ai_jobs_history ON ai_jobs(userID,createdAt);
CREATE INDEX ai_job_items_due ON ai_job_items(state,nextAttemptAt);
CREATE INDEX ai_job_items_leases ON ai_job_items(state,leaseExpiresAt);

CREATE TABLE ai_analyses (
 userID TEXT NOT NULL,
 id TEXT NOT NULL,
 jobID TEXT NOT NULL,
 itemID TEXT NOT NULL,
 assetID TEXT NOT NULL,
 outcome TEXT NOT NULL CHECK(outcome IN ('located','ambiguous','unknown')),
 payload TEXT NOT NULL,
 metadata TEXT NOT NULL,
 createdAt INTEGER NOT NULL,
 PRIMARY KEY(userID,id),
 UNIQUE(userID,jobID,itemID),
 FOREIGN KEY(userID,jobID,itemID) REFERENCES ai_job_items(userID,jobID,id) ON DELETE CASCADE
);
CREATE INDEX ai_analyses_history ON ai_analyses(userID,assetID,createdAt);

-- +goose StatementBegin
CREATE TRIGGER ai_analyses_immutable BEFORE UPDATE ON ai_analyses
BEGIN
 SELECT RAISE(ABORT,'AI analyses are immutable');
END;
-- +goose StatementEnd

-- +goose Down
DROP TABLE ai_analyses;
DROP TABLE ai_job_items;
DROP TABLE ai_jobs;
