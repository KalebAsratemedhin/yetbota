-- +goose Up
-- +goose StatementBegin

ALTER TABLE places
    ADD COLUMN last_confirmed_at timestamptz,
    ADD COLUMN needs_check boolean NOT NULL DEFAULT false;

CREATE INDEX places_needs_check_idx
    ON places (needs_check, last_activity_at DESC NULLS LAST)
    WHERE needs_check = true;

CREATE TABLE place_bookmarks (
    user_id    uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    place_id   uuid NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, place_id)
);

CREATE INDEX place_bookmarks_place_idx ON place_bookmarks (place_id);

CREATE TABLE place_confirmations (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    place_id     uuid NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    user_id      uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    note         text,
    confirmed_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (place_id, user_id)
);

CREATE INDEX place_confirmations_place_confirmed_idx
    ON place_confirmations (place_id, confirmed_at DESC);

CREATE TYPE place_problem_kind AS ENUM (
    'PERMANENTLY_CLOSED',
    'DOES_NOT_EXIST',
    'DUPLICATE',
    'WRONG_LOCATION'
);

CREATE TYPE place_problem_status AS ENUM ('PENDING', 'RESOLVED', 'DISMISSED');

CREATE TABLE place_problem_reports (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    place_id      uuid NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    user_id       uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    kind          place_problem_kind NOT NULL,
    note          text,
    duplicate_of  uuid REFERENCES places(id) ON DELETE SET NULL,
    status        place_problem_status NOT NULL DEFAULT 'PENDING',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT place_problem_duplicate_requires_target
        CHECK (kind != 'DUPLICATE' OR duplicate_of IS NOT NULL)
);

CREATE INDEX place_problem_reports_place_status_idx
    ON place_problem_reports (place_id, status, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS place_problem_reports;
DROP TYPE IF EXISTS place_problem_status;
DROP TYPE IF EXISTS place_problem_kind;
DROP TABLE IF EXISTS place_confirmations;
DROP TABLE IF EXISTS place_bookmarks;
DROP INDEX IF EXISTS places_needs_check_idx;
ALTER TABLE places
    DROP COLUMN IF EXISTS needs_check,
    DROP COLUMN IF EXISTS last_confirmed_at;

-- +goose StatementEnd
