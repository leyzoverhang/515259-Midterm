-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  difficulties (
    id varchar(50) PRIMARY KEY,
    name varchar(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    deleted_at timestamptz
  );

-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO
  difficulties (id, name)
VALUES
  ('easy', 'Easy'),
  ('medium', 'Medium'),
  ('hard', 'Hard');

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS difficulties;

-- +goose StatementEnd