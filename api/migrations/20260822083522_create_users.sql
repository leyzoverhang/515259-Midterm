-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE
  users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    email varchar(255) NOT NULL,
    uid varchar(255) NOT NULL,
    name varchar(255),
    bio text,
    image_url text,
    preferred_username varchar(255),
    last_signed_in_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    deleted_at timestamptz,
    CONSTRAINT uq_users_uid UNIQUE (uid)
  );

-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_users_email ON users (email);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;

-- +goose StatementEnd