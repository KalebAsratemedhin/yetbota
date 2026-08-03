-- +goose Up
alter table photos add column url_mobile text;
alter table photos add column url_web    text;

-- +goose Down
alter table photos drop column url_web;
alter table photos drop column url_mobile;
