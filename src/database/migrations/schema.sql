CREATE TABLE IF NOT EXISTS sn_whitelist (
    subnet TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_snwl_subnet ON sn_whitelist (subnet);

CREATE TABLE IF NOT EXISTS sn_blacklist (
    subnet TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_snbl_subnet ON sn_blacklist (subnet);

CREATE TABLE IF NOT EXISTS dm_whitelist (
    domain TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_dmwl_domain ON dm_whitelist (domain);

CREATE TABLE IF NOT EXISTS dm_blacklist (
    domain TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_dmbl_domain ON dm_blacklist (domain);
