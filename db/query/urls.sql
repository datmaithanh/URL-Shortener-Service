-- name: CreateUrl :one
INSERT INTO urls (
    code,
    original_url,
    title
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetUrl :one
SELECT * FROM urls
WHERE id = $1 LIMIT 1;

-- name: ListUrl :many
SELECT * FROM urls
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateUrl :one
UPDATE urls
SET code = $2,
    original_url = $3,
    title = $4
WHERE id = $1
RETURNING *;

-- name: DeleteUrl :exec
DELETE FROM urls
WHERE id = $1;