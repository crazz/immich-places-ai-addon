-- +goose Up
DROP TRIGGER ai_stack_write_deleted_account_guard;
-- +goose StatementBegin
CREATE TRIGGER ai_stack_write_deleted_account_guard BEFORE DELETE ON users
BEGIN
 DELETE FROM ai_write_target_guards WHERE token IN (
  SELECT t.guardToken FROM ai_stack_write_targets t
  WHERE t.userID=OLD.ID AND t.senderActive=0 AND t.completionKnown=1
   AND NOT EXISTS (SELECT 1 FROM ai_mirror_write_steps m
    WHERE m.userID=t.userID AND m.installationID=t.installationID
     AND m.operationID=t.operationID AND m.assetID=t.assetID
     AND (m.senderActive=1 OR m.completionKnown=0))
 );
END;
-- +goose StatementEnd

DROP VIEW ai_stack_write_status;
CREATE VIEW ai_stack_write_status AS
 WITH standard AS (
  SELECT userID,installationID,operationID,
   CASE WHEN sum(status='writing')>0 THEN 'writing'
    WHEN sum(status='verifying')>0 THEN 'verifying'
    WHEN sum(status='queued')>0 THEN 'queued'
    WHEN count(DISTINCT status)=1 THEN min(status) ELSE 'partial' END AS status,
   min(verified AND completionKnown AND NOT senderActive) AS complete
  FROM ai_stack_write_targets GROUP BY userID,installationID,operationID
 )
 SELECT s.userID,s.installationID,s.operationID,
  CASE
   WHEN m.operationID IS NULL OR s.status IN ('writing','verifying','queued') THEN s.status
   WHEN m.status IN ('writing','verifying','queued') THEN m.status
   WHEN m.status='blocked' AND t.status='succeeded' AND t.verified=1 AND t.completionKnown=1 AND t.senderActive=0 THEN 'queued'
   WHEN m.status='succeeded' AND m.verified=1 AND m.completionKnown=1 AND m.senderActive=0 AND s.complete=1 THEN s.status
   ELSE 'partial'
  END AS status
 FROM standard s LEFT JOIN ai_mirror_write_steps m
  ON m.userID=s.userID AND m.installationID=s.installationID AND m.operationID=s.operationID
 LEFT JOIN ai_stack_write_targets t ON t.userID=m.userID AND t.installationID=m.installationID AND t.operationID=m.operationID AND t.assetID=m.assetID;

-- +goose Down
CREATE TEMP TABLE ai_mirror_lifecycle_rollback_guard (empty INTEGER CHECK(empty=0));
INSERT INTO ai_mirror_lifecycle_rollback_guard SELECT count(*) FROM ai_mirror_write_steps;
DROP TABLE ai_mirror_lifecycle_rollback_guard;
DROP VIEW ai_stack_write_status;
CREATE VIEW ai_stack_write_status AS
 SELECT userID,installationID,operationID,
 CASE
  WHEN sum(status='writing')>0 THEN 'writing'
  WHEN sum(status='verifying')>0 THEN 'verifying'
  WHEN sum(status='queued')>0 THEN 'queued'
  WHEN count(DISTINCT status)=1 THEN min(status)
  ELSE 'partial'
 END AS status
 FROM ai_stack_write_targets GROUP BY userID,installationID,operationID;

DROP TRIGGER ai_stack_write_deleted_account_guard;
-- +goose StatementBegin
CREATE TRIGGER ai_stack_write_deleted_account_guard BEFORE DELETE ON users
BEGIN
 DELETE FROM ai_write_target_guards WHERE token IN (
  SELECT guardToken FROM ai_stack_write_targets
  WHERE userID=OLD.ID AND senderActive=0 AND completionKnown=1
 );
END;
-- +goose StatementEnd
