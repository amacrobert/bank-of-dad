-- Chore templates created by parents
CREATE TABLE chores (
    id BIGSERIAL PRIMARY KEY,
    family_id BIGINT NOT NULL REFERENCES families(id),
    created_by_parent_id BIGINT NOT NULL REFERENCES parents(id),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    reward_cents INTEGER NOT NULL DEFAULT 0,
    recurrence VARCHAR(10) NOT NULL DEFAULT 'one_time',
    day_of_week SMALLINT,
    day_of_month SMALLINT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_reward_cents_non_negative CHECK (reward_cents >= 0),
    CONSTRAINT chk_recurrence_valid CHECK (recurrence IN ('one_time', 'daily', 'weekly', 'monthly')),
    CONSTRAINT chk_day_of_week_range CHECK (day_of_week IS NULL OR (day_of_week >= 0 AND day_of_week <= 6)),
    CONSTRAINT chk_day_of_month_range CHECK (day_of_month IS NULL OR (day_of_month >= 1 AND day_of_month <= 31))
);

CREATE INDEX idx_chores_family_id ON chores(family_id);

-- Many-to-many: which children are assigned to each chore
CREATE TABLE chore_assignments (
    id BIGSERIAL PRIMARY KEY,
    chore_id BIGINT NOT NULL REFERENCES chores(id) ON DELETE CASCADE,
    child_id BIGINT NOT NULL REFERENCES children(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_chore_assignments_chore_child UNIQUE (chore_id, child_id)
);

CREATE INDEX idx_chore_assignments_chore_id ON chore_assignments(chore_id);
CREATE INDEX idx_chore_assignments_child_id ON chore_assignments(child_id);

-- Individual occurrences of a chore for a specific child
CREATE TABLE chore_instances (
    id BIGSERIAL PRIMARY KEY,
    chore_id BIGINT NOT NULL REFERENCES chores(id),
    child_id BIGINT NOT NULL REFERENCES children(id),
    reward_cents INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    period_start DATE,
    period_end DATE,
    completed_at TIMESTAMPTZ,
    reviewed_at TIMESTAMPTZ,
    reviewed_by_parent_id BIGINT REFERENCES parents(id),
    rejection_reason VARCHAR(500),
    transaction_id BIGINT REFERENCES transactions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_instance_status_valid CHECK (status IN ('available', 'pending_approval', 'approved', 'expired'))
);

CREATE INDEX idx_chore_instances_chore_id ON chore_instances(chore_id);
CREATE INDEX idx_chore_instances_child_id ON chore_instances(child_id);
CREATE INDEX idx_chore_instances_status ON chore_instances(status);
CREATE UNIQUE INDEX uq_chore_instances_chore_child_period ON chore_instances(chore_id, child_id, period_start) WHERE period_start IS NOT NULL;
