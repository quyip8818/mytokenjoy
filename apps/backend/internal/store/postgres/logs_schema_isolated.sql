CREATE TABLE IF NOT EXISTS logs (
    id                  SERIAL PRIMARY KEY,
    user_id             INT,
    created_at          BIGINT NOT NULL DEFAULT 0,
    type                INT NOT NULL DEFAULT 0,
    content             TEXT DEFAULT '',
    token_id            INT NOT NULL DEFAULT 0,
    model_name          TEXT DEFAULT '',
    quota               INT NOT NULL DEFAULT 0,
    prompt_tokens       INT NOT NULL DEFAULT 0,
    completion_tokens   INT NOT NULL DEFAULT 0,
    use_time            INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS reconcile_cursors (
    stream       TEXT PRIMARY KEY,
    last_log_id  BIGINT NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO reconcile_cursors (stream, last_log_id)
VALUES ('newapi_consume', 0)
ON CONFLICT (stream) DO NOTHING;

CREATE TABLE IF NOT EXISTS ingest_jobs (
    id           TEXT PRIMARY KEY,
    log_id       BIGINT NOT NULL UNIQUE,
    source       TEXT NOT NULL,
    error        TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    attempts     INT NOT NULL DEFAULT 0,
    next_retry   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ingest_jobs_pending
    ON ingest_jobs (next_retry)
    WHERE status = 'pending' AND attempts < 20;

CREATE INDEX IF NOT EXISTS idx_logs_consume_cursor
    ON logs (id)
    WHERE type = 2 AND token_id > 0;
