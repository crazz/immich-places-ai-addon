-- +goose Up
ALTER TABLE ai_provider_capability_checks ADD COLUMN usageJSON TEXT;
ALTER TABLE ai_provider_capability_checks ADD COLUMN inputMayBeConsumed INTEGER NOT NULL DEFAULT 0 CHECK (inputMayBeConsumed IN (0, 1));
-- Earlier records do not distinguish pre-dispatch failures from sent input.
UPDATE ai_provider_capability_checks SET inputMayBeConsumed = 1;

-- +goose Down
ALTER TABLE ai_provider_capability_checks DROP COLUMN inputMayBeConsumed;
ALTER TABLE ai_provider_capability_checks DROP COLUMN usageJSON;
