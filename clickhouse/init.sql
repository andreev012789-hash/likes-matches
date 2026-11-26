CREATE DATABASE IF NOT EXISTS likes;
CREATE TABLE IF NOT EXISTS likes.events
(
    event_id   UUID,
    type       LowCardinality(String),
    from_id    Int64,
    to_id      Int64,
    created_at DateTime64(3, 'UTC')
)
ENGINE = MergeTree
ORDER BY (created_at, type);