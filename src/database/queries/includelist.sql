-- name: GetAllIPIncludeList :many
SELECT * FROM ip_includelist;

-- name: UpsertIPIncludeList :exec
INSERT INTO ip_includelist (subnet)
    VALUES (@subnet)
    ON CONFLICT (subnet) DO NOTHING;

-- name: RemoveIPIncludeList :exec
DELETE FROM ip_includelist
    WHERE subnet = @subnet;

-- name: GetAllASNIncludeList :many
SELECT * FROM asn_includelist;

-- name: UpsertASNIncludeList :exec
INSERT INTO asn_includelist (asn)
    VALUES (@asn)
    ON CONFLICT (asn) DO NOTHING;

-- name: RemoveASNIncludeList :exec
DELETE FROM asn_includelist
    WHERE asn = @asn;

-- name: GetAllJA3IncludeList :many
SELECT * FROM ja3_includelist;

-- name: UpsertJA3IncludeList :exec
INSERT INTO ja3_includelist (hash)
    VALUES (@hash)
    ON CONFLICT (hash) DO NOTHING;

-- name: RemoveJA3IncludeList :exec
DELETE FROM ja3_includelist
    WHERE hash = @hash;

-- name: GetAllDomainIncludeList :many
SELECT * FROM domain_includelist;

-- name: UpsertDomainIncludeList :exec
INSERT INTO domain_includelist (domain)
    VALUES (@domain)
    ON CONFLICT (domain) DO NOTHING;

-- name: RemoveDomainIncludeList :exec
DELETE FROM domain_includelist
    WHERE domain = @domain;
