-- +goose Up
CREATE TABLE profiles (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    age        INT         NOT NULL CHECK (age >= 18),
    city       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE profiles;