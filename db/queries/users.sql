-- name: CreateUser :one
INSERT INTO users (
  email,
  full_name,
  phone,
  role,
  is_active
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at
LIMIT $1
OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET
  email = COALESCE($2, email),
  full_name = COALESCE($3, full_name),
  phone = COALESCE($4, phone),
  role = COALESCE($5, role),
  is_active = COALESCE($6, is_active),
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
