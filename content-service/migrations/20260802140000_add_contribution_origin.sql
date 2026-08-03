-- +goose Up
-- +goose StatementBegin

CREATE TYPE contribution_origin AS ENUM ('USER', 'POST_EXTRACTED');

ALTER TABLE place_contributions
    ADD COLUMN origin contribution_origin NOT NULL DEFAULT 'USER',
    ADD COLUMN source_post_id uuid REFERENCES posts(id) ON DELETE SET NULL;

ALTER TABLE place_contributions
    ADD CONSTRAINT contributions_origin_source_post_check
    CHECK (
        (origin = 'USER' AND source_post_id IS NULL)
        OR (origin = 'POST_EXTRACTED' AND source_post_id IS NOT NULL)
    );

CREATE INDEX place_contributions_origin_idx
    ON place_contributions (place_id, origin, created_at DESC);

CREATE INDEX place_contributions_source_post_idx
    ON place_contributions (source_post_id)
    WHERE source_post_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS place_contributions_source_post_idx;
DROP INDEX IF EXISTS place_contributions_origin_idx;

ALTER TABLE place_contributions
    DROP CONSTRAINT IF EXISTS contributions_origin_source_post_check;

ALTER TABLE place_contributions
    DROP COLUMN IF EXISTS source_post_id,
    DROP COLUMN IF EXISTS origin;

DROP TYPE IF EXISTS contribution_origin;

-- +goose StatementEnd
