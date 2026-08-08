CREATE TABLE IF NOT EXISTS refresh_tokens (
    id            UUID PRIMARY KEY,
    token         VARCHAR(50),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at    TIMESTAMP NOT NULL,
    revoked       BOOLEAN NOT NULL DEFAULT FALSE,
    issued_at     TIMESTAMP NOT NULL DEFAULT NOW()
);