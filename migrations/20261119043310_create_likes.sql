-- +goose Up
CREATE TABLE likes (
    from_id    BIGINT      NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    to_id      BIGINT      NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (from_id, to_id),
    CHECK (from_id <> to_id)
);

CREATE INDEX likes_to_id_idx ON likes (to_id);

-- +goose Down
DROP TABLE likes;