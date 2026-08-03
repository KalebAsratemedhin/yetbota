-- +goose Up
alter table posts add column address varchar(255);

-- +goose Down
alter table posts drop column address;
