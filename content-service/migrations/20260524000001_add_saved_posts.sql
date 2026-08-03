-- +goose Up
-- +goose StatementBegin
create table saved_posts (
    user_id    uuid        not null,
    post_id    uuid        not null references posts(id) on delete cascade,
    created_at timestamptz not null default now(),
    primary key (user_id, post_id)
);

create index on saved_posts (user_id, created_at desc);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists saved_posts;
-- +goose StatementEnd
