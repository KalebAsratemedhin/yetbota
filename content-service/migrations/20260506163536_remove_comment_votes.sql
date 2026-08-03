-- +goose Up
-- +goose StatementBegin
select 'up SQL query';
drop table if exists comment_votes;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
select 'down SQL query';

create table comment_votes (
    user_id    uuid               not null,
    comment_id uuid               not null references comments(id) on delete cascade,
    vote_type  comment_vote_type  not null,
    created_at timestamptz        not null default now(),
    primary key (user_id, comment_id)
);


create index on comment_votes (comment_id);
-- +goose StatementEnd
