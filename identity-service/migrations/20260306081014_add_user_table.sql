-- +goose Up
-- +goose StatementBegin
select 'up SQL query';

create type user_badges as enum();

create type user_status as enum('ACTIVE', 'INACTIVE');

create type roles as enum('ADMIN', 'USER');

create table users (
    id              uuid            primary key default gen_random_uuid(),
    first_name      varchar(255)    not null,
    last_name       varchar(255)    not null,
    username        varchar(255)    not null,
    mobile          varchar(20)     not null,
    badges          user_badges[]   not null default '{}',
    rating          integer         not null default 0,
    contributions   integer         not null default 0,
    status          user_status     not null default 'ACTIVE',
    followers       integer         not null default 0,
    following       integer         not null default 0,
    role            roles           not null default 'USER',
    password        varchar(255)    not null,
    created_at      timestamp       not null default current_timestamp,
    updated_at      timestamp       not null default current_timestamp
);

-- unique lookups
create unique index users_username_uindex on users (username);
create unique index users_mobile_uindex on users (mobile);


-- leaderboard queries
create index users_rating_desc_index on users (rating desc);
create index users_followers_desc_index on users (followers desc);
create index users_contributions_desc_index on users (contributions desc);
create index users_status_index on users (status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
select 'down SQL query';

drop table users;
-- +goose StatementEnd

