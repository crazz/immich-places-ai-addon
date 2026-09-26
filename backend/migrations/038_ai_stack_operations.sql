-- +goose Up
CREATE TABLE ai_stack_write_operations (
 userID TEXT NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
 installationID TEXT NOT NULL,
 id TEXT NOT NULL,
 previewID TEXT NOT NULL,
 draftID TEXT NOT NULL,
 revision INTEGER NOT NULL,
 assetID TEXT NOT NULL,
 idempotencyKey TEXT NOT NULL,
 payload TEXT NOT NULL CHECK(length(CAST(payload AS BLOB))<=1048576),
 digest TEXT NOT NULL CHECK(length(digest)=64),
 approvedAt INTEGER NOT NULL,
 credentialHash TEXT NOT NULL,
 status TEXT NOT NULL DEFAULT 'queued' CHECK(status IN ('queued','writing','verifying','retryable','succeeded','partial','conflict','failed','canceled','expired')),
 code TEXT NOT NULL DEFAULT '',
 PRIMARY KEY(userID,installationID,id),
 UNIQUE(userID,installationID,idempotencyKey),
 UNIQUE(userID,installationID,previewID),
 FOREIGN KEY(userID,installationID,previewID) REFERENCES ai_stack_write_previews(userID,installationID,id) DEFERRABLE INITIALLY DEFERRED,
 FOREIGN KEY(userID,installationID,draftID,revision) REFERENCES ai_draft_revisions(userID,installationID,draftID,revision) DEFERRABLE INITIALLY DEFERRED
);
CREATE TABLE ai_stack_write_targets (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 operationID TEXT NOT NULL,
 assetID TEXT NOT NULL,
 ordinal INTEGER NOT NULL CHECK(ordinal BETWEEN 0 AND 49),
 guardToken TEXT NOT NULL UNIQUE,
 status TEXT NOT NULL DEFAULT 'queued' CHECK(status IN ('queued','writing','verifying','retryable','succeeded','partial','conflict','failed','canceled','expired')),
 code TEXT NOT NULL DEFAULT '',
 attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 2),
 generation INTEGER NOT NULL DEFAULT 0,
 retryFrom INTEGER NOT NULL DEFAULT -1,
 leaseToken TEXT NOT NULL DEFAULT '',
 leaseUntil INTEGER NOT NULL DEFAULT 0,
 completionKnown INTEGER NOT NULL DEFAULT 1 CHECK(completionKnown IN (0,1)),
 senderActive INTEGER NOT NULL DEFAULT 0 CHECK(senderActive IN (0,1)),
 reads INTEGER NOT NULL DEFAULT 0 CHECK(reads BETWEEN 0 AND 3),
 dueAt INTEGER NOT NULL DEFAULT 0,
 observed TEXT,
 verified INTEGER NOT NULL DEFAULT 0 CHECK(verified IN (0,1)),
 refreshed INTEGER NOT NULL DEFAULT 0 CHECK(refreshed IN (0,1)),
 noop INTEGER NOT NULL DEFAULT 0 CHECK(noop IN (0,1)),
 PRIMARY KEY(userID,installationID,operationID,assetID),
 UNIQUE(userID,installationID,operationID,ordinal),
 FOREIGN KEY(userID,installationID,operationID) REFERENCES ai_stack_write_operations(userID,installationID,id) ON DELETE CASCADE
);
CREATE TABLE ai_stack_write_events (
 sequence INTEGER PRIMARY KEY AUTOINCREMENT,
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 operationID TEXT NOT NULL,
 assetID TEXT NOT NULL,
 code TEXT NOT NULL,
 at INTEGER NOT NULL,
 attempt INTEGER NOT NULL,
 FOREIGN KEY(userID,installationID,operationID,assetID) REFERENCES ai_stack_write_targets(userID,installationID,operationID,assetID) ON DELETE CASCADE
);
CREATE INDEX ai_stack_write_draft ON ai_stack_write_operations(userID,installationID,draftID);
CREATE INDEX ai_stack_write_pending ON ai_stack_write_targets(installationID,status,dueAt);
-- +goose StatementBegin
CREATE TRIGGER ai_stack_write_approval_immutable BEFORE UPDATE OF userID,installationID,id,previewID,draftID,revision,assetID,idempotencyKey,payload,digest,approvedAt,credentialHash ON ai_stack_write_operations
BEGIN
 SELECT RAISE(ABORT,'Stack write approval is immutable');
END;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TRIGGER ai_stack_write_scope_immutable BEFORE UPDATE OF userID,installationID,operationID,assetID,ordinal,guardToken ON ai_stack_write_targets
BEGIN
 SELECT RAISE(ABORT,'Stack write target scope is immutable');
END;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TRIGGER ai_stack_write_events_immutable BEFORE UPDATE ON ai_stack_write_events
BEGIN
 SELECT RAISE(ABORT,'Stack write events are immutable');
END;
-- +goose StatementEnd

-- +goose Down
DROP TABLE ai_stack_write_events;
DROP TABLE ai_stack_write_targets;
DROP TABLE ai_stack_write_operations;
