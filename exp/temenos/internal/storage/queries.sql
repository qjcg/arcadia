-- name: GetModule :one
SELECT id, title, description, category, created_at, updated_at
FROM modules
WHERE id = ?;

-- name: ListModules :many
SELECT id, title, description, category, created_at, updated_at
FROM modules;

-- name: CreateModule :exec
INSERT INTO modules (id, title, description, category, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetCard :one
SELECT id, type, module_id, front, back, metadata, created_at, updated_at
FROM cards
WHERE id = ?;

-- name: ListCardsByModule :many
SELECT id, type, module_id, front, back, metadata, created_at, updated_at
FROM cards
WHERE module_id = ?;

-- name: ListCardsForReview :many
SELECT c.id, c.type, c.module_id, c.front, c.back, c.metadata, c.created_at, c.updated_at
FROM cards c
LEFT JOIN reviews r ON c.id = r.card_id
WHERE c.module_id = ?
AND (r.next_review_at IS NULL OR r.next_review_at <= ?)
ORDER BY COALESCE(r.next_review_at, '1970-01-01')
LIMIT ?;

-- name: CreateCard :exec
INSERT INTO cards (id, type, module_id, front, back, metadata, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: CreateReview :exec
INSERT INTO reviews (id, card_id, session_id, quality, interval_days, ease_factor, review_count, next_review_at, reviewed_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetReviewForCard :one
SELECT id, card_id, session_id, quality, interval_days, ease_factor, review_count, next_review_at, reviewed_at
FROM reviews
WHERE card_id = ?
ORDER BY reviewed_at DESC
LIMIT 1;

-- name: UpdateReview :exec
UPDATE reviews
SET quality = ?, interval_days = ?, ease_factor = ?, review_count = ?, next_review_at = ?, reviewed_at = ?
WHERE id = ?;

-- name: CreateSession :exec
INSERT INTO sessions (id, module_id, mode, started_at, ended_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetSession :one
SELECT id, module_id, mode, started_at, ended_at
FROM sessions
WHERE id = ?;

-- name: CountCardsByModule :one
SELECT COUNT(*) as count
FROM cards
WHERE module_id = ?;

-- name: CountReviews :one
SELECT COUNT(*) as count
FROM reviews;

-- name: CountReviewsByModule :one
SELECT COUNT(*) as count
FROM reviews r
JOIN cards c ON r.card_id = c.id
WHERE c.module_id = ?;

-- name: CountCardsDueForReview :one
SELECT COUNT(*) as count
FROM cards c
LEFT JOIN reviews r ON c.id = r.card_id
WHERE r.next_review_at IS NULL OR r.next_review_at <= ?;

-- name: CountCardsDueForReviewByModule :one
SELECT COUNT(*) as count
FROM cards c
LEFT JOIN reviews r ON c.id = r.card_id
WHERE c.module_id = ?
AND (r.next_review_at IS NULL OR r.next_review_at <= ?);

-- name: GetModuleStats :one
SELECT
  COUNT(DISTINCT c.id) as total_cards,
  COUNT(DISTINCT r.id) as reviewed_cards,
  COALESCE(AVG(r.quality), 0) as avg_quality
FROM cards c
LEFT JOIN reviews r ON c.id = r.card_id
WHERE c.module_id = ?;

-- name: GetAllModuleStats :many
SELECT
  c.module_id,
  COUNT(DISTINCT c.id) as total_cards,
  COUNT(DISTINCT r.id) as reviewed_cards,
  COALESCE(AVG(r.quality), 0) as avg_quality
FROM cards c
LEFT JOIN reviews r ON c.id = r.card_id
GROUP BY c.module_id;

-- name: GetReviewDates :many
SELECT reviewed_at
FROM reviews
ORDER BY reviewed_at DESC;
