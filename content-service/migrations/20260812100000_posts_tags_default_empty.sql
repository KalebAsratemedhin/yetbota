-- +goose Up
-- +goose StatementBegin

-- Empty tags must be '{}' not NULL (NOT NULL column). Harden default for raw inserts.
ALTER TABLE posts ALTER COLUMN tags SET DEFAULT '{}'::text[];
UPDATE posts SET tags = '{}'::text[] WHERE tags IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE posts ALTER COLUMN tags DROP DEFAULT;

-- +goose StatementEnd
