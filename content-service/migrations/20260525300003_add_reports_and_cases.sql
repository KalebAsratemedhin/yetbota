-- +goose Up
-- +goose StatementBegin
create table if not exists reports (
    id           uuid        primary key,
    content_type text        not null check (content_type in ('POST', 'COMMENT')),
    content_id   uuid        not null,
    reporter_id  uuid        not null,
    reason       text        not null check (reason in ('SPAM', 'OFFENSIVE', 'INCORRECT', 'OTHER')),
    details      text,
    created_at   timestamptz not null default now(),
    unique (reporter_id, content_type, content_id)
);

create index if not exists reports_content_index on reports (content_type, content_id);

create table if not exists moderation_cases (
    id                uuid        primary key,
    content_type      text        not null check (content_type in ('POST', 'COMMENT')),
    content_id        uuid        not null,
    report_count      int         not null default 0,
    status            text        not null default 'PENDING' check (status in ('PENDING', 'RESOLVED', 'REJECTED')),
    severity          int         not null default 0,
    auto_hidden       boolean     not null default false,
    first_reported_at timestamptz not null default now(),
    last_reported_at  timestamptz not null default now(),
    resolved_by       uuid,
    resolved_at       timestamptz,
    resolution        text        check (resolution in ('DELETED', 'DISMISSED')),
    version           int         not null default 0,
    unique (content_type, content_id)
);

create index if not exists moderation_cases_status_severity_index on moderation_cases (status, severity desc, last_reported_at desc);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists reports;
drop table if exists moderation_cases;
-- +goose StatementEnd
