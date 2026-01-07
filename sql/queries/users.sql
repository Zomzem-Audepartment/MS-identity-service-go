-- name: CreateUser :one
INSERT INTO users (
  username, password_hash, full_name, email, phone, avatar, role_id, status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 AND deleted_at IS NULL LIMIT 1;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: ListUsers :many
SELECT u.*, r.name as role_name, r.code as role_code
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
WHERE u.deleted_at IS NULL
ORDER BY u.created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET full_name = $2, email = $3, phone = $4, avatar = $5, role_id = $6, status = $7, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
UPDATE users SET deleted_at = NOW(), status = 'DELETED' WHERE id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(), last_login_ip = $2
WHERE id = $1;

-- name: AssignEmployeeId :exec
UPDATE users
SET employee_id = $2
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL LIMIT 1;

-- name: UpdateUserExternalLogin :exec
UPDATE users
SET external_login = $2
WHERE id = $1;
