-- +goose Up
-- +goose StatementBegin
alter table posts add column if not exists moderation_status text not null default 'VISIBLE' check (moderation_status in ('VISIBLE', 'HIDDEN', 'REMOVED'));
alter table posts add column if not exists deleted_at timestamptz;

create index if not exists posts_moderation_status_index on posts (moderation_status);

alter table comments add column if not exists moderation_status text not null default 'VISIBLE' check (moderation_status in ('VISIBLE', 'HIDDEN', 'REMOVED'));
alter table comments add column if not exists deleted_at timestamptz;

create index if not exists comments_moderation_status_index on comments (moderation_status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index if exists posts_moderation_status_index;

alter table posts drop column if exists moderation_status;
alter table posts drop column if exists deleted_at;

drop index if exists comments_moderation_status_index;

alter table comments drop column if exists moderation_status;
alter table comments drop column if exists deleted_at;
-- +goose StatementEnd
