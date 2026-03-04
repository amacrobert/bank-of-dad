-- Savings goals: children can set target amounts to save toward
CREATE TABLE savings_goals (
    id             SERIAL PRIMARY KEY,
    child_id       INTEGER NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    target_cents   BIGINT NOT NULL CHECK(target_cents > 0),
    saved_cents    BIGINT NOT NULL DEFAULT 0 CHECK(saved_cents >= 0),
    emoji          TEXT,
    target_date    DATE,
    status         TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'completed')),
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_savings_goals_child_status ON savings_goals(child_id, status);
CREATE INDEX idx_savings_goals_child_created ON savings_goals(child_id, created_at DESC);

-- Goal allocations: audit trail of fund transfers to/from goals
CREATE TABLE goal_allocations (
    id             SERIAL PRIMARY KEY,
    goal_id        INTEGER NOT NULL REFERENCES savings_goals(id) ON DELETE CASCADE,
    child_id       INTEGER NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    amount_cents   BIGINT NOT NULL CHECK(amount_cents != 0),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_goal_allocations_goal ON goal_allocations(goal_id, created_at DESC);
