-- +goose Up
-- +goose StatementBegin
drop index if exists users_mobile_uindex;
alter table users drop column if exists mobile;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users add column mobile varchar(20) not null default '';
alter table users alter column mobile drop default;
create unique index users_mobile_uindex on users (mobile);
-- +goose StatementEnd
