-- +goose Up
-- +goose StatementBegin
create table if not exists user_seen_posts (
    user_id text        not null,
    post_id text        not null,
    seen_at timestamptz not null default now(),
    primary key (user_id, post_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists user_seen_posts;
-- +goose StatementEnd
