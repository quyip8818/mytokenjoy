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
    balance_point             NUMERIC(28, 10) NOT NULL DEFAULT 0,
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
    newapi_sync_ref  TEXT,
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
    model_id     BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
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
    budget            NUMERIC(18, 6) NOT NULL DEFAULT 0,
    reserved_pool     NUMERIC(18, 6),
    period            TEXT NOT NULL CHECK (period IN ('monthly')),
    default_model_id  BIGINT,
    fallback_model_id BIGINT,
    routing_inherited BOOLEAN NOT NULL DEFAULT FALSE,
    member_avg_budget NUMERIC(18, 6) NOT NULL DEFAULT 0,
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
CREATE TABLE IF NOT EXISTS budget_groups (
    id         TEXT NOT NULL,
    company_id BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    budget     NUMERIC(18, 6) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, id)
);

CREATE TABLE IF NOT EXISTS budget_group_members (
    company_id BIGINT NOT NULL,
    group_id   TEXT NOT NULL,
    member_id  TEXT NOT NULL,
    PRIMARY KEY (company_id, group_id, member_id),
    FOREIGN KEY (company_id, group_id) REFERENCES budget_groups (company_id, id) ON DELETE CASCADE,
    FOREIGN KEY (company_id, member_id) REFERENCES members (company_id, id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS budget_group_departments (
    company_id    BIGINT NOT NULL,
    group_id      TEXT NOT NULL,
    department_id TEXT NOT NULL,
    PRIMARY KEY (company_id, group_id, department_id),
    FOREIGN KEY (company_id, group_id) REFERENCES budget_groups (company_id, id) ON DELETE CASCADE,
    FOREIGN KEY (company_id, department_id) REFERENCES org_nodes (company_id, id) ON DELETE CASCADE
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
    budget_group_id TEXT,
    status          TEXT NOT NULL,
    budget          NUMERIC(18, 6) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, id),
    FOREIGN KEY (company_id, member_id) REFERENCES members (company_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (company_id, budget_group_id) REFERENCES budget_groups (company_id, id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_keys_key_hash ON platform_keys (key_hash);
CREATE INDEX IF NOT EXISTS idx_platform_keys_member ON platform_keys (company_id, member_id) WHERE member_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_platform_keys_budget_group ON platform_keys (company_id, budget_group_id) WHERE budget_group_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_platform_keys_active ON platform_keys (company_id, status) WHERE status = 'active';

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
    budget_group_id  TEXT,
    platform_key_id  TEXT NOT NULL,
    source           TEXT NOT NULL,
    occurred_at      TIMESTAMPTZ NOT NULL,
    period_key       TEXT NOT NULL,
    model            TEXT NOT NULL,
    input_tokens     BIGINT NOT NULL DEFAULT 0,
    output_tokens    BIGINT NOT NULL DEFAULT 0,
    call_status      TEXT,
    caller_id        TEXT,
    caller_name      TEXT,
    preview_snippet  TEXT,
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

CREATE INDEX IF NOT EXISTS idx_usage_ledger_caller_occurred
    ON usage_ledger (company_id, caller_id, occurred_at DESC)
    WHERE caller_id IS NOT NULL;

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
    axis_kind  TEXT NOT NULL CHECK (axis_kind IN ('org_node', 'budget_group', 'platform_key', 'member')),
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

-- River job queue (github.com/riverqueue/river v0.40.0, line main migrations 001-007)
CREATE TABLE river_migration(
  id bigserial PRIMARY KEY,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  version bigint NOT NULL,
  CONSTRAINT version CHECK (version >= 1)
);

CREATE UNIQUE INDEX ON river_migration USING btree(version);

CREATE TYPE river_job_state AS ENUM(
  'available',
  'cancelled',
  'completed',
  'discarded',
  'retryable',
  'running',
  'scheduled'
);

CREATE TABLE river_job(
  id bigserial PRIMARY KEY,
  state river_job_state NOT NULL DEFAULT 'available',
  attempt smallint NOT NULL DEFAULT 0,
  max_attempts smallint NOT NULL,
  attempted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  finalized_at timestamptz,
  scheduled_at timestamptz NOT NULL DEFAULT NOW(),
  priority smallint NOT NULL DEFAULT 1,
  args jsonb,
  attempted_by text[],
  errors jsonb[],
  kind text NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}',
  queue text NOT NULL DEFAULT 'default',
  tags varchar(255)[],
  CONSTRAINT finalized_or_finalized_at_null CHECK ((state IN ('cancelled', 'completed', 'discarded') AND finalized_at IS NOT NULL) OR finalized_at IS NULL),
  CONSTRAINT max_attempts_is_positive CHECK (max_attempts > 0),
  CONSTRAINT priority_in_range CHECK (priority >= 1 AND priority <= 4),
  CONSTRAINT queue_length CHECK (char_length(queue) > 0 AND char_length(queue) < 128),
  CONSTRAINT kind_length CHECK (char_length(kind) > 0 AND char_length(kind) < 128)
);

CREATE INDEX river_job_kind ON river_job USING btree(kind);
CREATE INDEX river_job_state_and_finalized_at_index ON river_job USING btree(state, finalized_at) WHERE finalized_at IS NOT NULL;
CREATE INDEX river_job_prioritized_fetching_index ON river_job USING btree(state, queue, priority, scheduled_at, id);
CREATE INDEX river_job_args_index ON river_job USING GIN(args);
CREATE INDEX river_job_metadata_index ON river_job USING GIN(metadata);

CREATE OR REPLACE FUNCTION river_job_notify()
  RETURNS TRIGGER
  AS $$
DECLARE
  payload json;
BEGIN
  IF NEW.state = 'available' THEN
    payload = json_build_object('queue', NEW.queue);
    PERFORM pg_notify('river_insert', payload::text);
  END IF;
  RETURN NULL;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER river_notify
  AFTER INSERT ON river_job
  FOR EACH ROW
  EXECUTE PROCEDURE river_job_notify();

CREATE UNLOGGED TABLE river_leader(
    elected_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    leader_id text NOT NULL,
    name text PRIMARY KEY,
    CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128),
    CONSTRAINT leader_id_length CHECK (char_length(leader_id) > 0 AND char_length(leader_id) < 128)
);

ALTER TABLE river_job ALTER COLUMN tags SET DEFAULT '{}';
UPDATE river_job SET tags = '{}' WHERE tags IS NULL;
ALTER TABLE river_job ALTER COLUMN tags SET NOT NULL;

ALTER TABLE river_job ALTER COLUMN args SET DEFAULT '{}';
UPDATE river_job SET args = '{}' WHERE args IS NULL;
ALTER TABLE river_job ALTER COLUMN args SET NOT NULL;
ALTER TABLE river_job ALTER COLUMN args DROP DEFAULT;

ALTER TABLE river_job ALTER COLUMN metadata SET DEFAULT '{}';
UPDATE river_job SET metadata = '{}' WHERE metadata IS NULL;
ALTER TABLE river_job ALTER COLUMN metadata SET NOT NULL;

ALTER TYPE river_job_state ADD VALUE IF NOT EXISTS 'pending' AFTER 'discarded';

ALTER TABLE river_job DROP CONSTRAINT finalized_or_finalized_at_null;
ALTER TABLE river_job ADD CONSTRAINT finalized_or_finalized_at_null CHECK (
    (finalized_at IS NULL AND state NOT IN ('cancelled', 'completed', 'discarded')) OR
    (finalized_at IS NOT NULL AND state IN ('cancelled', 'completed', 'discarded'))
);

DROP TRIGGER river_notify ON river_job;
DROP FUNCTION river_job_notify;

CREATE TABLE river_queue (
    name text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}' ::jsonb,
    paused_at timestamptz,
    updated_at timestamptz NOT NULL
);

ALTER TABLE river_leader
    ALTER COLUMN name SET DEFAULT 'default',
    DROP CONSTRAINT name_length,
    ADD CONSTRAINT name_length CHECK (name = 'default');

DO
$body$
BEGIN
    IF (SELECT to_regclass('river_migration') IS NOT NULL) THEN
        ALTER TABLE river_migration RENAME TO river_migration_old;
        CREATE TABLE river_migration(
            line TEXT NOT NULL,
            version bigint NOT NULL,
            created_at timestamptz NOT NULL DEFAULT NOW(),
            CONSTRAINT line_length CHECK (char_length(line) > 0 AND char_length(line) < 128),
            CONSTRAINT version_gte_1 CHECK (version >= 1),
            PRIMARY KEY (line, version)
        );
        INSERT INTO river_migration (created_at, line, version)
        SELECT created_at, 'main', version FROM river_migration_old;
        DROP TABLE river_migration_old;
    END IF;
END;
$body$
LANGUAGE 'plpgsql';

ALTER TABLE river_job ADD COLUMN IF NOT EXISTS unique_key bytea;
CREATE UNIQUE INDEX IF NOT EXISTS river_job_kind_unique_key_idx ON river_job (kind, unique_key) WHERE unique_key IS NOT NULL;

CREATE UNLOGGED TABLE river_client (
    id text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}',
    paused_at timestamptz,
    updated_at timestamptz NOT NULL,
    CONSTRAINT name_length CHECK (char_length(id) > 0 AND char_length(id) < 128)
);

CREATE UNLOGGED TABLE river_client_queue (
    river_client_id text NOT NULL REFERENCES river_client (id) ON DELETE CASCADE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    max_workers bigint NOT NULL DEFAULT 0,
    metadata jsonb NOT NULL DEFAULT '{}',
    num_jobs_completed bigint NOT NULL DEFAULT 0,
    num_jobs_running bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (river_client_id, name),
    CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128),
    CONSTRAINT num_jobs_completed_zero_or_positive CHECK (num_jobs_completed >= 0),
    CONSTRAINT num_jobs_running_zero_or_positive CHECK (num_jobs_running >= 0)
);

CREATE OR REPLACE FUNCTION river_job_state_in_bitmask(bitmask BIT(8), state river_job_state)
RETURNS boolean
LANGUAGE SQL
IMMUTABLE
AS $$
    SELECT CASE state
        WHEN 'available' THEN get_bit(bitmask, 7)
        WHEN 'cancelled' THEN get_bit(bitmask, 6)
        WHEN 'completed' THEN get_bit(bitmask, 5)
        WHEN 'discarded' THEN get_bit(bitmask, 4)
        WHEN 'pending'   THEN get_bit(bitmask, 3)
        WHEN 'retryable' THEN get_bit(bitmask, 2)
        WHEN 'running'   THEN get_bit(bitmask, 1)
        WHEN 'scheduled' THEN get_bit(bitmask, 0)
        ELSE 0
    END = 1;
$$;

ALTER TABLE river_job ADD COLUMN IF NOT EXISTS unique_states BIT(8);
CREATE UNIQUE INDEX IF NOT EXISTS river_job_unique_idx ON river_job (unique_key)
    WHERE unique_key IS NOT NULL
      AND unique_states IS NOT NULL
      AND river_job_state_in_bitmask(unique_states, state);
DROP INDEX IF EXISTS river_job_kind_unique_key_idx;

CREATE TABLE river_notification (
    id bigserial PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now(),
    payload text NOT NULL,
    topic text NOT NULL,
    CONSTRAINT topic_length CHECK (length(topic) > 0 AND length(topic) < 128)
);

CREATE INDEX river_notification_created_at_idx ON river_notification (created_at);
CREATE INDEX river_notification_topic_id_idx ON river_notification (topic, id);

DROP TABLE IF EXISTS river_client_queue;
DROP TABLE IF EXISTS river_client;

ALTER TABLE river_job ALTER COLUMN max_attempts SET DEFAULT 25;
ALTER TABLE river_queue ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

INSERT INTO river_migration (line, version, created_at)
SELECT 'main', v, NOW()
FROM unnest(ARRAY[1::bigint, 2, 3, 4, 5, 6, 7]) AS v
ON CONFLICT DO NOTHING;

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

CREATE UNIQUE INDEX IF NOT EXISTS idx_budget_groups_unique_name ON budget_groups(company_id, name);
