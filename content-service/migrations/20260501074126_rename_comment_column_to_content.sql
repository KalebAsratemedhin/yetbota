-- +goose Up
-- +goose StatementBegin
select 'up SQL query';
alter table comments rename column comment to content;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
select 'down SQL query';
alter table comments rename column content to comment;
-- +goose StatementEnd
