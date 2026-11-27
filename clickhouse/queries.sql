SELECT
    toDate(created_at) AS day,
    uniqExactIf(event_id, type = 'like.created') AS likes
FROM likes.events
GROUP BY day
ORDER BY day;

SELECT
    uniqExactIf(event_id, type = 'match.created') AS matches,
    uniqExactIf(event_id, type = 'like.created') AS likes,
    round(matches / likes, 4) AS match_ratio
FROM likes.events;

SELECT
    to_id,
    uniqExact(event_id) AS likes_received
FROM likes.events
WHERE type = 'like.created'
GROUP BY to_id
ORDER BY likes_received DESC, to_id
LIMIT 10;