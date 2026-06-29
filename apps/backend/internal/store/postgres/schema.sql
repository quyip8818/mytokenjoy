-- Org domain
CREATE TABLE IF NOT EXISTS permissions (
    id    TEXT PRIMARY KEY,
    name  TEXT NOT NULL,
    grp   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS roles (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    type         TEXT NOT NULL,
    member_count INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS role_permission_grants (
    role_id        TEXT NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_ref TEXT NOT NULL,
    PRIMARY KEY (role_id, permission_ref)
);

CREATE TABLE IF NOT EXISTS departments (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    parent_id     TEXT REFERENCES departments (id) ON DELETE SET NULL,
    member_count  INT NOT NULL DEFAULT 0,
    external_id   TEXT,
    source        TEXT,
    manager_id    TEXT,
    sort_order    INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_departments_parent ON departments (parent_id);

CREATE TABLE IF NOT EXISTS members (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    phone           TEXT NOT NULL DEFAULT '',
    email           TEXT NOT NULL DEFAULT '',
    department_id   TEXT NOT NULL REFERENCES departments (id),
    department_name TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL,
    source          TEXT NOT NULL DEFAULT '',
    external_id     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_members_department ON members (department_id);
CREATE INDEX IF NOT EXISTS idx_members_status ON members (status);

ALTER TABLE departments DROP CONSTRAINT IF EXISTS departments_manager_id_fkey;
ALTER TABLE departments
    ADD CONSTRAINT departments_manager_id_fkey
    FOREIGN KEY (manager_id) REFERENCES members (id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS member_roles (
    member_id TEXT NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    role_id   TEXT NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    PRIMARY KEY (member_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_member_roles_role ON member_roles (role_id);

CREATE TABLE IF NOT EXISTS org_data_source_status (
    id               INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    platform         TEXT,
    connected        BOOLEAN NOT NULL DEFAULT FALSE,
    last_import      TIMESTAMPTZ,
    last_import_ok   INT,
    last_import_fail INT,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS org_sync_config (
    id                          INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    enabled                     BOOLEAN NOT NULL DEFAULT FALSE,
    start_time                  TEXT NOT NULL DEFAULT '',
    frequency_hours             INT NOT NULL DEFAULT 24,
    delete_member_threshold     INT NOT NULL DEFAULT 0,
    delete_department_threshold INT NOT NULL DEFAULT 0,
    notify_phone                BOOLEAN NOT NULL DEFAULT FALSE,
    notify_email                BOOLEAN NOT NULL DEFAULT FALSE,
    notify_im                   BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS org_sync_logs (
    id     TEXT PRIMARY KEY,
    time   TIMESTAMPTZ NOT NULL,
    type   TEXT NOT NULL,
    result TEXT NOT NULL,
    detail TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_org_sync_logs_time ON org_sync_logs (time DESC);

CREATE TABLE IF NOT EXISTS org_import_failures (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    employee_id TEXT NOT NULL DEFAULT '',
    reason      TEXT NOT NULL
);

-- Budget domain
CREATE TABLE IF NOT EXISTS budget_nodes (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    parent_id     TEXT REFERENCES budget_nodes (id) ON DELETE SET NULL,
    budget        NUMERIC(18, 6) NOT NULL DEFAULT 0,
    consumed      NUMERIC(18, 6) NOT NULL DEFAULT 0,
    reserved_pool NUMERIC(18, 6),
    period        TEXT NOT NULL,
    sort_order    INT NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_budget_nodes_parent ON budget_nodes (parent_id);

CREATE TABLE IF NOT EXISTS budget_groups (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    budget     NUMERIC(18, 6) NOT NULL DEFAULT 0,
    consumed   NUMERIC(18, 6) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS budget_group_members (
    group_id  TEXT NOT NULL REFERENCES budget_groups (id) ON DELETE CASCADE,
    member_id TEXT NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, member_id)
);

CREATE TABLE IF NOT EXISTS budget_group_departments (
    group_id      TEXT NOT NULL REFERENCES budget_groups (id) ON DELETE CASCADE,
    department_id TEXT NOT NULL REFERENCES departments (id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, department_id)
);

CREATE TABLE IF NOT EXISTS overrun_policy (
    id            INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    thresholds    INT[] NOT NULL DEFAULT '{}',
    notify_email  BOOLEAN NOT NULL DEFAULT FALSE,
    notify_phone  BOOLEAN NOT NULL DEFAULT FALSE,
    notify_im     BOOLEAN NOT NULL DEFAULT FALSE,
    block_message TEXT NOT NULL DEFAULT '',
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alert_rules (
    id         TEXT PRIMARY KEY,
    node_id    TEXT NOT NULL REFERENCES budget_nodes (id) ON DELETE CASCADE,
    node_name  TEXT NOT NULL,
    thresholds INT[] NOT NULL DEFAULT '{}',
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alert_rule_notify_roles (
    rule_id TEXT NOT NULL REFERENCES alert_rules (id) ON DELETE CASCADE,
    role_id TEXT NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    PRIMARY KEY (rule_id, role_id)
);

CREATE TABLE IF NOT EXISTS member_quota_pools (
    member_id      TEXT PRIMARY KEY REFERENCES members (id) ON DELETE CASCADE,
    personal_quota NUMERIC(18, 6) NOT NULL DEFAULT 0,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Models domain (before keys FK references)
CREATE TABLE IF NOT EXISTS models (
    id           TEXT PRIMARY KEY,
    provider     TEXT NOT NULL,
    name         TEXT NOT NULL,
    display_name TEXT NOT NULL,
    input_price  NUMERIC(18, 8) NOT NULL DEFAULT 0,
    output_price NUMERIC(18, 8) NOT NULL DEFAULT 0,
    max_context  INT NOT NULL DEFAULT 0,
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS model_capabilities (
    model_id   TEXT NOT NULL REFERENCES models (id) ON DELETE CASCADE,
    capability TEXT NOT NULL,
    PRIMARY KEY (model_id, capability)
);

CREATE TABLE IF NOT EXISTS routing_rules (
    id             TEXT PRIMARY KEY,
    node_id        TEXT NOT NULL,
    node_name      TEXT NOT NULL,
    default_model  TEXT,
    fallback_model TEXT,
    inherited      BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_routing_rules_node ON routing_rules (node_id);

CREATE TABLE IF NOT EXISTS routing_rule_models (
    rule_id    TEXT NOT NULL REFERENCES routing_rules (id) ON DELETE CASCADE,
    model_name TEXT NOT NULL,
    PRIMARY KEY (rule_id, model_name)
);

-- Keys domain
CREATE TABLE IF NOT EXISTS provider_keys (
    id               TEXT PRIMARY KEY,
    provider         TEXT NOT NULL,
    name             TEXT NOT NULL,
    key_prefix       TEXT NOT NULL,
    secret_key       TEXT NOT NULL,
    relay_channel_id INT NOT NULL DEFAULT 0,
    status           TEXT NOT NULL,
    balance          NUMERIC(18, 6),
    last_used        TIMESTAMPTZ,
    rotate_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS platform_keys (
    id                TEXT PRIMARY KEY,
    name              TEXT NOT NULL,
    key_prefix        TEXT NOT NULL,
    full_key          TEXT,
    member_id         TEXT REFERENCES members (id) ON DELETE SET NULL,
    member_name       TEXT,
    app_name          TEXT,
    budget_group_id   TEXT REFERENCES budget_groups (id) ON DELETE SET NULL,
    budget_group_name TEXT,
    status            TEXT NOT NULL,
    quota             NUMERIC(18, 6) NOT NULL DEFAULT 0,
    used              NUMERIC(18, 6) NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at        TIMESTAMPTZ,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_platform_keys_member ON platform_keys (member_id);
CREATE INDEX IF NOT EXISTS idx_platform_keys_budget_group ON platform_keys (budget_group_id);

CREATE TABLE IF NOT EXISTS platform_key_models (
    platform_key_id TEXT NOT NULL REFERENCES platform_keys (id) ON DELETE CASCADE,
    model_name      TEXT NOT NULL,
    PRIMARY KEY (platform_key_id, model_name)
);

CREATE TABLE IF NOT EXISTS key_approvals (
    id              TEXT PRIMARY KEY,
    type            TEXT NOT NULL,
    applicant       TEXT NOT NULL,
    applicant_id    TEXT NOT NULL REFERENCES members (id),
    department      TEXT NOT NULL,
    reason          TEXT NOT NULL,
    requested_quota NUMERIC(18, 6) NOT NULL DEFAULT 0,
    status          TEXT NOT NULL,
    approver        TEXT,
    reject_reason   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_key_approvals_status ON key_approvals (status, created_at DESC);

CREATE TABLE IF NOT EXISTS key_approval_models (
    approval_id TEXT NOT NULL REFERENCES key_approvals (id) ON DELETE CASCADE,
    model_name  TEXT NOT NULL,
    PRIMARY KEY (approval_id, model_name)
);

-- Audit domain
CREATE TABLE IF NOT EXISTS audit_settings (
    id                        INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    content_retention_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS operation_logs (
    id          TEXT PRIMARY KEY,
    action      TEXT NOT NULL,
    operator    TEXT NOT NULL,
    operator_id TEXT NOT NULL,
    target      TEXT NOT NULL DEFAULT '',
    detail      TEXT NOT NULL DEFAULT '',
    ip          TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_operation_logs_created ON operation_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_operation_logs_operator ON operation_logs (operator_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_operation_logs_action ON operation_logs (action, created_at DESC);

CREATE TABLE IF NOT EXISTS call_logs (
    id             TEXT PRIMARY KEY,
    caller         TEXT NOT NULL,
    caller_id      TEXT NOT NULL,
    caller_type    TEXT NOT NULL,
    model          TEXT NOT NULL,
    provider       TEXT NOT NULL,
    input_tokens   NUMERIC(18, 2) NOT NULL DEFAULT 0,
    output_tokens  NUMERIC(18, 2) NOT NULL DEFAULT 0,
    latency_ms     NUMERIC(18, 2) NOT NULL DEFAULT 0,
    status         TEXT NOT NULL,
    cost           NUMERIC(18, 6) NOT NULL DEFAULT 0,
    input_preview  TEXT NOT NULL DEFAULT '',
    output_preview TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_call_logs_created ON call_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_call_logs_caller ON call_logs (caller_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_call_logs_model ON call_logs (model, created_at DESC);

-- Infrastructure (unchanged)
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

CREATE TABLE IF NOT EXISTS usage_buckets (
    bucket_start   TIMESTAMPTZ NOT NULL,
    department_id  TEXT NOT NULL,
    member_id      TEXT NOT NULL DEFAULT '',
    model          TEXT NOT NULL,
    cost_cny       NUMERIC(18, 6) NOT NULL DEFAULT 0,
    call_count     INT NOT NULL DEFAULT 0,
    input_tokens   BIGINT NOT NULL DEFAULT 0,
    output_tokens  BIGINT NOT NULL DEFAULT 0,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (bucket_start, department_id, member_id, model)
);

CREATE INDEX IF NOT EXISTS idx_usage_buckets_dept_time ON usage_buckets (department_id, bucket_start);
CREATE INDEX IF NOT EXISTS idx_usage_buckets_time ON usage_buckets (bucket_start);

CREATE TABLE IF NOT EXISTS notification_log (
    id           TEXT PRIMARY KEY,
    channel      TEXT NOT NULL,
    event_type   TEXT NOT NULL,
    recipient    TEXT,
    payload      JSONB NOT NULL,
    status       TEXT NOT NULL DEFAULT 'sent',
    error        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
