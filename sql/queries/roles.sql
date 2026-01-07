-- name: GetRoleById :one
SELECT * FROM roles WHERE id = $1 LIMIT 1;

-- name: GetRoleByCode :one
SELECT * FROM roles WHERE code = $1 LIMIT 1;

-- name: ListRoles :many
SELECT * FROM roles ORDER BY level ASC;

-- name: CreateRole :one
INSERT INTO roles (code, name, description, level, is_system, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles
SET name = $2, description = $3, level = $4, status = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1 AND is_system = false;

-- name: GetRolePermissions :many
SELECT p.*, rp.data_scope
FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1;

-- name: GetUserPermissions :many
SELECT p.code as permission_code, rp.data_scope
FROM users u
JOIN roles r ON u.role_id = r.id
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE u.id = $1;

-- name: ListPermissions :many
SELECT * FROM permissions ORDER BY module, action;

-- name: ListRolePermissions :many
SELECT rp.*, p.code as permission_code
FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.id
WHERE rp.role_id = $1;

-- name: AssignPermissionToRole :one
INSERT INTO role_permissions (role_id, permission_id, data_scope)
VALUES ($1, $2, $3)
ON CONFLICT (role_id, permission_id) DO UPDATE SET data_scope = EXCLUDED.data_scope
RETURNING *;

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2;
