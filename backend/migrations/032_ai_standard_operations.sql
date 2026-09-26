-- +goose Up
CREATE TABLE ai_standard_write_operations (
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
 status TEXT NOT NULL DEFAULT 'queued' CHECK(status IN ('queued','writing','verifying','retryable','succeeded','partial','conflict','failed','canceled')),
 code TEXT NOT NULL DEFAULT '',
 PRIMARY KEY(userID,installationID,id),
 UNIQUE(userID,installationID,idempotencyKey),
 UNIQUE(userID,installationID,previewID),
 FOREIGN KEY(userID,installationID,previewID) REFERENCES ai_standard_write_previews(userID,installationID,id) DEFERRABLE INITIALLY DEFERRED,
 FOREIGN KEY(userID,installationID,draftID,revision) REFERENCES ai_draft_revisions(userID,installationID,draftID,revision) DEFERRABLE INITIALLY DEFERRED
);
CREATE TABLE ai_standard_write_targets (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 operationID TEXT NOT NULL,
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
 PRIMARY KEY(userID,installationID,operationID),
 FOREIGN KEY(userID,installationID,operationID) REFERENCES ai_standard_write_operations(userID,installationID,id) ON DELETE CASCADE
);
CREATE TABLE ai_standard_write_events (
 sequence INTEGER PRIMARY KEY AUTOINCREMENT,
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 operationID TEXT NOT NULL,
 code TEXT NOT NULL,
 at INTEGER NOT NULL,
 attempt INTEGER NOT NULL,
 FOREIGN KEY(userID,installationID,operationID) REFERENCES ai_standard_write_operations(userID,installationID,id) ON DELETE CASCADE
);
CREATE INDEX ai_standard_write_pending ON ai_standard_write_operations(installationID,status,approvedAt);
CREATE INDEX ai_standard_write_draft ON ai_standard_write_operations(userID,installationID,draftID);
-- +goose StatementBegin
CREATE TRIGGER ai_standard_write_approval_immutable BEFORE UPDATE OF userID,installationID,id,previewID,draftID,revision,assetID,idempotencyKey,payload,digest,approvedAt,credentialHash ON ai_standard_write_operations
BEGIN
 SELECT RAISE(ABORT,'Standard write approval is immutable');
END;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TRIGGER ai_standard_write_events_immutable BEFORE UPDATE ON ai_standard_write_events
BEGIN
 SELECT RAISE(ABORT,'Standard write events are immutable');
END;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TRIGGER ai_standard_write_deleted_account_guard BEFORE DELETE ON users
BEGIN
 DELETE FROM ai_write_target_guards WHERE token IN (
  SELECT o.id FROM ai_standard_write_operations o JOIN ai_standard_write_targets t ON t.userID=o.userID AND t.installationID=o.installationID AND t.operationID=o.id
  WHERE o.userID=OLD.ID AND t.senderActive=0 AND t.completionKnown=1
 );
END;
-- +goose StatementEnd
CREATE VIEW ai_all_write_operations AS
 SELECT * FROM ai_write_operations UNION ALL SELECT * FROM ai_standard_write_operations;
CREATE VIEW ai_all_write_targets AS
 SELECT * FROM ai_write_targets UNION ALL SELECT * FROM ai_standard_write_targets;

-- +goose Down
DROP VIEW ai_all_write_targets;
DROP VIEW ai_all_write_operations;
DROP TRIGGER ai_standard_write_deleted_account_guard;
DROP TABLE ai_standard_write_events;
DROP TABLE ai_standard_write_targets;
DROP TABLE ai_standard_write_operations;
