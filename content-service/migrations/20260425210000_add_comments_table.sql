-- +goose Up
-- +goose StatementBegin
select 'up SQL query';

create table comments (
    id          uuid        primary key default gen_random_uuid(),
    comment     text        not null,
    upvote      integer     not null default 0,
    downvote    integer     not null default 0,
    user_id     uuid        not null,
    post_id     uuid        not null,
    is_answer   boolean     not null default false,
    comment_id  uuid,
    created_at  timestamptz not null default now(),
    updated_at  timestamptz not null default now()
);

alter table comments add constraint comments_post_id_fk
    foreign key (post_id) references posts (id)
    on update cascade on delete cascade;

alter table comments add constraint comments_comment_id_fk
    foreign key (comment_id) references comments (id)
    on update cascade on delete cascade;

create index comments_post_id_idx on comments (post_id, created_at desc);
create index comments_comment_id_idx on comments (comment_id, created_at desc);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
select 'down SQL query';

drop table comments;
-- +goose StatementEnd
