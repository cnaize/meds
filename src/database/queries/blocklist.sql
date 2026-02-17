-- name: GetAllCountryBlockList :many
SELECT * FROM country_blocklist;

-- name: UpsertCountryBlockList :exec
INSERT INTO country_blocklist (country)
    VALUES (@country)
    ON CONFLICT (country) DO NOTHING;

-- name: RemoveCountryBlockList :exec
DELETE FROM country_blocklist
    WHERE country = @country;
