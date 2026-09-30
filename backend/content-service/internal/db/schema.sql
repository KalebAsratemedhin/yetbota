CREATE TABLE posts (
    id UUID PRIMARY KEY,
    author_id UUID NOT NULL,
    title TEXT NOT NULL,
    caption TEXT NOT NULL DEFAULT '',
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    media_mode TEXT NOT NULL,
    like_count BIGINT NOT NULL DEFAULT 0,
    dislike_count BIGINT NOT NULL DEFAULT 0,
    comment_count BIGINT NOT NULL DEFAULT 0,
    share_count BIGINT NOT NULL DEFAULT 0,
    bookmark_count BIGINT NOT NULL DEFAULT 0,
    -- search_vector lives in migrations only; omit from sqlc models (pgx can't scan tsvector into interface{}).
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE post_media (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    cloudinary_public_id TEXT NOT NULL,
    url TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    duration_ms INT NOT NULL DEFAULT 0,
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    mime TEXT NOT NULL DEFAULT ''
);

CREATE TABLE hashtags (
    id UUID PRIMARY KEY,
    tag TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE post_hashtags (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    hashtag_id UUID NOT NULL REFERENCES hashtags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, hashtag_id)
);

CREATE TABLE post_reactions (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    reaction TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

CREATE TABLE post_comments (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL,
    parent_id UUID REFERENCES post_comments(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    like_count BIGINT NOT NULL DEFAULT 0,
    dislike_count BIGINT NOT NULL DEFAULT 0,
    reply_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE comment_reactions (
    comment_id UUID NOT NULL REFERENCES post_comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    reaction TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (comment_id, user_id)
);

CREATE TABLE post_bookmarks (
    user_id UUID NOT NULL,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);

CREATE TABLE post_interaction_events (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL,
    actor_id UUID NOT NULL,
    kind TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE questions (
    id UUID PRIMARY KEY,
    author_id UUID NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    post_id UUID REFERENCES posts(id) ON DELETE SET NULL,
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    upvote_count BIGINT NOT NULL DEFAULT 0,
    downvote_count BIGINT NOT NULL DEFAULT 0,
    answer_count BIGINT NOT NULL DEFAULT 0,
    closed_at TIMESTAMPTZ,
    closed_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE question_votes (
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    vote TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (question_id, user_id)
);

CREATE TABLE answers (
    id UUID PRIMARY KEY,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    author_id UUID NOT NULL,
    body TEXT NOT NULL,
    upvote_count BIGINT NOT NULL DEFAULT 0,
    downvote_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE answer_votes (
    answer_id UUID NOT NULL REFERENCES answers(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    vote TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (answer_id, user_id)
);

CREATE TABLE places (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    summary_model TEXT NOT NULL DEFAULT '',
    summary_updated_at TIMESTAMPTZ,
    post_count BIGINT NOT NULL DEFAULT 0,
    embedding DOUBLE PRECISION[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE place_posts (
    place_id UUID NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    attach_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    attach_method TEXT NOT NULL DEFAULT '',
    attached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (place_id, post_id),
    UNIQUE (post_id)
);

CREATE TABLE place_entries (
    id UUID PRIMARY KEY,
    place_id UUID NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    category TEXT NOT NULL,
    body TEXT NOT NULL,
    source TEXT NOT NULL,
    source_post_id UUID REFERENCES posts(id) ON DELETE SET NULL,
    author_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
