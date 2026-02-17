-- name: GetAllIPExcludeList :many
SELECT * FROM ip_excludelist;

-- name: UpsertIPExcludeList :exec
INSERT INTO ip_excludelist (subnet)
    VALUES (@subnet)
    ON CONFLICT (subnet) DO NOTHING;

-- name: RemoveIPExcludeList :exec
DELETE FROM ip_excludelist
    WHERE subnet = @subnet;

-- name: GetAllASNExcludeList :many
SELECT * FROM asn_excludelist;

-- name: UpsertASNExcludeList :exec
INSERT INTO asn_excludelist (asn)
    VALUES (@asn)
    ON CONFLICT (asn) DO NOTHING;

-- name: RemoveASNExcludeList :exec
DELETE FROM asn_excludelist
    WHERE asn = @asn;

-- name: GetAllJA3ExcludeList :many
SELECT * FROM ja3_excludelist;

-- name: UpsertJA3ExcludeList :exec
INSERT INTO ja3_excludelist (hash)
    VALUES (@hash)
    ON CONFLICT (hash) DO NOTHING;

-- name: RemoveJA3ExcludeList :exec
DELETE FROM ja3_excludelist
    WHERE hash = @hash;

-- name: GetAllDomainExcludeList :many
SELECT * FROM domain_excludelist;

-- name: UpsertDomainExcludeList :exec
INSERT INTO domain_excludelist (domain)
    VALUES (@domain)
    ON CONFLICT (domain) DO NOTHING;

-- name: RemoveDomainExcludeList :exec
DELETE FROM domain_excludelist
    WHERE domain = @domain;
