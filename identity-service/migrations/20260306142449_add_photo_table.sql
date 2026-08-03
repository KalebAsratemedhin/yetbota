-- +goose Up
-- +goose StatementBegin
select 'up SQL query';

create type photo_bucket as enum('S3', 'GCS');

create table photos (
    id                 uuid            primary key default gen_random_uuid(),
    bucket_provider    photo_bucket    not null,
    mime_type          varchar(20)     not null,
    url                text            not null,
    created_at         timestamp       not null default current_timestamp,
    updated_at         timestamp       not null default current_timestamp
);

alter table users add column profile_photo_id uuid;
alter table users add constraint users_profile_photo_id_fk 
    foreign key (profile_photo_id) references photos (id)
    on delete set null on update cascade;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
select 'down SQL query';

alter table users drop constraint users_profile_photo_id_fk;
alter table users drop column profile_photo_id;
drop table photos;
-- +goose StatementEnd
