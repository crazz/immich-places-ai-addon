-- +goose Up
CREATE TABLE ai_mirror_write_steps (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 operationID TEXT NOT NULL,
 assetID TEXT NOT NULL,
 status TEXT NOT NULL DEFAULT 'blocked' CHECK(status IN ('blocked','queued','writing','verifying','retryable','succeeded','conflict','failed','canceled','expired')),
 code TEXT NOT NULL DEFAULT '',
 attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 2),
 generation INTEGER NOT NULL DEFAULT 0 CHECK(generation>=0),
 retryFrom INTEGER NOT NULL DEFAULT -1,
 leaseToken TEXT NOT NULL DEFAULT '',
 leaseUntil INTEGER NOT NULL DEFAULT 0,
 completionKnown INTEGER NOT NULL DEFAULT 1 CHECK(completionKnown IN (0,1)),
 senderActive INTEGER NOT NULL DEFAULT 0 CHECK(senderActive IN (0,1)),
 reads INTEGER NOT NULL DEFAULT 0 CHECK(reads BETWEEN 0 AND 3),
 dueAt INTEGER NOT NULL DEFAULT 0,
 observed TEXT CHECK(observed IS NULL OR length(CAST(observed AS BLOB))<=131072),
 verified INTEGER NOT NULL DEFAULT 0 CHECK(verified IN (0,1)),
 noop INTEGER NOT NULL DEFAULT 0 CHECK(noop IN (0,1)),
 updatedAt INTEGER NOT NULL DEFAULT 0,
 PRIMARY KEY(userID,installationID,operationID,assetID),
 FOREIGN KEY(userID,installationID,operationID,assetID) REFERENCES ai_stack_write_targets(userID,installationID,operationID,assetID) ON DELETE CASCADE
);
CREATE INDEX ai_mirror_write_pending ON ai_mirror_write_steps(installationID,status,dueAt);
-- +goose StatementBegin
CREATE TRIGGER ai_mirror_step_approved BEFORE INSERT ON ai_mirror_write_steps
WHEN NOT EXISTS (
 SELECT 1 FROM ai_stack_write_operations o
 WHERE o.userID=NEW.userID AND o.installationID=NEW.installationID AND o.id=NEW.operationID AND o.assetID=NEW.assetID
 AND json_extract(o.payload,'$.version')='mirror-preview-v4'
 AND json_extract(o.payload,'$.targetId')=NEW.assetID
 AND json_extract(o.payload,'$.mirror.key')='immich-places-ai-addon'
)
BEGIN
 SELECT RAISE(ABORT,'Metadata step requires its exact combined approval');
END;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TRIGGER ai_mirror_step_scope_immutable BEFORE UPDATE OF userID,installationID,operationID,assetID ON ai_mirror_write_steps
BEGIN
 SELECT RAISE(ABORT,'Metadata step scope is immutable');
END;
-- +goose StatementEnd

CREATE TABLE ai_mirror_write_events (
 sequence INTEGER PRIMARY KEY AUTOINCREMENT,
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 operationID TEXT NOT NULL,
 assetID TEXT NOT NULL,
 code TEXT NOT NULL,
 at INTEGER NOT NULL,
 attempt INTEGER NOT NULL CHECK(attempt BETWEEN 0 AND 2),
 generation INTEGER NOT NULL CHECK(generation>=0),
 FOREIGN KEY(userID,installationID,operationID,assetID) REFERENCES ai_mirror_write_steps(userID,installationID,operationID,assetID) ON DELETE CASCADE
);
-- +goose StatementBegin
CREATE TRIGGER ai_mirror_events_immutable BEFORE UPDATE ON ai_mirror_write_events
BEGIN
 SELECT RAISE(ABORT,'Metadata step events are immutable');
END;
-- +goose StatementEnd

CREATE TABLE ai_mirror_records (
 userID TEXT NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
 installationID TEXT NOT NULL,
 assetID TEXT NOT NULL,
 recordID TEXT NOT NULL UNIQUE,
 lastOperationID TEXT,
 value TEXT CHECK(value IS NULL OR length(CAST(value AS BLOB))<=65536),
 verifiedAt INTEGER,
 PRIMARY KEY(userID,installationID,assetID),
 FOREIGN KEY(userID,installationID,lastOperationID,assetID) REFERENCES ai_mirror_write_steps(userID,installationID,operationID,assetID) DEFERRABLE INITIALLY DEFERRED,
 CHECK((lastOperationID IS NULL AND value IS NULL AND verifiedAt IS NULL) OR (lastOperationID IS NOT NULL AND value IS NOT NULL AND verifiedAt IS NOT NULL))
);
-- +goose StatementBegin
CREATE TRIGGER ai_mirror_record_identity_immutable BEFORE UPDATE OF userID,installationID,assetID,recordID ON ai_mirror_records
BEGIN
 SELECT RAISE(ABORT,'Mirror record identity is immutable');
END;
-- +goose StatementEnd

-- +goose Down
CREATE TABLE ai_mirror_step_rollback_guard (retained INTEGER CHECK(retained=0));
INSERT INTO ai_mirror_step_rollback_guard SELECT count(*) FROM ai_mirror_write_steps;
INSERT INTO ai_mirror_step_rollback_guard SELECT count(*) FROM ai_mirror_records;
DROP TABLE ai_mirror_step_rollback_guard;
DROP TABLE ai_mirror_records;
DROP TABLE ai_mirror_write_events;
DROP TABLE ai_mirror_write_steps;
