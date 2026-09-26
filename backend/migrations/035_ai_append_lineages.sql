-- +goose Up
CREATE TABLE ai_write_append_lineages (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 assetID TEXT NOT NULL,
 operationID TEXT NOT NULL,
 lineageID TEXT NOT NULL,
 block TEXT NOT NULL CHECK(length(CAST(block AS BLOB))<=20480),
 hash TEXT NOT NULL CHECK(length(hash)=64),
 PRIMARY KEY(userID,installationID,assetID),
 FOREIGN KEY(userID,installationID,operationID) REFERENCES ai_standard_write_operations(userID,installationID,id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE ai_write_append_lineages;
