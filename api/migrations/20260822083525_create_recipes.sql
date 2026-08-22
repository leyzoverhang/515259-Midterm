-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  recipes (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name varchar(255) NOT NULL,
    description text NOT NULL,
    image_url text,
    difficulty_id varchar(50) REFERENCES difficulties (id),
    duration_id varchar(50) REFERENCES durations (id),
    average_rating float NOT NULL DEFAULT 0,
    creator_id uuid REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    deleted_at timestamptz
  );

-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_recipes_difficulty_id ON recipes (difficulty_id);

-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_recipes_duration_id ON recipes (duration_id);

-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_recipes_creator_id ON recipes (creator_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipes;

-- +goose StatementEnd