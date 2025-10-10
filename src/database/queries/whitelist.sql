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

-- name: UpsertWhiteListDomain :exec
INSERT INTO dm_whitelist (domain)
VALUES (@domain);

-- name: RemoveWhiteListDomain :exec
DELETE FROM dm_whitelist
WHERE domain = @domain;
