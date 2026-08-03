-- +goose Up
-- +goose StatementBegin
alter type user_badges add value if not exists 'explorer';
alter type user_badges add value if not exists 'trailblazer';
alter type user_badges add value if not exists 'hidden_gem_finder';
alter type user_badges add value if not exists 'spot_verified';
alter type user_badges add value if not exists 'question_master';
alter type user_badges add value if not exists 'answer_guru';
alter type user_badges add value if not exists 'contributor';
alter type user_badges add value if not exists 'trusted_voice';
alter type user_badges add value if not exists 'expert';
alter type user_badges add value if not exists 'master';
alter type user_badges add value if not exists 'grandmaster';

alter table users add column if not exists score bigint not null default 1500;

create index if not exists users_score_desc_index on users (score desc);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index if exists users_score_desc_index;

alter table users drop column if exists score;
-- +goose StatementEnd
