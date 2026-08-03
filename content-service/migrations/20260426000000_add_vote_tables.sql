-- +goose Up
create type comment_vote_type as enum ('upvote', 'downvote');
create type post_vote_type as enum ('like', 'dislike');

alter table posts add column dislikes integer not null default 0;

create table comment_votes (
    user_id    uuid               not null,
    comment_id uuid               not null references comments(id) on delete cascade,
    vote_type  comment_vote_type  not null,
    created_at timestamptz        not null default now(),
    primary key (user_id, comment_id)
);

create table post_votes (
    user_id    uuid            not null,
    post_id    uuid            not null references posts(id) on delete cascade,
    vote_type  post_vote_type  not null,
    created_at timestamptz     not null default now(),
    primary key (user_id, post_id)
);

create index on comment_votes (comment_id);
create index on post_votes (post_id);

-- +goose Down
drop table if exists post_votes;
drop table if exists comment_votes;
alter table posts drop column if exists dislikes;
drop type if exists post_vote_type;
drop type if exists comment_vote_type;
