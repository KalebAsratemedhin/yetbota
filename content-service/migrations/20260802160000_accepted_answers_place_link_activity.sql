-- +goose Up
-- +goose StatementBegin

ALTER TABLE place_contributions
    ADD COLUMN accepted_at timestamptz,
    ADD COLUMN accepted_by uuid REFERENCES users(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX contributions_one_accepted_answer_per_question
    ON place_contributions (parent_id)
    WHERE kind = 'answer' AND accepted_at IS NOT NULL;

ALTER TABLE posts
    ADD COLUMN place_id uuid REFERENCES places(id) ON DELETE SET NULL;

CREATE INDEX posts_place_id_idx ON posts (place_id) WHERE place_id IS NOT NULL;

CREATE TABLE place_activity_events (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    place_id   uuid NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    actor_id   uuid REFERENCES users(id) ON DELETE SET NULL,
    event_type varchar(64) NOT NULL,
    ref_id     uuid,
    summary    text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX place_activity_events_created_idx
    ON place_activity_events (created_at DESC);

CREATE INDEX place_activity_events_place_created_idx
    ON place_activity_events (place_id, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS place_activity_events;
DROP INDEX IF EXISTS posts_place_id_idx;
ALTER TABLE posts DROP COLUMN IF EXISTS place_id;
DROP INDEX IF EXISTS contributions_one_accepted_answer_per_question;
ALTER TABLE place_contributions
    DROP COLUMN IF EXISTS accepted_by,
    DROP COLUMN IF EXISTS accepted_at;

-- +goose StatementEnd
