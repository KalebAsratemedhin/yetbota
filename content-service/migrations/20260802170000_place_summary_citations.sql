-- +goose Up
-- +goose StatementBegin

ALTER TABLE places
    ADD COLUMN IF NOT EXISTS summary_generated_at timestamptz,
    ADD COLUMN IF NOT EXISTS summary_citations jsonb NOT NULL DEFAULT '[]'::jsonb;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE places
    DROP COLUMN IF EXISTS summary_citations,
    DROP COLUMN IF EXISTS summary_generated_at;

-- +goose StatementEnd
