-- +goose Up
-- +goose StatementBegin
alter table posts drop constraint if exists posts_moderation_status_check;
alter table posts add constraint posts_moderation_status_check check (moderation_status in ('VISIBLE', 'HIDDEN', 'REMOVED', 'DUPLICATE'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table posts drop constraint if exists posts_moderation_status_check;
alter table posts add constraint posts_moderation_status_check check (moderation_status in ('VISIBLE', 'HIDDEN', 'REMOVED'));
-- +goose StatementEnd
