-- name: CreateUser :one
INSERT INTO users (
    id,
    name,
    email,
    password_hash,
    profile_image_url
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING id, name, email, profile_image_url, active, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, name, email, profile_image_url, active, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserAuthByEmail :one
SELECT id, name, email, password_hash, profile_image_url, active, created_at, updated_at
FROM users
WHERE email = $1
  AND active = TRUE;

-- name: UpdateUser :one
UPDATE users
SET name = CASE
        WHEN sqlc.arg(update_name)::boolean THEN sqlc.arg(name)::text
        ELSE name
    END,
    profile_image_url = CASE
        WHEN sqlc.arg(update_image)::boolean THEN sqlc.arg(profile_image_url)::text
        ELSE profile_image_url
    END,
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, name, email, profile_image_url, active, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;