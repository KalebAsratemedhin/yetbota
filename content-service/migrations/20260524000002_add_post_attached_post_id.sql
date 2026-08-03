-- +goose Up
-- +goose StatementBegin
alter table posts add column if not exists attached_post_id uuid references posts(id) on delete set null;

create index if not exists posts_attached_post_id_index on posts (attached_post_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index if exists posts_attached_post_id_index;

alter table posts drop column if exists attached_post_id;
-- +goose StatementEnd
