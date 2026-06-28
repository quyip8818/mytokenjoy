CREATE TABLE IF NOT EXISTS usage_daily (
    date           DATE NOT NULL,
    department_id  TEXT NOT NULL,
    member_id      TEXT NOT NULL DEFAULT '',
    model          TEXT NOT NULL,
    cost_cny       DOUBLE PRECISION NOT NULL DEFAULT 0,
    call_count     INT NOT NULL DEFAULT 0,
    input_tokens   BIGINT NOT NULL DEFAULT 0,
    output_tokens  BIGINT NOT NULL DEFAULT 0,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (date, department_id, member_id, model)
);

CREATE INDEX IF NOT EXISTS idx_usage_daily_dept_date ON usage_daily (department_id, date);

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
