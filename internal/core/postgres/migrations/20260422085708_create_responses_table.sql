-- +goose Up
CREATE TABLE IF NOT EXISTS responses (
    id          BIGSERIAL PRIMARY KEY,
    webhook_id  BIGINT NOT NULL,
    type        VARCHAR(16) NOT NULL,
    status      VARCHAR(16) NOT NULL,
    status_code INT NOT NULL DEFAULT 0,
    body        BYTEA,
    headers     JSONB,
    started_at  TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    attempt     INT NOT NULL DEFAULT 1,
    duration    BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_responses_webhook_id ON responses(webhook_id);

-- +goose Down
DROP TABLE IF EXISTS responses;