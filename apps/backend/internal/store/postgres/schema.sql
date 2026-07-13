CREATE TABLE IF NOT EXISTS currencies (
    currency         CHAR(3) PRIMARY KEY,
    points_per_unit  BIGINT NOT NULL CHECK (points_per_unit > 0),
    enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Global: companies
CREATE TABLE IF NOT EXISTS companies (
    id                        BIGINT PRIMARY KEY,
    slug                      TEXT NOT NULL UNIQUE,
    name                      TEXT NOT NULL,
    status                    TEXT NOT NULL DEFAULT 'active',
    root_dept_id              TEXT,
    newapi_wallet_user_id     BIGINT,
    package_id                TEXT,
    authz_revision            BIGINT NOT NULL DEFAULT 0,
    billing_currency          CHAR(3) NOT NULL DEFAULT 'CNY' REFERENCES currencies (currency),
    fifo_head_lot_id          TEXT,
    wallet_remain             NUMERIC(28, 10) NOT NULL DEFAULT 0,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS company_invites (
    id           TEXT PRIMARY KEY,
    company_id   BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    email        TEXT NOT NULL,
    role         TEXT NOT NULL DEFAULT 'super_admin',
    invite_code  TEXT NOT NULL UNIQUE,
    expires_at   TIMESTAMPTZ NOT NULL,
    accepted_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_company_invites_invite_code ON company_invites (invite_code);
CREATE INDEX IF NOT EXISTS idx_company_invites_company ON company_invites (company_id);

CREATE TABLE IF NOT EXISTS platform_operators (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS company_recharge_orders (
    id               TEXT PRIMARY KEY,
    company_id       BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    amount           NUMERIC(18, 6) NOT NULL DEFAULT 0,
    currency         CHAR(3) NOT NULL,
    points_per_unit  BIGINT NOT NULL,
    points_granted   NUMERIC(28, 10) NOT NULL,
    source           TEXT NOT NULL,
    lot_kind         TEXT NOT NULL,
    idempotency_key  TEXT,
    status           TEXT NOT NULL DEFAULT 'pending',
    display_order_id TEXT NOT NULL DEFAULT '',
    payment_method   TEXT NOT NULL DEFAULT '',
    invoice_status   TEXT NOT NULL DEFAULT 'none',
    created_by       TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (lot_kind IN ('paid', 'gift', 'adjust', 'overdraft')),
    CHECK (points_granted > 0),
    CHECK (
        (lot_kind = 'paid' AND amount > 0)
        OR (lot_kind IN ('gift', 'overdraft') AND amount = 0)
        OR (lot_kind = 'adjust')
    )
);

CREATE TABLE IF NOT EXISTS company_recharge_lots (
    id                 TEXT PRIMARY KEY,
    company_id         BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    recharge_order_id  TEXT NOT NULL UNIQUE REFERENCES company_recharge_orders (id),
    billing_currency   CHAR(3) NOT NULL,
    lot_kind           TEXT NOT NULL,
    amount_display     NUMERIC(18, 6) NOT NULL,
    points_granted     NUMERIC(28, 10) NOT NULL,
    points_remaining   NUMERIC(28, 10) NOT NULL,
    unit_price_display NUMERIC(28, 18) NOT NULL,
    status             TEXT NOT NULL DEFAULT 'active',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (lot_kind IN ('paid', 'gift', 'adjust', 'overdraft')),
    CHECK (points_granted > 0),
    CHECK (points_remaining >= 0 AND points_remaining <= points_granted),
    CHECK (unit_price_display >= 0),
    CHECK (
        (lot_kind IN ('gift', 'overdraft') AND amount_display = 0 AND unit_price_display = 0)
        OR (lot_kind = 'paid' AND amount_display > 0 AND unit_price_display > 0)
        OR (lot_kind = 'adjust' AND amount_display >= 0 AND unit_price_display >= 0)
    )
);

CREATE INDEX IF NOT EXISTS idx_recharge_lots_fifo
    ON company_recharge_lots (company_id, created_at)
    WHERE status = 'active' AND points_remaining > 0;

CREATE INDEX IF NOT EXISTS idx_company_recharge_orders_company ON company_recharge_orders (company_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_company_recharge_orders_idempotency
    ON company_recharge_orders (company_id, idempotency_key) WHERE idempotency_key IS NOT NULL;

-- Org domain
CREATE TABLE IF NOT EXISTS permissions (
    id    TEXT PRIMARY KEY,
    name  TEXT NOT NULL,
    grp   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS roles (
    id         TEXT NOT NULL,
    company_id BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    type       TEXT NOT NULL,
    PRIMARY KEY (company_id, id),
    UNIQUE (company_id, name)
);

CREATE TABLE IF NOT EXISTS role_permission_grants (
    company_id    BIGINT NOT NULL,
    role_id       TEXT NOT NULL,
    permission_id TEXT NOT NULL,
    PRIMARY KEY (company_id, role_id, permission_id),
    FOREIGN KEY (company_id, role_id) REFERENCES roles (company_id, id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS models (
    model_id     BIGINT GENERATED BY DEFAULT AS IDENTITY (START WITH 100 INCREMENT BY 1) PRIMARY KEY,
    company_id   BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    provider     TEXT NOT NULL,
    type         TEXT NOT NULL,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    endpoint     TEXT,
    input_price  NUMERIC(18, 8) NOT NULL DEFAULT 0,
    output_price NUMERIC(18, 8) NOT NULL DEFAULT 0,
    max_context  INT NOT NULL DEFAULT 0,
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, provider, type)
);

CREATE INDEX IF NOT EXISTS idx_models_company ON models (company_id);
CREATE INDEX IF NOT EXISTS idx_models_company_type ON models (company_id, type);

CREATE TABLE IF NOT EXISTS org_nodes (
    id                TEXT NOT NULL,
    company_id        BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    parent_id         TEXT,
    path              LTREE NOT NULL,
    external_id       TEXT,
    source            TEXT,
    manager_id        TEXT,
    sort_order        INT NOT NULL DEFAULT 0,
    default_model_id  BIGINT,
    fallback_model_id BIGINT,
    routing_inherited BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, id),
    FOREIGN KEY (company_id, parent_id) REFERENCES org_nodes (company_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (default_model_id) REFERENCES models (model_id) ON DELETE RESTRICT,
    FOREIGN KEY (fallback_model_id) REFERENCES models (model_id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_org_nodes_parent ON org_nodes (company_id, parent_id);
CREATE INDEX IF NOT EXISTS idx_org_nodes_path ON org_nodes USING GIST (path);
CREATE UNIQUE INDEX IF NOT EXISTS idx_org_nodes_external
    ON org_nodes (company_id, external_id) WHERE external_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS org_node_budget (
    company_id        BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    node_id           TEXT NOT NULL,
    budget            NUMERIC(18, 6) NOT NULL DEFAULT 0,
    reserved_pool     NUMERIC(18, 6),
    period            TEXT NOT NULL CHECK (period IN ('monthly')),
    member_avg_budget NUMERIC(18, 6) NOT NULL DEFAULT 0,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, node_id),
    FOREIGN KEY (company_id, node_id) REFERENCES org_nodes (company_id, id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS members (
    id              TEXT NOT NULL,
    company_id      BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    phone           TEXT NOT NULL DEFAULT '',
    email           TEXT NOT NULL DEFAULT '',
    department_id   TEXT NOT NULL,
    status          TEXT NOT NULL,
    source          TEXT NOT NULL DEFAULT '',
    external_id     TEXT,
    password_hash   TEXT,
    personal_budget NUMERIC(18, 6) NOT NULL DEFAULT 5000,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, id),
    FOREIGN KEY (company_id, department_id) REFERENCES org_nodes (company_id, id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_members_department ON members (company_id, department_id);
CREATE INDEX IF NOT EXISTS idx_members_status ON members (company_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_members_email_company ON members (company_id, email) WHERE email <> '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_members_external
    ON members (company_id, external_id) WHERE external_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS member_roles (
    company_id BIGINT NOT NULL,
    member_id  TEXT NOT NULL,
    role_id    TEXT NOT NULL,
    PRIMARY KEY (company_id, member_id, role_id),
    FOREIGN KEY (company_id, member_id) REFERENCES members (company_id, id) ON DELETE CASCADE,
    FOREIGN KEY (company_id, role_id) REFERENCES roles (company_id, id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_member_roles_role ON member_roles (company_id, role_id);

CREATE TABLE IF NOT EXISTS org_integration (
    company_id                  BIGINT PRIMARY KEY REFERENCES companies (id) ON DELETE CASCADE,
    platform                    TEXT,
    connected                   BOOLEAN NOT NULL DEFAULT FALSE,
    last_import                 TIMESTAMPTZ,
    last_import_ok              INT,
    last_import_fail            INT,
    enabled                     BOOLEAN NOT NULL DEFAULT FALSE,
    start_time                  TEXT NOT NULL DEFAULT '',
    frequency_hours             INT NOT NULL DEFAULT 24,
    delete_member_threshold     INT NOT NULL DEFAULT 0,
    delete_department_threshold INT NOT NULL DEFAULT 0,
    notify_phone                BOOLEAN NOT NULL DEFAULT FALSE,
    notify_email                BOOLEAN NOT NULL DEFAULT FALSE,
    notify_im                   BOOLEAN NOT NULL DEFAULT FALSE,
    encrypted_credential        BYTEA,
    field_mappings              JSONB NOT NULL DEFAULT '[]',
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS org_sync_logs (
    id         TEXT NOT NULL,
    company_id BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    time       TIMESTAMPTZ NOT NULL,
    type       TEXT NOT NULL,
    result     TEXT NOT NULL,
    detail     TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (company_id, id)
);

CREATE INDEX IF NOT EXISTS idx_org_sync_logs_company_time ON org_sync_logs (company_id, time DESC);

CREATE TABLE IF NOT EXISTS org_import_failures (
    id          TEXT NOT NULL,
    company_id  BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    employee_id TEXT NOT NULL DEFAULT '',
    reason      TEXT NOT NULL,
    PRIMARY KEY (company_id, id)
);

-- Budget domain
CREATE TABLE IF NOT EXISTS projects (
    id                   TEXT NOT NULL,
    company_id           BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    name                 TEXT NOT NULL,
    budget               NUMERIC(18, 6) NOT NULL DEFAULT 0,
    owner_department_id  TEXT NOT NULL,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, id),
    FOREIGN KEY (company_id, owner_department_id) REFERENCES org_nodes (company_id, id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS project_members (
    company_id BIGINT NOT NULL,
    project_id TEXT NOT NULL,
    member_id  TEXT NOT NULL,
    member_budget DOUBLE PRECISION NOT NULL DEFAULT 0,
    PRIMARY KEY (company_id, project_id, member_id),
    FOREIGN KEY (company_id, project_id) REFERENCES projects (company_id, id) ON DELETE CASCADE,
    FOREIGN KEY (company_id, member_id) REFERENCES members (company_id, id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS overrun_policy (
    company_id    BIGINT PRIMARY KEY REFERENCES companies (id) ON DELETE CASCADE,
    thresholds    INT[] NOT NULL DEFAULT '{}',
    notify_email  BOOLEAN NOT NULL DEFAULT FALSE,
    notify_phone  BOOLEAN NOT NULL DEFAULT FALSE,
    notify_im     BOOLEAN NOT NULL DEFAULT FALSE,
    block_message TEXT NOT NULL DEFAULT '',
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alert_rules (
    id         TEXT NOT NULL,
    company_id BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    node_id    TEXT NOT NULL,
    thresholds INT[] NOT NULL DEFAULT '{}',
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, id),
    FOREIGN KEY (company_id, node_id) REFERENCES org_nodes (company_id, id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS alert_rule_notify_roles (
    company_id BIGINT NOT NULL,
    rule_id    TEXT NOT NULL,
    role_id    TEXT NOT NULL,
    PRIMARY KEY (company_id, rule_id, role_id),
    FOREIGN KEY (company_id, rule_id) REFERENCES alert_rules (company_id, id) ON DELETE CASCADE,
    FOREIGN KEY (company_id, role_id) REFERENCES roles (company_id, id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS budget_approvals (
    id              TEXT NOT NULL,
    company_id      BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    applicant_id    TEXT,
    applicant_name  TEXT NOT NULL,
    department_name TEXT NOT NULL,
    amount          NUMERIC(18, 6) NOT NULL,
    reason          TEXT NOT NULL,
    status          TEXT NOT NULL,
    reject_reason   TEXT,
    created_at      TIMESTAMPTZ NOT NULL,
    resolved_at     TIMESTAMPTZ,
    PRIMARY KEY (company_id, id)
);

CREATE INDEX IF NOT EXISTS idx_budget_approvals_status
    ON budget_approvals (company_id, status, created_at DESC);

CREATE TABLE IF NOT EXISTS model_capabilities (
    model_id   BIGINT NOT NULL,
    capability TEXT NOT NULL,
    PRIMARY KEY (model_id, capability),
    FOREIGN KEY (model_id) REFERENCES models (model_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS model_allowlist (
    company_id BIGINT NOT NULL,
    owner_type TEXT NOT NULL,
    owner_id   TEXT NOT NULL,
    model_id   BIGINT NOT NULL,
    PRIMARY KEY (company_id, owner_type, owner_id, model_id),
    FOREIGN KEY (model_id) REFERENCES models (model_id) ON DELETE CASCADE,
    CONSTRAINT chk_model_allowlist_owner_type
        CHECK (owner_type IN ('platform_key', 'org_node', 'key_approval'))
);

CREATE INDEX IF NOT EXISTS idx_model_allowlist_owner
    ON model_allowlist (company_id, owner_type, owner_id);

-- Keys domain
CREATE TABLE IF NOT EXISTS provider_keys (
    id               TEXT PRIMARY KEY,
    provider         TEXT NOT NULL,
    name             TEXT NOT NULL,
    key_prefix       TEXT NOT NULL,
    secret_key       TEXT NOT NULL,
    newapi_channel_id INT NOT NULL DEFAULT 0,
    status           TEXT NOT NULL,
    balance          NUMERIC(18, 6),
    last_used        TIMESTAMPTZ,
    rotate_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_provider_keys_status ON provider_keys (status);

CREATE TABLE IF NOT EXISTS platform_keys (
    id              TEXT NOT NULL,
    company_id      BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    key_prefix      TEXT NOT NULL,
    key_hash        TEXT NOT NULL,
    member_id       TEXT,
    project_id      TEXT,
    scope           TEXT NOT NULL,
    status          TEXT NOT NULL,
    budget          NUMERIC(18, 6) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    gateway_soft_remain  NUMERIC(18, 6),
    gateway_soft_at      TIMESTAMPTZ,
    gateway_soft_version BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (company_id, id),
    FOREIGN KEY (company_id, member_id) REFERENCES members (company_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (company_id, project_id) REFERENCES projects (company_id, id) ON DELETE RESTRICT,
    CONSTRAINT chk_platform_keys_scope
        CHECK (scope IN ('member', 'project', 'project_member'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_keys_key_hash ON platform_keys (key_hash);
CREATE INDEX IF NOT EXISTS idx_platform_keys_scope ON platform_keys (company_id, scope);
CREATE INDEX IF NOT EXISTS idx_platform_keys_member ON platform_keys (company_id, member_id) WHERE member_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_platform_keys_project ON platform_keys (company_id, project_id) WHERE project_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_platform_keys_active ON platform_keys (company_id, status) WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_platform_keys_soft_stale
    ON platform_keys (company_id, gateway_soft_at)
    WHERE status = 'active' AND gateway_soft_remain IS NOT NULL;

COMMENT ON COLUMN platform_keys.gateway_soft_remain IS
    'Projector/Reconcile maintained minimum budget remain; projection, not SSOT';
COMMENT ON COLUMN platform_keys.gateway_soft_version IS
    'Monotonic soft-summary version shared with the optional Redis cache';

CREATE TABLE IF NOT EXISTS key_approvals (
    id              TEXT NOT NULL,
    company_id      BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    type            TEXT NOT NULL,
    applicant       TEXT NOT NULL,
    applicant_id    TEXT NOT NULL,
    department      TEXT NOT NULL,
    reason          TEXT NOT NULL,
    requested_budget NUMERIC(18, 6) NOT NULL DEFAULT 0,
    status          TEXT NOT NULL,
    approver        TEXT,
    reject_reason   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at     TIMESTAMPTZ,
    PRIMARY KEY (company_id, id)
);

CREATE INDEX IF NOT EXISTS idx_key_approvals_company_status
    ON key_approvals (company_id, status, created_at DESC);

-- Audit domain
CREATE TABLE IF NOT EXISTS audit_settings (
    company_id                BIGINT PRIMARY KEY REFERENCES companies (id) ON DELETE CASCADE,
    content_retention_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS operation_logs (
    id          TEXT NOT NULL,
    company_id  BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    action      TEXT NOT NULL,
    operator    TEXT NOT NULL,
    operator_id TEXT NOT NULL,
    target      TEXT NOT NULL DEFAULT '',
    detail      TEXT NOT NULL DEFAULT '',
    ip          TEXT NOT NULL DEFAULT '',
    actor_type  TEXT NOT NULL DEFAULT 'member',
    created_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (company_id, id, created_at)
) PARTITION BY RANGE (created_at);

CREATE INDEX IF NOT EXISTS idx_operation_logs_created ON operation_logs (company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_operation_logs_operator ON operation_logs (company_id, operator_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_operation_logs_action ON operation_logs (company_id, action, created_at DESC);

CREATE TABLE IF NOT EXISTS usage_ledger (
    id               TEXT NOT NULL,
    company_id       BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    event_type       TEXT NOT NULL,
    idempotency_key  TEXT NOT NULL,
    segment_index    INT NOT NULL DEFAULT 0,
    lot_id           TEXT NOT NULL REFERENCES company_recharge_lots (id),
    amount           NUMERIC(28, 10) NOT NULL DEFAULT 0,
    display_amount   NUMERIC(18, 6) NOT NULL DEFAULT 0,
    billing_currency CHAR(3) NOT NULL,
    department_id    TEXT NOT NULL,
    member_id        TEXT,
    project_id       TEXT,
    platform_key_id  TEXT NOT NULL,
    platform_key_scope TEXT NOT NULL CHECK (platform_key_scope IN ('member', 'project', 'project_member')),
    source           TEXT NOT NULL,
    occurred_at      TIMESTAMPTZ NOT NULL,
    period_key       TEXT NOT NULL,
    model            TEXT NOT NULL,
    input_tokens     BIGINT NOT NULL DEFAULT 0,
    output_tokens    BIGINT NOT NULL DEFAULT 0,
    call_detail      JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, id, occurred_at),
    UNIQUE (company_id, idempotency_key, lot_id, occurred_at)
) PARTITION BY RANGE (occurred_at);

CREATE INDEX IF NOT EXISTS idx_usage_ledger_call_settled_occurred
    ON usage_ledger (company_id, occurred_at DESC)
    WHERE event_type = 'call_settled';

CREATE INDEX IF NOT EXISTS idx_usage_ledger_projector_cursor
    ON usage_ledger (company_id, occurred_at ASC, id ASC)
    WHERE event_type = 'call_settled';

CREATE INDEX IF NOT EXISTS idx_usage_ledger_dept_occurred
    ON usage_ledger (company_id, department_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_usage_ledger_platform_key_occurred
    ON usage_ledger (company_id, platform_key_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS usage_buckets (
    company_id    BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    bucket_start  TIMESTAMPTZ NOT NULL,
    department_id TEXT NOT NULL,
    member_id     TEXT,
    member_scope  TEXT GENERATED ALWAYS AS (COALESCE(member_id, '')) STORED,
    model         TEXT NOT NULL,
    cost          NUMERIC(28, 10) NOT NULL DEFAULT 0,
    call_count    INT NOT NULL DEFAULT 0,
    input_tokens  BIGINT NOT NULL DEFAULT 0,
    output_tokens BIGINT NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, bucket_start, department_id, member_scope, model)
) PARTITION BY RANGE (bucket_start);

CREATE INDEX IF NOT EXISTS idx_usage_buckets_dept_time ON usage_buckets (company_id, department_id, bucket_start);
CREATE INDEX IF NOT EXISTS idx_usage_buckets_time ON usage_buckets (company_id, bucket_start);
CREATE INDEX IF NOT EXISTS idx_usage_buckets_member_time ON usage_buckets (company_id, member_id, bucket_start) WHERE member_id IS NOT NULL;


-- Runtime
CREATE TABLE IF NOT EXISTS platform_key_mappings (
    company_id              BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    platform_key_id         TEXT NOT NULL,
    newapi_key_id           BIGINT,
    newapi_group            TEXT NOT NULL,
    sync_status             TEXT NOT NULL DEFAULT 'pending',
    synced_at               TIMESTAMPTZ,
    newapi_key_remain_quota BIGINT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, platform_key_id),
    FOREIGN KEY (company_id, platform_key_id) REFERENCES platform_keys (company_id, id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_key_mappings_key
    ON platform_key_mappings (company_id, newapi_key_id) WHERE newapi_key_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_platform_key_mappings_sync_pending
    ON platform_key_mappings (company_id, sync_status) WHERE sync_status = 'pending';

CREATE TABLE IF NOT EXISTS budget_consumed (
    company_id BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    axis_kind  TEXT NOT NULL CHECK (axis_kind IN ('project', 'platform_key', 'member')),
    axis_id    TEXT NOT NULL,
    period_key TEXT NOT NULL,
    consumed   NUMERIC(18, 6) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, axis_kind, axis_id, period_key)
);

CREATE INDEX IF NOT EXISTS idx_budget_consumed_overrun
    ON budget_consumed (company_id, axis_kind, period_key);

CREATE TABLE IF NOT EXISTS budget_projection_progress (
    company_id       BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    stream           TEXT NOT NULL DEFAULT 'ledger_consumed',
    last_occurred_at TIMESTAMPTZ,
    last_ledger_id   TEXT,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, stream)
);

CREATE TABLE IF NOT EXISTS dashboard_projection_progress (
    company_id       BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    stream           TEXT NOT NULL DEFAULT 'dashboard_buckets',
    last_occurred_at TIMESTAMPTZ,
    last_ledger_id   TEXT,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, stream)
);

CREATE TABLE IF NOT EXISTS tenant_background_state (
    company_id                  BIGINT PRIMARY KEY REFERENCES companies (id) ON DELETE CASCADE,
    next_org_sync_at            TIMESTAMPTZ,
    last_org_sync_at            TIMESTAMPTZ,
    last_rebalanced_period      VARCHAR(7) NOT NULL DEFAULT '',
    last_budget_reconcile_at    TIMESTAMPTZ,
    last_dashboard_reconcile_at TIMESTAMPTZ,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tbs_org_sync_due
    ON tenant_background_state (next_org_sync_at)
    WHERE next_org_sync_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tbs_rebalance_period
    ON tenant_background_state (last_rebalanced_period);

CREATE TABLE IF NOT EXISTS scheduler_locks (
    lock_name   TEXT PRIMARY KEY,
    holder      TEXT NOT NULL,
    lease_until TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notification_log (
    id         TEXT PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    channel    TEXT NOT NULL,
    event_type TEXT NOT NULL,
    recipient  TEXT,
    payload    JSONB NOT NULL,
    status     TEXT NOT NULL DEFAULT 'sent',
    error      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_log_company_time
    ON notification_log (company_id, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_unique_name ON projects(company_id, name);
