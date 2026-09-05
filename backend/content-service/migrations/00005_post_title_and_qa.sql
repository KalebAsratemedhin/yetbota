-- +goose Up
-- Local/dev wipe so title can be NOT NULL without backfill.
TRUNCATE TABLE post_interaction_events, post_bookmarks, comment_reactions, post_comments,
    post_reactions, post_hashtags, post_media, posts CASCADE;

ALTER TABLE posts ADD COLUMN IF NOT EXISTS title TEXT NOT NULL;

CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY,
    author_id UUID NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    post_id UUID REFERENCES posts(id) ON DELETE SET NULL,
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    upvote_count BIGINT NOT NULL DEFAULT 0,
    downvote_count BIGINT NOT NULL DEFAULT 0,
    answer_count BIGINT NOT NULL DEFAULT 0,
    closed_at TIMESTAMPTZ,
    closed_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ((lat IS NULL AND lng IS NULL) OR (lat IS NOT NULL AND lng IS NOT NULL)),
    CHECK ((closed_at IS NULL AND closed_by IS NULL) OR (closed_at IS NOT NULL AND closed_by IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS questions_created_idx ON questions (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS questions_post_created_idx ON questions (post_id, created_at DESC, id DESC)
    WHERE post_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS questions_author_created_idx ON questions (author_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS question_votes (
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    vote TEXT NOT NULL CHECK (vote IN ('up', 'down')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (question_id, user_id)
);

CREATE TABLE IF NOT EXISTS answers (
    id UUID PRIMARY KEY,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    author_id UUID NOT NULL,
    body TEXT NOT NULL,
    upvote_count BIGINT NOT NULL DEFAULT 0,
    downvote_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS answers_question_created_idx ON answers (question_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS answer_votes (
    answer_id UUID NOT NULL REFERENCES answers(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    vote TEXT NOT NULL CHECK (vote IN ('up', 'down')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (answer_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS answer_votes;
DROP TABLE IF EXISTS answers;
DROP TABLE IF EXISTS question_votes;
DROP TABLE IF EXISTS questions;
ALTER TABLE posts DROP COLUMN IF EXISTS title;
