-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  recipe_ratings (
    user_id uuid NOT NULL REFERENCES users (id),
    recipe_id integer NOT NULL REFERENCES recipes (id),
    score float NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    deleted_at timestamptz,
    PRIMARY KEY (user_id, recipe_id)
  );

-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_recipe_ratings_recipe_id ON recipe_ratings (recipe_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipe_ratings;

-- +goose StatementEnd