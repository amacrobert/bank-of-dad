-- Create refresh_tokens table
CREATE TABLE refresh_tokens (
    id         SERIAL PRIMARY KEY,
    token_hash TEXT NOT NULL UNIQUE,
    user_type  TEXT NOT NULL CHECK(user_type IN ('parent', 'child')),
    user_id    INTEGER NOT NULL,
    family_id  INTEGER NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_type, user_id);

-- Drop sessions table
DROP TABLE IF EXISTS sessions;
