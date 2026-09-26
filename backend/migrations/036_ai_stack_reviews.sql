-- +goose Up
CREATE TABLE ai_stack_reviews (
 userID TEXT NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
 installationID TEXT NOT NULL,
 draftID TEXT NOT NULL,
 revision INTEGER NOT NULL CHECK(revision > 0),
 id TEXT NOT NULL,
 expiresAt INTEGER NOT NULL,
 content TEXT NOT NULL CHECK(length(CAST(content AS BLOB)) <= 1048576),
 PRIMARY KEY(userID,installationID,draftID,id),
 FOREIGN KEY(userID,installationID,draftID,revision) REFERENCES ai_draft_revisions(userID,installationID,draftID,revision) ON DELETE CASCADE
);
CREATE INDEX ai_stack_review_expiry ON ai_stack_reviews(userID,installationID,draftID,expiresAt);
-- +goose StatementBegin
CREATE TRIGGER ai_stack_reviews_immutable BEFORE UPDATE ON ai_stack_reviews
BEGIN
 SELECT RAISE(ABORT,'Stack reviews are immutable');
END;
-- +goose StatementEnd

-- +goose Down
DROP TABLE ai_stack_reviews;
