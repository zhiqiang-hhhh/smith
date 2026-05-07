-- name: CreateTrace :one
INSERT INTO traces (
    id,
    session_id,
    started_at,
    stopped_at,
    event_count,
    data_jsonl,
    created_at
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    strftime('%s', 'now')
) RETURNING *;

-- name: GetTraceByID :one
SELECT *
FROM traces
WHERE id = ?
LIMIT 1;

-- name: ListTracesBySession :many
SELECT *
FROM traces
WHERE session_id = ?
ORDER BY created_at DESC;
