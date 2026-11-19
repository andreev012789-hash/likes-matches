-- +goose Up
CREATE TABLE matches (
    user_a     BIGINT      NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    user_b     BIGINT      NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_a, user_b),
    CHECK (user_a < user_b)
);

-- +goose Down
DROP TABLE matches;