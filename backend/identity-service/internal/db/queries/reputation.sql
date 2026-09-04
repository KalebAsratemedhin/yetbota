-- name: AddReputationScore :one
UPDATE users SET reputation_score = GREATEST(0, reputation_score + $2)
WHERE id = $1
RETURNING *;

-- name: SetReputationScore :one
UPDATE users SET reputation_score = GREATEST(0, sqlc.arg(score)::bigint)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: GetUserReputation :one
SELECT reputation_score FROM users WHERE id = $1;

-- name: InsertReputationEvent :one
INSERT INTO reputation_events (id, user_id, actor_id, event_type, subject_type, subject_id, delta)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: SumReputationDeltaToday :one
SELECT COALESCE(SUM(delta), 0)::bigint AS total
FROM reputation_events
WHERE user_id = $1
  AND created_at >= date_trunc('day', NOW() AT TIME ZONE 'UTC')
  AND delta > 0;

-- name: CountReputationEventsByType :one
SELECT COUNT(*)::bigint AS count
FROM reputation_events
WHERE user_id = $1 AND event_type = $2;

-- name: ListBadges :many
SELECT id, name, description FROM badges ORDER BY id;

-- name: ListUserBadges :many
SELECT b.id, b.name, b.description, ub.awarded_at
FROM user_badges ub
INNER JOIN badges b ON b.id = ub.badge_id
WHERE ub.user_id = $1
ORDER BY ub.awarded_at ASC;

-- name: AwardBadge :execrows
INSERT INTO user_badges (user_id, badge_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: HasBadge :one
SELECT EXISTS(
    SELECT 1 FROM user_badges WHERE user_id = $1 AND badge_id = $2
) AS has_badge;
