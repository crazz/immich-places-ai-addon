-- +goose Up
ALTER TABLE ai_standard_write_fields ADD COLUMN wasVerified INTEGER NOT NULL DEFAULT 0 CHECK(wasVerified IN (0,1));
UPDATE ai_standard_write_fields SET wasVerified=1 WHERE status='verified';

-- +goose Down
ALTER TABLE ai_standard_write_fields DROP COLUMN wasVerified;
