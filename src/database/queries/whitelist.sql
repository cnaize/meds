-- name: GetAllWhiteListSubnets :many
SELECT * FROM subnet_whitelist;

-- name: UpsertWhiteListSubnet :exec
INSERT INTO subnet_whitelist (subnet)
VALUES (@subnet);

-- name: RemoveWhiteListSubnet :exec
DELETE FROM subnet_whitelist
WHERE subnet = @subnet;
