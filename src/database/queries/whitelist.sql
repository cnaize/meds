-- name: GetAllWhiteListSubnets :many
SELECT * FROM subnet_whitelist;

-- name: UpsertWhiteListSubnet :exec
INSERT INTO subnet_whitelist (subnet)
VALUES (@subnet);

-- name: RemoveWhiteListSubnet :exec
DELETE FROM subnet_whitelist
WHERE subnet = @subnet;

-- name: GetAllWhiteListDomains :many
SELECT * FROM domain_whitelist;

-- name: UpsertWhiteListDomain :exec
INSERT INTO domain_whitelist (domain)
VALUES (@domain);

-- name: RemoveWhiteListDomain :exec
DELETE FROM domain_whitelist
WHERE domain = @domain;
