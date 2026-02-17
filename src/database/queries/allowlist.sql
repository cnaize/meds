-- name: GetAllIPAllowList :many
SELECT * FROM ip_allowlist;

-- name: UpsertIPAllowList :exec
INSERT INTO ip_allowlist (subnet)
    VALUES (@subnet)
    ON CONFLICT (subnet) DO NOTHING;

-- name: RemoveIPAllowList :exec
DELETE FROM ip_allowlist
    WHERE subnet = @subnet;
