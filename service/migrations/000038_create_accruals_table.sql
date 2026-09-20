-- +goose Up
CREATE TABLE accruals (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    period_id BIGINT NOT NULL,
    name TEXT,
    description TEXT,
    reversal_date DATE,
    state TEXT NOT NULL DEFAULT 'posted',
    movement_id BIGINT,
    reversal_movement_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_accruals_organization FOREIGN KEY (organization_id) REFERENCES organizations(id),
    CONSTRAINT fk_accruals_period FOREIGN KEY (period_id) REFERENCES tax_periods(id)
);

CREATE INDEX idx_accruals_org ON accruals(organization_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE accruals;