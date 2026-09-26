-- +goose Up
CREATE TABLE ai_stack_write_fields (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 operationID TEXT NOT NULL,
 assetID TEXT NOT NULL,
 field TEXT NOT NULL CHECK(field IN ('gps','description')),
 status TEXT NOT NULL CHECK(status IN ('verified','baseline','conflict','unavailable')),
 observed TEXT CHECK(length(CAST(observed AS BLOB))<=400000),
 wasVerified INTEGER NOT NULL DEFAULT 0 CHECK(wasVerified IN (0,1)),
 PRIMARY KEY(userID,installationID,operationID,assetID,field),
 FOREIGN KEY(userID,installationID,operationID,assetID) REFERENCES ai_stack_write_targets(userID,installationID,operationID,assetID) ON DELETE CASCADE
);
-- +goose StatementBegin
CREATE TRIGGER ai_stack_write_field_scope_immutable BEFORE UPDATE OF userID,installationID,operationID,assetID,field ON ai_stack_write_fields
BEGIN
 SELECT RAISE(ABORT,'Stack field scope is immutable');
END;
-- +goose StatementEnd

-- +goose Down
DROP TABLE ai_stack_write_fields;
