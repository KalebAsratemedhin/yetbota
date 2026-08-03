-- +goose Up
-- +goose StatementBegin
select 'up SQL query';

alter type photo_bucket add value if not exists 'CLOUDINARY';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
select 'down SQL query';
-- Postgres does not support removing enum values; down migration is a no-op.
-- +goose StatementEnd
