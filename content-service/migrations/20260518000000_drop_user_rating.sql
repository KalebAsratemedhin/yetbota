-- +goose Up
drop index if exists users_rating_desc_index;
alter table users drop column if exists rating;

-- +goose Down
alter table users add column if not exists rating integer not null default 0;
create index if not exists users_rating_desc_index on users (rating desc);
