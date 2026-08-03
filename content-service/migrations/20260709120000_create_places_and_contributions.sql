-- +goose Up
-- +goose StatementBegin

CREATE TYPE place_status AS ENUM ('PENDING', 'VISIBLE', 'HIDDEN', 'REMOVED', 'DUPLICATE');
CREATE TYPE contribution_kind AS ENUM ('tip', 'change', 'important', 'question', 'answer');

CREATE TABLE places (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name             varchar(255) NOT NULL,
    description      text NOT NULL DEFAULT '',
    address          varchar(255),
    location         geography(Point, 4326),
    created_by       uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    status           place_status NOT NULL DEFAULT 'PENDING',
    duplicate_of     uuid REFERENCES places(id) ON DELETE SET NULL,
    confidence_score smallint NOT NULL DEFAULT 0,
    last_activity_at timestamptz,
    summary_text     text,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX places_location_idx ON places USING gist (location);
CREATE INDEX places_status_idx ON places (status, created_at DESC);
CREATE INDEX places_duplicate_of_idx ON places (duplicate_of) WHERE duplicate_of IS NOT NULL;

CREATE TABLE place_photos (
    place_id uuid NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    photo_id uuid NOT NULL REFERENCES photos(id) ON DELETE CASCADE,
    position int NOT NULL DEFAULT 0,
    PRIMARY KEY (place_id, position)
);

CREATE TABLE place_contributions (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    place_id          uuid NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    user_id           uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    kind              contribution_kind NOT NULL,
    content           text NOT NULL,
    parent_id         uuid REFERENCES place_contributions(id) ON DELETE CASCADE,
    upvotes           int NOT NULL DEFAULT 0,
    downvotes         int NOT NULL DEFAULT 0,
    moderation_status varchar(32) NOT NULL DEFAULT 'VISIBLE',
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT contributions_answer_has_parent
        CHECK (kind != 'answer' OR parent_id IS NOT NULL),
    CONSTRAINT contributions_question_no_parent
        CHECK (kind != 'question' OR parent_id IS NULL),
    CONSTRAINT contributions_non_answer_no_parent
        CHECK (kind = 'answer' OR kind = 'question' OR parent_id IS NULL)
);

CREATE INDEX place_contributions_place_kind_idx
    ON place_contributions (place_id, kind, created_at DESC);
CREATE INDEX place_contributions_parent_idx
    ON place_contributions (parent_id, created_at ASC)
    WHERE parent_id IS NOT NULL;

CREATE TABLE contribution_votes (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    contribution_id uuid NOT NULL REFERENCES place_contributions(id) ON DELETE CASCADE,
    user_id         uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    vote_type       varchar(16) NOT NULL CHECK (vote_type IN ('like', 'dislike')),
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (contribution_id, user_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS contribution_votes;
DROP TABLE IF EXISTS place_contributions;
DROP TABLE IF EXISTS place_photos;
DROP TABLE IF EXISTS places;
DROP TYPE IF EXISTS contribution_kind;
DROP TYPE IF EXISTS place_status;

-- +goose StatementEnd
