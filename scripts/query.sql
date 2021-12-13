-- name: CreateRule :execresult
INSERT INTO rules (object, AlarmCategoryOrder, AlarmLogic, TriggerValue, AlarmCategory, AlamrMessage)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListAllRules :many
SELECT *
FROM rules;

-- name: UpdateRule :exec
UPDATE rules
set AlarmCategoryOrder=?, 
    AlarmLogic=?, 
    TriggerValue=?,
    AlarmCategory=?,
    AlamrMessage=?,
    created_at=?
WHERE id = ?;

-- name: DeleteRule :exec
DELETE
FROM rules
WHERE id = ?;


-- name: CreateHistory :execresult
INSERT INTO history (eventID, Object, AlarmCategory, AckMessage)
VALUES (?, ?, ?, ?);

-- name: ListAllHistory :many
SELECT *
FROM history
where created_at>=? and created_at<?
;

-- name: TruncateRules :exec
TRUNCATE rules;