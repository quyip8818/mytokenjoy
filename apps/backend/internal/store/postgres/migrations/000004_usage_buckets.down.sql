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

INSERT INTO usage_daily (
    date, department_id, member_id, model,
    cost_cny, call_count, input_tokens, output_tokens
)
SELECT
    (bucket_start AT TIME ZONE 'Asia/Shanghai')::date,
    department_id, member_id, model,
    cost_cny::double precision, call_count, input_tokens, output_tokens
FROM usage_buckets
ON CONFLICT (date, department_id, member_id, model) DO NOTHING;

DROP TABLE IF EXISTS usage_buckets;
