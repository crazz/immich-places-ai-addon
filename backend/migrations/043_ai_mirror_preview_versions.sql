-- +goose Up
PRAGMA defer_foreign_keys=ON;
DROP VIEW ai_all_write_previews;
CREATE TABLE ai_stack_preview_upgrade AS SELECT * FROM ai_stack_write_previews;
DROP TABLE ai_stack_write_previews;
CREATE TABLE ai_stack_write_previews (
 userID TEXT NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
 installationID TEXT NOT NULL,
 id TEXT NOT NULL,
 draftID TEXT NOT NULL,
 revision INTEGER NOT NULL CHECK(revision > 0),
 version TEXT NOT NULL CHECK(version IN ('stack-preview-v3','mirror-preview-v4')),
 payload TEXT NOT NULL CHECK(length(CAST(payload AS BLOB)) <= 1048576),
 digest TEXT NOT NULL CHECK(length(digest) = 64),
 createdAt INTEGER NOT NULL,
 expiresAt INTEGER NOT NULL,
 invalidated INTEGER NOT NULL DEFAULT 0 CHECK(invalidated IN (0,1)),
 protected INTEGER NOT NULL DEFAULT 0 CHECK(protected IN (0,1)),
 PRIMARY KEY(userID,installationID,id),
 FOREIGN KEY(userID,installationID,draftID,revision) REFERENCES ai_draft_revisions(userID,installationID,draftID,revision) DEFERRABLE INITIALLY DEFERRED,
 CHECK(expiresAt > createdAt)
);
INSERT INTO ai_stack_write_previews SELECT * FROM ai_stack_preview_upgrade;
DROP TABLE ai_stack_preview_upgrade;
CREATE INDEX ai_stack_preview_draft ON ai_stack_write_previews(userID,installationID,draftID,expiresAt);
CREATE INDEX ai_stack_preview_expiry ON ai_stack_write_previews(expiresAt);
-- +goose StatementBegin
CREATE TRIGGER ai_stack_previews_immutable BEFORE UPDATE OF userID,installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt ON ai_stack_write_previews
BEGIN
 SELECT RAISE(ABORT,'Stack previews are immutable');
END;
-- +goose StatementEnd
CREATE VIEW ai_all_write_previews AS
 SELECT * FROM ai_write_previews UNION ALL SELECT * FROM ai_standard_write_previews UNION ALL SELECT * FROM ai_stack_write_previews;

-- +goose Down
CREATE TABLE ai_mirror_rollback_guard (retained INTEGER CHECK(retained=0));
INSERT INTO ai_mirror_rollback_guard SELECT count(*) FROM ai_stack_write_previews WHERE version='mirror-preview-v4';
DROP TABLE ai_mirror_rollback_guard;
PRAGMA defer_foreign_keys=ON;
DROP VIEW ai_all_write_previews;
CREATE TABLE ai_stack_preview_upgrade AS SELECT * FROM ai_stack_write_previews;
DROP TABLE ai_stack_write_previews;
CREATE TABLE ai_stack_write_previews (
 userID TEXT NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
 installationID TEXT NOT NULL,
 id TEXT NOT NULL,
 draftID TEXT NOT NULL,
 revision INTEGER NOT NULL CHECK(revision > 0),
 version TEXT NOT NULL CHECK(version = 'stack-preview-v3'),
 payload TEXT NOT NULL CHECK(length(CAST(payload AS BLOB)) <= 1048576),
 digest TEXT NOT NULL CHECK(length(digest) = 64),
 createdAt INTEGER NOT NULL,
 expiresAt INTEGER NOT NULL,
 invalidated INTEGER NOT NULL DEFAULT 0 CHECK(invalidated IN (0,1)),
 protected INTEGER NOT NULL DEFAULT 0 CHECK(protected IN (0,1)),
 PRIMARY KEY(userID,installationID,id),
 FOREIGN KEY(userID,installationID,draftID,revision) REFERENCES ai_draft_revisions(userID,installationID,draftID,revision) DEFERRABLE INITIALLY DEFERRED,
 CHECK(expiresAt > createdAt)
);
INSERT INTO ai_stack_write_previews SELECT * FROM ai_stack_preview_upgrade;
DROP TABLE ai_stack_preview_upgrade;
CREATE INDEX ai_stack_preview_draft ON ai_stack_write_previews(userID,installationID,draftID,expiresAt);
CREATE INDEX ai_stack_preview_expiry ON ai_stack_write_previews(expiresAt);
-- +goose StatementBegin
CREATE TRIGGER ai_stack_previews_immutable BEFORE UPDATE OF userID,installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt ON ai_stack_write_previews
BEGIN
 SELECT RAISE(ABORT,'Stack previews are immutable');
END;
-- +goose StatementEnd
CREATE VIEW ai_all_write_previews AS
 SELECT * FROM ai_write_previews UNION ALL SELECT * FROM ai_standard_write_previews UNION ALL SELECT * FROM ai_stack_write_previews;
