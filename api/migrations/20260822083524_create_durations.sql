-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  durations (
    id varchar(50) PRIMARY KEY,
    name varchar(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    deleted_at timestamptz
  );

-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO
  durations (id, name)
VALUES
  ('10m', '5 - 10 mins'),
  ('30m', '10 - 30 mins'),
  ('60m', '~1 Hour'),
  ('long', 'More than 1 hour');

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS durations;

-- +goose StatementEnd