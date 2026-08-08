-- +goose Up
-- +goose StatementBegin

-- Places are public on create so digester/geo linking and place pages work immediately.
-- PENDING remains available for admin/moderation holds.

ALTER TABLE places ALTER COLUMN status SET DEFAULT 'VISIBLE';

UPDATE places SET status = 'VISIBLE', updated_at = now()
WHERE status = 'PENDING';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE places ALTER COLUMN status SET DEFAULT 'PENDING';

-- +goose StatementEnd
