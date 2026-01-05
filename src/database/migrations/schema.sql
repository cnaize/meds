CREATE TABLE IF NOT EXISTS subnet_whitelist (
    subnet TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_snwl_subnet ON subnet_whitelist (subnet);

CREATE TABLE IF NOT EXISTS subnet_blacklist (
    subnet TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_snbl_subnet ON subnet_blacklist (subnet);

CREATE TABLE IF NOT EXISTS country_blacklist (
    country TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_crbl_country ON country_blacklist (country);
