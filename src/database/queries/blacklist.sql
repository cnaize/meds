-- name: GetAllBlackListSubnets :many
SELECT * FROM sn_blacklist;

-- name: UpsertBlackListSubnet :exec
INSERT INTO sn_blacklist (subnet)
VALUES (@subnet);

-- name: RemoveBlackListSubnet :exec
DELETE FROM sn_blacklist
WHERE subnet = @subnet;

-- name: GetAllBlackListDomains :many
SELECT * FROM dm_blacklist;

-- name: UpsertBlackListDomain :exec
INSERT INTO dm_blacklist (domain)
VALUES (@domain);

-- name: RemoveBlackListDomain :exec
DELETE FROM dm_blacklist
WHERE domain = @domain;
