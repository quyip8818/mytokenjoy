CREATE TABLE IF NOT EXISTS datasource_credentials (
    id          INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    platform    TEXT NOT NULL,
    encrypted   BYTEA NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS scheduler_locks (
    lock_name    TEXT PRIMARY KEY,
    holder       TEXT NOT NULL,
    lease_until  TIMESTAMPTZ NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
