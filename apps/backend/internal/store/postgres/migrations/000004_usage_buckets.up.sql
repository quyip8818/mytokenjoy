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

INSERT INTO usage_buckets (
    bucket_start, department_id, member_id, model,
    cost_cny, call_count, input_tokens, output_tokens
)
SELECT
    (date::timestamp AT TIME ZONE 'Asia/Shanghai') AS bucket_start,
    department_id, member_id, model,
    cost_cny, call_count, input_tokens, output_tokens
FROM usage_daily
ON CONFLICT (bucket_start, department_id, member_id, model) DO NOTHING;

DROP TABLE IF EXISTS usage_daily;
