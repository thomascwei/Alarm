-- name: CreateRule :execresult
INSERT INTO rules (object, AlarmCategoryOrder, AlarmLogic, TriggerValue, AlarmCategory, AlarmMessage, AckMethod)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListAllRules :many
SELECT *
FROM rules;

-- name: UpdateRule :exec
UPDATE rules
set AlarmCategoryOrder=?,
    AlarmLogic=?,
    TriggerValue=?,
    AlarmCategory=?,
    AlarmMessage=?,
    AckMethod=?,
    created_at=?
WHERE id = ?;

-- name: DeleteRule :exec
DELETE
FROM rules
WHERE id = ?;


-- name: CreateAlarmEvent :execresult
INSERT INTO history_event (Object, AlarmCategoryOrder, HighestAlarmCategory, AckMessage, start_time)
VALUES (?, ?, ?, ?, ?);

-- name: UpgradeAlarmCategory :exec
UPDATE history_event
SET AlarmCategoryOrder   = ?,
    HighestAlarmCategory = ?
where id = ?
  and end_time is null;

-- name: UpdateAlarmAckMessage :exec
UPDATE history_event
SET AckMessage = ?
where id = ?
  and end_time is null;

-- name: SetAlarmEventEndTime :exec
UPDATE history_event
SET end_time = ?
where id = ?
  and end_time is null;

-- name: CreateAlarmEventDetail :execresult
INSERT INTO history_event_detail (Event_id, Object, AlarmCategory, created_at)
VALUES (?, ?, ?, ?);

-- name: ListAllHistoryBaseOnStartTime :many
SELECT *
FROM history_event
where start_time >= ?
  and start_time < ?
;

-- name: TruncateRules :exec
TRUNCATE rules;