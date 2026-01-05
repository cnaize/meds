-- name: GetAllBlackListSubnets :many
SELECT * FROM subnet_blacklist;

-- name: UpsertBlackListSubnet :exec
INSERT INTO subnet_blacklist (subnet)
VALUES (@subnet);

-- name: RemoveBlackListSubnet :exec
DELETE FROM subnet_blacklist
WHERE subnet = @subnet;

-- name: GetAllBlackListCountries :many
SELECT * FROM country_blacklist;

-- name: UpsertBlackListCountry :exec
INSERT INTO country_blacklist (country)
VALUES (@country);

-- name: RemoveBlackListCountry :exec
DELETE FROM country_blacklist
WHERE country = @country;
