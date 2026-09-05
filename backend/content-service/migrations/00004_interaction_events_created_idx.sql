-- +goose Up
CREATE INDEX IF NOT EXISTS post_interaction_events_created_idx
    ON post_interaction_events (created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS post_interaction_events_created_idx;
