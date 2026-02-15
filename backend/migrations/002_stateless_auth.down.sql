-- Recreate sessions table
CREATE TABLE sessions (
    token      TEXT PRIMARY KEY,
    user_type  TEXT NOT NULL CHECK(user_type IN ('parent', 'child')),
    user_id    INTEGER NOT NULL,
    family_id  INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- Drop refresh_tokens table
DROP TABLE IF EXISTS refresh_tokens;
