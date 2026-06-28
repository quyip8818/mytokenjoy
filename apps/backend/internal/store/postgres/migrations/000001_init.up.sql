CREATE TABLE IF NOT EXISTS domain_snapshot (
    id         INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    data       JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS relay_mappings (
    platform_key_id    TEXT PRIMARY KEY,
    newapi_token_id    BIGINT,
    member_id          TEXT,
    department_id      TEXT NOT NULL,
    budget_group_id    TEXT,
    relay_group        TEXT NOT NULL,
    sync_status        TEXT NOT NULL DEFAULT 'pending',
    synced_at          TIMESTAMPTZ,
    relay_remain_quota BIGINT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_relay_mappings_newapi_token ON relay_mappings (newapi_token_id);
CREATE INDEX IF NOT EXISTS idx_relay_mappings_member ON relay_mappings (member_id);
CREATE INDEX IF NOT EXISTS idx_relay_mappings_department ON relay_mappings (department_id);
CREATE INDEX IF NOT EXISTS idx_relay_mappings_budget_group ON relay_mappings (budget_group_id);

CREATE TABLE IF NOT EXISTS relay_outbox (
    id           TEXT PRIMARY KEY,
    kind         TEXT NOT NULL,
    payload      JSONB NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    attempts     INT NOT NULL DEFAULT 0,
    next_retry   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_relay_outbox_pending ON relay_outbox (status, next_retry);

CREATE TABLE IF NOT EXISTS webhook_outbox (
    id           TEXT PRIMARY KEY,
    payload      JSONB NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    attempts     INT NOT NULL DEFAULT 0,
    next_retry   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhook_outbox_pending ON webhook_outbox (status, next_retry);

CREATE TABLE IF NOT EXISTS ingested_log_ids (
    log_id       BIGINT PRIMARY KEY,
    ingested_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS relay_sync_cursors (
    id           INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    last_log_id  BIGINT NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO relay_sync_cursors (id, last_log_id) VALUES (1, 0) ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS rebalance_queue (
    id           TEXT PRIMARY KEY,
    axis_kind    TEXT NOT NULL,
    axis_id      TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (axis_kind, axis_id, status)
);

CREATE INDEX IF NOT EXISTS idx_rebalance_queue_pending ON rebalance_queue (status, created_at);
