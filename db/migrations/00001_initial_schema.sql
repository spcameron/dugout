-- +goose Up
CREATE TABLE schema_migrations_guard (
    id int PRIMARY KEY
);

-- +goose Down
DROP TABLE schema_migrations_guard;

