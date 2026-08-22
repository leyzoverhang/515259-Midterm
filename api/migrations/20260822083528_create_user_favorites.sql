-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  user_favorites (
    user_id uuid NOT NULL REFERENCES users (id),
    recipe_id integer NOT NULL REFERENCES recipes (id),
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    deleted_at timestamptz,
    PRIMARY KEY (user_id, recipe_id)
  );

-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_user_favorites_recipe_id ON user_favorites (recipe_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_favorites;

-- +goose StatementEnd