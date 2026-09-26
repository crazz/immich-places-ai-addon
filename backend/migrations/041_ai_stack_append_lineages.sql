-- +goose Up
CREATE TABLE ai_stack_append_lineages (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 assetID TEXT NOT NULL,
 operationID TEXT NOT NULL,
 lineageID TEXT NOT NULL,
 block TEXT NOT NULL CHECK(length(CAST(block AS BLOB))<=20480),
 hash TEXT NOT NULL CHECK(length(hash)=64),
 PRIMARY KEY(userID,installationID,assetID),
 FOREIGN KEY(userID,installationID,operationID,assetID) REFERENCES ai_stack_write_targets(userID,installationID,operationID,assetID) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE ai_stack_append_lineages;
