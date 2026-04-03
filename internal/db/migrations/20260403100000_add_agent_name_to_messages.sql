-- +goose Up
ALTER TABLE messages ADD COLUMN agent_name TEXT DEFAULT '' NOT NULL;

-- +goose Down
ALTER TABLE messages DROP COLUMN agent_name;
