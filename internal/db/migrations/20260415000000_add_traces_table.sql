-- +goose Up
CREATE TABLE IF NOT EXISTS traces (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    started_at INTEGER NOT NULL,
    stopped_at INTEGER NOT NULL,
    event_count INTEGER NOT NULL DEFAULT 0,
    data_jsonl TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_traces_session_id ON traces(session_id);
CREATE INDEX IF NOT EXISTS idx_traces_created_at ON traces(created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_traces_created_at;
DROP INDEX IF EXISTS idx_traces_session_id;
DROP TABLE IF EXISTS traces;
