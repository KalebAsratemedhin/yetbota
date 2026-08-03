-- +goose Up
-- +goose StatementBegin
alter type user_status add value if not exists 'BANNED';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
select 'down SQL query';
-- +goose StatementEnd
