-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS reputation_score BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS reputation_events (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id UUID,
    event_type TEXT NOT NULL,
    subject_type TEXT NOT NULL DEFAULT '',
    subject_id UUID,
    delta INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS reputation_events_user_created_idx
    ON reputation_events (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS reputation_events_user_type_idx
    ON reputation_events (user_id, event_type);

CREATE TABLE IF NOT EXISTS badges (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_badges (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_id TEXT NOT NULL REFERENCES badges(id) ON DELETE CASCADE,
    awarded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, badge_id)
);

INSERT INTO badges (id, name, description) VALUES
    ('first_post', 'First Post', 'Published your first post'),
    ('ten_likes_received', 'Ten Likes', 'Received 10 likes on your posts'),
    ('first_follower', 'First Follower', 'Gained your first follower'),
    ('level_contributor', 'Contributor', 'Reached Contributor reputation level')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS user_badges;
DROP TABLE IF EXISTS badges;
DROP INDEX IF EXISTS reputation_events_user_type_idx;
DROP INDEX IF EXISTS reputation_events_user_created_idx;
DROP TABLE IF EXISTS reputation_events;
ALTER TABLE users DROP COLUMN IF EXISTS reputation_score;
