-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER ai_stack_write_deleted_account_guard BEFORE DELETE ON users
BEGIN
 DELETE FROM ai_write_target_guards WHERE token IN (
  SELECT guardToken FROM ai_stack_write_targets
  WHERE userID=OLD.ID AND senderActive=0 AND completionKnown=1
 );
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER ai_stack_write_deleted_account_guard;
