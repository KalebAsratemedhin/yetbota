-- +goose Up
-- +goose StatementBegin
select 'up SQL query';

create extension if not exists postgis;

create table posts (
    id            uuid                     primary key default gen_random_uuid(),
    title         varchar(255)             not null,
    description   text                     not null,
    likes         integer                  not null default 0,
    comments      integer                  not null default 0,
    user_id       uuid                     not null,
    tags          text[]                   not null,
    is_question   boolean                  not null default false,
    photos        uuid[]                   not null,
    location      geography(Point, 4326),
    created_at    timestamptz              not null default now(),
    updated_at    timestamptz              not null default now()
);

alter table posts add constraint posts_user_id_fk
    foreign key (user_id) references users (id)
    on update cascade on delete cascade;

create index posts_location_idx on posts using gist (location);
create index posts_tags_idx on posts using gin (tags);
create index posts_user_created_idx on posts (user_id, created_at desc);
create index posts_created_at_idx on posts (created_at desc);
create index posts_questions_idx on posts (created_at desc) where is_question = true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
select 'down SQL query';
-- +goose StatementEnd
