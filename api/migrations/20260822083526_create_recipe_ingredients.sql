-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  recipe_ingredients (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    recipe_id integer NOT NULL REFERENCES recipes (id),
    description text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    deleted_at timestamptz
  );

-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_recipe_ingredients_recipe_id ON recipe_ingredients (recipe_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipe_ingredients;

-- +goose StatementEnd