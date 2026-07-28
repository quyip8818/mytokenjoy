-- schema.sql
-- 幂等：所有 CREATE 带 IF NOT EXISTS
-- 注意：UUID v7 由 Go 应用层生成（uuid.NewV7），DB DEFAULT 用 gen_random_uuid() 作为安全网

CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ====== users ======
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(64) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    real_name     VARCHAR(64) NOT NULL DEFAULT '',
    email         VARCHAR(128),
    role          VARCHAR(32) NOT NULL DEFAULT 'viewer'
                  CHECK (role IN ('admin', 'buyer', 'viewer')),
    status        SMALLINT NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER set_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ====== sessions ======
CREATE TABLE IF NOT EXISTS sessions (
    token      VARCHAR(64) PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- ====== suppliers ======
CREATE TABLE IF NOT EXISTS suppliers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(128) NOT NULL,
    code        VARCHAR(64) UNIQUE NOT NULL,
    category    VARCHAR(64),
    website     VARCHAR(256),
    status      VARCHAR(32) NOT NULL DEFAULT 'potential'
                CHECK (status IN ('potential', 'active', 'frozen', 'blacklisted')),
    description TEXT,
    created_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_suppliers_status ON suppliers(status);
CREATE TRIGGER set_updated_at BEFORE UPDATE ON suppliers
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ====== supplier_contacts ======
CREATE TABLE IF NOT EXISTS supplier_contacts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    name        VARCHAR(64) NOT NULL,
    position    VARCHAR(64),
    phone       VARCHAR(32),
    email       VARCHAR(128),
    is_primary  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_supplier_contacts_supplier ON supplier_contacts(supplier_id);

-- ====== models (AI 模型) ======
CREATE TABLE IF NOT EXISTS models (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id    UUID NOT NULL REFERENCES suppliers(id),
    model_name     VARCHAR(128) NOT NULL,
    model_id       VARCHAR(128),
    model_type     VARCHAR(32),
    context_length INT,
    input_price    NUMERIC(12,6),
    output_price   NUMERIC(12,6),
    discount       NUMERIC(5,2),
    status         VARCHAR(32) NOT NULL DEFAULT 'available'
                   CHECK (status IN ('available', 'deprecated')),
    description    TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_models_supplier ON models(supplier_id);
CREATE INDEX IF NOT EXISTS idx_models_status ON models(status);
CREATE TRIGGER set_updated_at BEFORE UPDATE ON models
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ====== contracts ======
CREATE TABLE IF NOT EXISTS contracts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID NOT NULL REFERENCES suppliers(id),
    contract_no VARCHAR(64) UNIQUE NOT NULL,
    title       VARCHAR(256) NOT NULL,
    amount      NUMERIC(14,2),
    sign_date   DATE,
    start_date  DATE,
    end_date    DATE,
    status      VARCHAR(32) NOT NULL DEFAULT 'draft'
                CHECK (status IN ('draft', 'active', 'expired', 'terminated')),
    remarks     TEXT,
    created_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_contracts_supplier ON contracts(supplier_id);
CREATE INDEX IF NOT EXISTS idx_contracts_status ON contracts(status);
CREATE INDEX IF NOT EXISTS idx_contracts_end_date ON contracts(end_date);
CREATE TRIGGER set_updated_at BEFORE UPDATE ON contracts
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ====== contract_attachments ======
CREATE TABLE IF NOT EXISTS contract_attachments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    file_name   VARCHAR(256) NOT NULL,
    file_path   VARCHAR(512) NOT NULL,
    file_size   BIGINT NOT NULL DEFAULT 0,
    uploaded_by UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_contract_attachments_contract ON contract_attachments(contract_id);

-- ====== purchase_orders ======
CREATE TABLE IF NOT EXISTS purchase_orders (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_no     VARCHAR(64) UNIQUE NOT NULL,
    supplier_id  UUID NOT NULL REFERENCES suppliers(id),
    contract_id  UUID REFERENCES contracts(id),
    total_amount NUMERIC(14,2),
    order_date   DATE,
    status       VARCHAR(32) NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'approved', 'delivered', 'completed', 'cancelled')),
    description  TEXT,
    created_by   UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_supplier ON purchase_orders(supplier_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON purchase_orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_contract ON purchase_orders(contract_id);
CREATE TRIGGER set_updated_at BEFORE UPDATE ON purchase_orders
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ====== evaluations ======
CREATE TABLE IF NOT EXISTS evaluations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id  UUID NOT NULL REFERENCES suppliers(id),
    evaluator_id UUID NOT NULL REFERENCES users(id),
    period       VARCHAR(32) NOT NULL,
    quality      INT NOT NULL CHECK (quality BETWEEN 0 AND 100),
    performance  INT NOT NULL CHECK (performance BETWEEN 0 AND 100),
    price        INT NOT NULL CHECK (price BETWEEN 0 AND 100),
    service      INT NOT NULL CHECK (service BETWEEN 0 AND 100),
    compliance   INT NOT NULL CHECK (compliance BETWEEN 0 AND 100),
    total_score  NUMERIC(5,2) NOT NULL,
    grade        VARCHAR(2) NOT NULL CHECK (grade IN ('A', 'B', 'C', 'D')),
    comment      TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (supplier_id, evaluator_id, period)
);
CREATE INDEX IF NOT EXISTS idx_evaluations_supplier ON evaluations(supplier_id);
CREATE INDEX IF NOT EXISTS idx_evaluations_period ON evaluations(period);

-- ====== evaluation_weights ======
CREATE TABLE IF NOT EXISTS evaluation_weights (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dimension VARCHAR(32) UNIQUE NOT NULL,
    weight    INT NOT NULL CHECK (weight >= 0 AND weight <= 100)
);

-- ====== oauth_clients (for cross-system API access) ======
CREATE TABLE IF NOT EXISTS oauth_clients (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id           TEXT UNIQUE NOT NULL,
    client_secret_hash  TEXT NOT NULL,
    scope               TEXT NOT NULL DEFAULT 'sync:read',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ====== sync_versions (partition version counter for incremental sync) ======
CREATE TABLE IF NOT EXISTS sync_versions (
    partition  TEXT PRIMARY KEY,
    version    INT NOT NULL DEFAULT 0
);

-- Initialize partitions
INSERT INTO sync_versions (partition, version) VALUES
    ('channels', 0),
    ('models', 0),
    ('currencies', 0)
ON CONFLICT (partition) DO NOTHING;

-- Auto-increment sync version on models changes
CREATE OR REPLACE FUNCTION bump_sync_version_models()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE sync_versions SET version = version + 1 WHERE partition = 'models';
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_bump_sync_models ON models;
CREATE TRIGGER trg_bump_sync_models
    AFTER INSERT OR UPDATE OR DELETE ON models
    FOR EACH STATEMENT EXECUTE FUNCTION bump_sync_version_models();
