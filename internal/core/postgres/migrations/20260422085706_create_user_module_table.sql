-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    email      TEXT NOT NULL UNIQUE,
    password   TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_id ON users(id);

CREATE TABLE IF NOT EXISTS modules (
    id          SERIAL PRIMARY KEY,
    owner_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT DEFAULT '',
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS webhooks (
    id          SERIAL PRIMARY KEY,
    module_id   INT NOT NULL REFERENCES modules(id) ON DELETE RESTRICT,
    description TEXT DEFAULT '',
    interval_s  BIGINT NOT NULL,
    timeout_s   BIGINT NOT NULL,
    url         TEXT NOT NULL,
    method      TEXT NOT NULL,
    headers     JSONB DEFAULT '{}',
    body        BYTEA,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_modules_owner_id ON modules(owner_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_module_id ON webhooks(module_id);

-- +goose Down
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS users;