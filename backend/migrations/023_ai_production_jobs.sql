-- +goose Up
CREATE TABLE ai_job_admissions (
 userID TEXT NOT NULL,
 jobID TEXT NOT NULL,
 version TEXT NOT NULL CHECK(version='production-v1'),
 requestDigest TEXT NOT NULL,
 requestJSON TEXT NOT NULL,
 selectionJSON TEXT NOT NULL,
 policyID TEXT NOT NULL,
 policyJSON TEXT NOT NULL,
 PRIMARY KEY(userID,jobID),
 FOREIGN KEY(userID,jobID) REFERENCES ai_jobs(userID,id) ON DELETE CASCADE
);

CREATE TABLE ai_job_usage (
 userID TEXT NOT NULL,
 jobID TEXT NOT NULL,
 itemID TEXT NOT NULL,
 leaseToken TEXT NOT NULL,
 policyID TEXT NOT NULL,
 inputReserved INTEGER NOT NULL CHECK(inputReserved>0),
 outputReserved INTEGER NOT NULL CHECK(outputReserved>0),
 estimatedMicros INTEGER,
 inputReported INTEGER,
 outputReported INTEGER,
 totalReported INTEGER,
 createdAt INTEGER NOT NULL,
 PRIMARY KEY(userID,jobID,itemID,leaseToken),
 FOREIGN KEY(userID,jobID,itemID) REFERENCES ai_job_items(userID,jobID,id) ON DELETE CASCADE
);

CREATE TABLE ai_execution_policy_violations (
 userID TEXT NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
 policyID TEXT NOT NULL,
 createdAt INTEGER NOT NULL,
 PRIMARY KEY(userID,policyID)
);

CREATE TABLE ai_job_launch (
 userID TEXT NOT NULL,
 jobID TEXT NOT NULL,
 assetID TEXT NOT NULL,
 captureTime TEXT,
 albumID TEXT,
 albumLabel TEXT,
 PRIMARY KEY(userID,jobID,assetID),
 FOREIGN KEY(userID,jobID) REFERENCES ai_jobs(userID,id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE ai_job_launch;
DROP TABLE ai_execution_policy_violations;
DROP TABLE ai_job_usage;
DROP TABLE ai_job_admissions;
