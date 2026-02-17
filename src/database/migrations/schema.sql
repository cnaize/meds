CREATE TABLE IF NOT EXISTS ip_allowlist (
    subnet TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS country_blocklist (
    country TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS ip_includelist (
    subnet TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS ip_excludelist (
    subnet TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS asn_includelist (
    asn INTEGER NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS asn_excludelist (
    asn INTEGER NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS ja3_includelist (
    hash TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS ja3_excludelist (
    hash TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS domain_includelist (
    domain TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS domain_excludelist (
    domain TEXT NOT NULL UNIQUE
);
