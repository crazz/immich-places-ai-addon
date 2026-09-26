-- +goose Up
CREATE TABLE ai_draft_description_baselines (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 draftID TEXT NOT NULL,
 revision INTEGER NOT NULL,
 presence TEXT NOT NULL CHECK(presence IN ('absent','null','value')),
 value TEXT NOT NULL CHECK(length(CAST(value AS BLOB)) <= 65536),
 CHECK(presence='value' OR value=''),
 PRIMARY KEY(userID,installationID,draftID,revision),
 FOREIGN KEY(userID,installationID,draftID,revision) REFERENCES ai_draft_revisions(userID,installationID,draftID,revision) ON DELETE CASCADE
);
-- +goose StatementBegin
CREATE TRIGGER ai_draft_description_baselines_immutable BEFORE UPDATE ON ai_draft_description_baselines
BEGIN
 SELECT RAISE(ABORT,'Description baselines are immutable');
END;
-- +goose StatementEnd

CREATE TABLE ai_draft_description_observations (
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 draftID TEXT NOT NULL,
 observationID TEXT NOT NULL,
 presence TEXT NOT NULL CHECK(presence IN ('absent','null','value')),
 value TEXT NOT NULL CHECK(length(CAST(value AS BLOB)) <= 65536),
 CHECK(presence='value' OR value=''),
 PRIMARY KEY(userID,installationID,draftID,observationID),
 FOREIGN KEY(userID,installationID,draftID,observationID) REFERENCES ai_draft_baseline_observations(userID,installationID,draftID,id) ON DELETE CASCADE
);
-- +goose StatementBegin
CREATE TRIGGER ai_draft_description_observations_immutable BEFORE UPDATE ON ai_draft_description_observations
BEGIN
 SELECT RAISE(ABORT,'Description observations are immutable');
END;
-- +goose StatementEnd

-- +goose Down
DROP TABLE ai_draft_description_observations;
DROP TABLE ai_draft_description_baselines;
