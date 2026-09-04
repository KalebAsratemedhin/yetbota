-- +goose Up
CREATE TABLE IF NOT EXISTS post_interaction_events (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL,
    actor_id UUID NOT NULL,
    kind TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS post_interaction_events_post_created_idx
    ON post_interaction_events (post_id, created_at DESC);

CREATE INDEX IF NOT EXISTS post_interaction_events_author_created_idx
    ON post_interaction_events (author_id, created_at DESC);

CREATE INDEX IF NOT EXISTS post_interaction_events_author_kind_idx
    ON post_interaction_events (author_id, kind);

-- +goose Down
DROP INDEX IF EXISTS post_interaction_events_author_kind_idx;
DROP INDEX IF EXISTS post_interaction_events_author_created_idx;
DROP INDEX IF EXISTS post_interaction_events_post_created_idx;
DROP TABLE IF EXISTS post_interaction_events;
