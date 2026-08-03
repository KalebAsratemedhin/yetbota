-- +goose Up
-- +goose StatementBegin
create table user_devices (
    id         uuid            primary key default gen_random_uuid(),
    user_id    uuid            not null references users(id) on update cascade on delete cascade,
    device_id  varchar(255)    not null,
    token      varchar(255),
    oem        varchar(255),
    device     varchar(255),
    os         varchar(255),
    long       double precision,
    lat        double precision,
    created_at timestamp       not null default current_timestamp,
    updated_at timestamp       not null,
    constraint user_device_token_unique unique (device_id, token)
);

create index user_devices_user_id_idx on user_devices (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table user_devices cascade;
-- +goose StatementEnd
