CREATE TABLE IF NOT EXISTS sn_whitelist (
    subnet TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sn_blacklist (
    subnet TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS dm_whitelist (
    domain TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS dm_blacklist (
    domain TEXT NOT NULL
);
