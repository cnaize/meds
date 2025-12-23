-- name: GetAllBlackListSubnets :many
SELECT * FROM subnet_blacklist;

-- name: UpsertBlackListSubnet :exec
INSERT INTO subnet_blacklist (subnet)
VALUES (@subnet);

-- name: RemoveBlackListSubnet :exec
DELETE FROM subnet_blacklist
WHERE subnet = @subnet;

-- name: GetAllBlackListDomains :many
SELECT * FROM domain_blacklist;

-- name: UpsertBlackListDomain :exec
INSERT INTO domain_blacklist (domain)
VALUES (@domain);

-- name: RemoveBlackListDomain :exec
DELETE FROM domain_blacklist
WHERE domain = @domain;

-- name: GetAllBlackListCountries :many
SELECT * FROM country_blacklist;

-- name: UpsertBlackListCountry :exec
INSERT INTO country_blacklist (country)
VALUES (@country);

-- name: RemoveBlackListCountry :exec
DELETE FROM country_blacklist
WHERE country = @country;
