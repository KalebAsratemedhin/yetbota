-- +goose Up
CREATE TABLE IF NOT EXISTS places (
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

CREATE INDEX IF NOT EXISTS places_created_idx ON places (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS places_name_lower_idx ON places (lower(name));
CREATE INDEX IF NOT EXISTS places_geo_idx ON places (lat, lng);

CREATE TABLE IF NOT EXISTS place_posts (
    place_id UUID NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    attach_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    attach_method TEXT NOT NULL DEFAULT '',
    attached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (place_id, post_id),
    UNIQUE (post_id)
);

CREATE INDEX IF NOT EXISTS place_posts_place_attached_idx
    ON place_posts (place_id, attached_at DESC, post_id DESC);

CREATE TABLE IF NOT EXISTS place_entries (
    id UUID PRIMARY KEY,
    place_id UUID NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    category TEXT NOT NULL CHECK (category IN ('tip', 'change', 'important')),
    body TEXT NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('ai', 'user')),
    source_post_id UUID REFERENCES posts(id) ON DELETE SET NULL,
    author_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS place_entries_place_created_idx
    ON place_entries (place_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS place_entries_place_category_created_idx
    ON place_entries (place_id, category, created_at DESC, id DESC);

-- +goose Down
DROP TABLE IF EXISTS place_entries;
DROP TABLE IF EXISTS place_posts;
DROP TABLE IF EXISTS places;
