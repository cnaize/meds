-- name: GetAllWhiteListSubnets :many
SELECT * FROM sn_whitelist;

-- name: UpsertWhiteListSubnet :exec
INSERT INTO sn_whitelist (subnet)
VALUES (@subnet);

-- name: RemoveWhiteListSubnet :exec
DELETE FROM sn_whitelist
WHERE subnet = @subnet;

-- name: GetAllWhiteListDomains :many
SELECT * FROM dm_whitelist;
