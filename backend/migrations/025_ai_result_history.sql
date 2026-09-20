-- +goose Up
CREATE TABLE ai_result_history (
 sequence INTEGER PRIMARY KEY AUTOINCREMENT,
 userID TEXT NOT NULL,
 installationID TEXT NOT NULL,
 jobID TEXT NOT NULL,
 itemID TEXT NOT NULL,
 assetID TEXT NOT NULL,
 terminalAt INTEGER NOT NULL,
 executionState TEXT NOT NULL CHECK(executionState IN ('succeeded','failed','canceled')),
 captureDay TEXT,
 albumID TEXT,
 label TEXT,
 UNIQUE(userID,jobID,itemID),
 FOREIGN KEY(userID,jobID,itemID) REFERENCES ai_job_items(userID,jobID,id) ON DELETE CASCADE
);
CREATE INDEX ai_result_history_page ON ai_result_history(userID,installationID,terminalAt DESC,jobID DESC,itemID DESC);
CREATE INDEX ai_result_history_date ON ai_result_history(userID,installationID,captureDay);
CREATE INDEX ai_result_history_album ON ai_result_history(userID,installationID,albumID);
CREATE INDEX ai_result_history_asset ON ai_result_history(userID,installationID,assetID);

INSERT INTO ai_result_history(userID,installationID,jobID,itemID,assetID,terminalAt,executionState,captureDay,albumID,label)
SELECT i.userID,j.installationID,i.jobID,i.id,i.assetID,COALESCE(i.finishedAt,j.createdAt),i.state,
 CASE WHEN date(substr(l.captureTime,1,10),'+0 days')=substr(l.captureTime,1,10) THEN substr(l.captureTime,1,10) END,l.albumID,
 (SELECT substr(json_extract(CASE WHEN json_valid(d.value) THEN d.value ELSE '{}' END,'$.text'),1,200) FROM json_each(CASE WHEN length(CAST(a.payload AS BLOB))<=1048576 AND json_valid(a.payload) THEN a.payload ELSE '{}' END,'$.descriptions') d
  WHERE json_extract(CASE WHEN json_valid(d.value) THEN d.value ELSE '{}' END,'$.language')=json_extract(CASE WHEN length(CAST(j.requestJSON AS BLOB))<=131072 AND json_valid(j.requestJSON) THEN j.requestJSON ELSE '{}' END,'$.PrimaryLanguage') AND json_type(CASE WHEN json_valid(d.value) THEN d.value ELSE '{}' END,'$.text')='text' LIMIT 1)
FROM ai_job_items i JOIN ai_jobs j ON j.userID=i.userID AND j.id=i.jobID
LEFT JOIN ai_job_launch l ON l.userID=i.userID AND l.jobID=i.jobID AND l.assetID=i.assetID
LEFT JOIN ai_analyses a ON a.userID=i.userID AND a.jobID=i.jobID AND a.itemID=i.id
WHERE i.state IN ('succeeded','failed','canceled') ORDER BY COALESCE(i.finishedAt,j.createdAt),i.jobID,i.id;

-- +goose StatementBegin
CREATE TRIGGER ai_result_history_terminal AFTER UPDATE OF state ON ai_job_items
WHEN NEW.state IN ('succeeded','failed','canceled') AND OLD.state NOT IN ('succeeded','failed','canceled')
BEGIN
 INSERT INTO ai_result_history(userID,installationID,jobID,itemID,assetID,terminalAt,executionState,captureDay,albumID,label)
 SELECT NEW.userID,j.installationID,NEW.jobID,NEW.id,NEW.assetID,COALESCE(NEW.finishedAt,j.createdAt),NEW.state,
  CASE WHEN date(substr(l.captureTime,1,10),'+0 days')=substr(l.captureTime,1,10) THEN substr(l.captureTime,1,10) END,l.albumID,
  (SELECT substr(json_extract(CASE WHEN json_valid(d.value) THEN d.value ELSE '{}' END,'$.text'),1,200) FROM json_each(CASE WHEN length(CAST(a.payload AS BLOB))<=1048576 AND json_valid(a.payload) THEN a.payload ELSE '{}' END,'$.descriptions') d
   WHERE json_extract(CASE WHEN json_valid(d.value) THEN d.value ELSE '{}' END,'$.language')=json_extract(CASE WHEN length(CAST(j.requestJSON AS BLOB))<=131072 AND json_valid(j.requestJSON) THEN j.requestJSON ELSE '{}' END,'$.PrimaryLanguage') AND json_type(CASE WHEN json_valid(d.value) THEN d.value ELSE '{}' END,'$.text')='text' LIMIT 1)
 FROM ai_jobs j LEFT JOIN ai_job_launch l ON l.userID=j.userID AND l.jobID=j.id AND l.assetID=NEW.assetID
 LEFT JOIN ai_analyses a ON a.userID=j.userID AND a.jobID=j.id AND a.itemID=NEW.id
 WHERE j.userID=NEW.userID AND j.id=NEW.jobID;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER ai_result_history_immutable BEFORE UPDATE ON ai_result_history
BEGIN
 SELECT RAISE(ABORT,'AI terminal history is immutable');
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER ai_result_history_terminal;
DROP TABLE ai_result_history;
