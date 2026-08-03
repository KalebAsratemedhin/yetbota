-- +goose Up
alter table posts rename column comments to comment_count;

-- +goose Down
alter table posts rename column comment_count to comments;
