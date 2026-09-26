-- +goose Up
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
DROP VIEW ai_all_write_operations;
CREATE VIEW ai_all_write_operations AS
 SELECT * FROM ai_write_operations
 UNION ALL SELECT * FROM ai_standard_write_operations
 UNION ALL
 SELECT o.userID,o.installationID,o.id,o.previewID,o.draftID,o.revision,o.assetID,o.idempotencyKey,o.payload,o.digest,o.approvedAt,o.credentialHash,s.status,o.code
 FROM ai_stack_write_operations o JOIN ai_stack_write_status s
 ON s.userID=o.userID AND s.installationID=o.installationID AND s.operationID=o.id;

-- +goose Down
DROP VIEW ai_all_write_operations;
CREATE VIEW ai_all_write_operations AS
 SELECT * FROM ai_write_operations UNION ALL SELECT * FROM ai_standard_write_operations;
DROP VIEW ai_stack_write_status;
