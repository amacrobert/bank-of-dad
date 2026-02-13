-- families
CREATE TABLE families (
    id         SERIAL PRIMARY KEY,
    slug       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- parents
CREATE TABLE parents (
    id           SERIAL PRIMARY KEY,
    google_id    TEXT NOT NULL UNIQUE,
    email        TEXT NOT NULL,
    display_name TEXT NOT NULL,
    family_id    INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- children
CREATE TABLE children (
    id                    SERIAL PRIMARY KEY,
    family_id             INTEGER NOT NULL REFERENCES families(id),
    first_name            TEXT NOT NULL,
    password_hash         TEXT NOT NULL,
    is_locked             BOOLEAN NOT NULL DEFAULT FALSE,
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    balance_cents         BIGINT NOT NULL DEFAULT 0,
    interest_rate_bps     INTEGER NOT NULL DEFAULT 0,
    last_interest_at      TIMESTAMPTZ,
    avatar                TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(family_id, first_name)
);

-- sessions
CREATE TABLE sessions (
    token      TEXT PRIMARY KEY,
    user_type  TEXT NOT NULL CHECK(user_type IN ('parent', 'child')),
    user_id    INTEGER NOT NULL,
    family_id  INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- auth_events
CREATE TABLE auth_events (
    id          SERIAL PRIMARY KEY,
    event_type  TEXT NOT NULL,
    user_type   TEXT NOT NULL,
    user_id     INTEGER,
    family_id   INTEGER,
    ip_address  TEXT NOT NULL,
    details     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_events_created ON auth_events(created_at);
CREATE INDEX idx_auth_events_family ON auth_events(family_id);

-- transactions
CREATE TABLE transactions (
    id               SERIAL PRIMARY KEY,
    child_id         INTEGER NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    parent_id        INTEGER NOT NULL REFERENCES parents(id) ON DELETE RESTRICT,
    amount_cents     BIGINT NOT NULL,
    transaction_type TEXT NOT NULL CHECK(transaction_type IN ('deposit', 'withdrawal', 'allowance', 'interest')),
    note             TEXT,
    schedule_id      INTEGER,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_child_created ON transactions(child_id, created_at DESC);

-- allowance_schedules
CREATE TABLE allowance_schedules (
    id           SERIAL PRIMARY KEY,
    child_id     INTEGER NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    parent_id    INTEGER NOT NULL REFERENCES parents(id) ON DELETE RESTRICT,
    amount_cents BIGINT NOT NULL,
    frequency    TEXT NOT NULL CHECK(frequency IN ('weekly', 'biweekly', 'monthly')),
    day_of_week  INTEGER CHECK(day_of_week >= 0 AND day_of_week <= 6),
    day_of_month INTEGER CHECK(day_of_month >= 1 AND day_of_month <= 31),
    note         TEXT,
    status       TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'paused')),
    next_run_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK(
        (frequency = 'weekly' AND day_of_week IS NOT NULL) OR
        (frequency = 'biweekly' AND day_of_week IS NOT NULL) OR
        (frequency = 'monthly' AND day_of_month IS NOT NULL)
    )
);

CREATE UNIQUE INDEX idx_allowance_schedules_unique_child ON allowance_schedules(child_id);
CREATE INDEX idx_schedules_due ON allowance_schedules(status, next_run_at);
CREATE INDEX idx_schedules_child ON allowance_schedules(child_id);

-- Add schedule_id FK now that allowance_schedules table exists
ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_schedule
    FOREIGN KEY (schedule_id) REFERENCES allowance_schedules(id) ON DELETE SET NULL;

-- interest_schedules
CREATE TABLE interest_schedules (
    id           SERIAL PRIMARY KEY,
    child_id     INTEGER NOT NULL UNIQUE REFERENCES children(id) ON DELETE CASCADE,
    parent_id    INTEGER NOT NULL REFERENCES parents(id) ON DELETE RESTRICT,
    frequency    TEXT NOT NULL CHECK(frequency IN ('weekly', 'biweekly', 'monthly')),
    day_of_week  INTEGER CHECK(day_of_week >= 0 AND day_of_week <= 6),
    day_of_month INTEGER CHECK(day_of_month >= 1 AND day_of_month <= 31),
    status       TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'paused')),
    next_run_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK(
        (frequency = 'weekly' AND day_of_week IS NOT NULL) OR
        (frequency = 'biweekly' AND day_of_week IS NOT NULL) OR
        (frequency = 'monthly' AND day_of_month IS NOT NULL)
    )
);

CREATE INDEX idx_interest_schedules_due ON interest_schedules(status, next_run_at);
