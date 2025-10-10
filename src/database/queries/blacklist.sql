-- name: GetAllBlackListSubnets :many
SELECT * FROM sn_blacklist;

-- name: GetAllBlackListDomains :many
SELECT * FROM dm_blacklist;
