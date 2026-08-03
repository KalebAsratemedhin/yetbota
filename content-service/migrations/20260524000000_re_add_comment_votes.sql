-- +goose Up
-- +goose StatementBegin
-- The comment_vote_type enum was created in 20260426000000 and never dropped,
-- so we only recreate the table here.
create table comment_votes (
    user_id    uuid               not null,
    comment_id uuid               not null references comments(id) on delete cascade,
    vote_type  comment_vote_type  not null,
    created_at timestamptz        not null default now(),
    primary key (user_id, comment_id)
);

create index on comment_votes (comment_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists comment_votes;
-- +goose StatementEnd
