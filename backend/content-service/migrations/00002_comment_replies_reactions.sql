-- +goose Up
ALTER TABLE post_comments
    ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES post_comments(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS like_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS dislike_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reply_count BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS post_comments_post_top_created_idx
    ON post_comments (post_id, created_at DESC, id DESC)
    WHERE parent_id IS NULL;

CREATE INDEX IF NOT EXISTS post_comments_parent_created_idx
    ON post_comments (parent_id, created_at DESC, id DESC)
    WHERE parent_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS comment_reactions (
    comment_id UUID NOT NULL REFERENCES post_comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    reaction TEXT NOT NULL CHECK (reaction IN ('like', 'dislike')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (comment_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS comment_reactions;
DROP INDEX IF EXISTS post_comments_parent_created_idx;
DROP INDEX IF EXISTS post_comments_post_top_created_idx;
ALTER TABLE post_comments
    DROP COLUMN IF EXISTS reply_count,
    DROP COLUMN IF EXISTS dislike_count,
    DROP COLUMN IF EXISTS like_count,
    DROP COLUMN IF EXISTS parent_id;
