/*
 id: a UUID that will serve as the primary key
created_at: a TIMESTAMP that can not be null
updated_at: a TIMESTAMP that can not be null
email: TEXT that can not be null and must be unique
 */

-- +goose Up
CREATE TABLE users (
                       id INTEGER PRIMARY KEY,
                       created_at TIMESTAMP NOT NULL,
                       update_at TIMESTAMP NOT NULL ,
                       email TEXT UNIQUE
);
-- +goose Down
DROP TABLE users;