-- +goose Up
-- +goose StatementBegin
create table if not exists notifications (
    id          text        not null primary key default gen_random_uuid()::text,
    user_id     text        not null,
    title       text        not null,
    body        text        not null,
    data        jsonb       not null default '{}',
    attachment  text,
    sent_at     timestamptz not null default now(),
    read_at     timestamptz
);

create index if not exists notifications_user_id_idx on notifications (user_id);
create index if not exists notifications_sent_at_idx on notifications (user_id, sent_at desc);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists notifications;
-- +goose StatementEnd
