-- +goose Up
-- +goose StatementBegin
create table if not exists moderation_actions (
    id          uuid        primary key,
    case_id     uuid        not null references moderation_cases(id) on delete cascade,
    admin_id    uuid        not null,
    action      text        not null check (action in ('DELETE', 'DISMISS', 'BAN', 'UNHIDE')),
    target_type text        not null,
    target_id   uuid        not null,
    note        text,
    created_at  timestamptz not null default now()
);

create index if not exists moderation_actions_case_index on moderation_actions (case_id, created_at desc);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists moderation_actions;
-- +goose StatementEnd
