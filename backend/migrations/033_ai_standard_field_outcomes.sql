-- +goose Up
CREATE TABLE ai_standard_write_fields (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 operationID TEXT NOT NULL,
 field TEXT NOT NULL CHECK(field IN ('gps','description')),
 status TEXT NOT NULL CHECK(status IN ('verified','baseline','conflict','unavailable')),
 observed TEXT CHECK(length(CAST(observed AS BLOB))<=400000),
 PRIMARY KEY(userID,installationID,operationID,field),
 FOREIGN KEY(userID,installationID,operationID) REFERENCES ai_standard_write_operations(userID,installationID,id) ON DELETE CASCADE
);
-- +goose StatementBegin
CREATE TRIGGER ai_standard_write_field_scope_immutable BEFORE UPDATE OF userID,installationID,operationID,field ON ai_standard_write_fields
BEGIN
 SELECT RAISE(ABORT,'Standard field scope is immutable');
END;
-- +goose StatementEnd

-- +goose Down
DROP TABLE ai_standard_write_fields;
