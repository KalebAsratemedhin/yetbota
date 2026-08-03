-- +goose Up
-- +goose StatementBegin
select 'up SQL query';

create table post_photos (
    id          uuid    primary key default gen_random_uuid(),
    post_id     uuid    not null,
    photo_id    uuid    not null,
    position    integer not null
);

alter table post_photos add constraint post_photos_post_id_fk
    foreign key (post_id) references posts (id)
    on update cascade on delete cascade;

alter table post_photos add constraint post_photos_photo_id_fk
    foreign key (photo_id) references photos (id)
    on update cascade on delete cascade;

create unique index post_photos_post_id_position_idx on post_photos (post_id, position);

alter table posts drop column photos;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
select 'down SQL query';

drop table post_photos;

alter table posts add column photos uuid[] not null default '{}';
-- +goose StatementEnd
