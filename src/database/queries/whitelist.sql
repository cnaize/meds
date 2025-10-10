-- name: GetAllWhiteListSubnets :many
SELECT * FROM sn_whitelist;

-- name: GetAllWhiteListDomains :many
SELECT * FROM dm_whitelist;
