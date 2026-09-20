-- +goose Up
CREATE TABLE fx_revaluations (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    period_id BIGINT NOT NULL,
    name TEXT,
    date DATE NOT NULL,
    state TEXT NOT NULL DEFAULT 'posted',
    total_gain_loss NUMERIC(18,4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_fx_revaluations_organization FOREIGN KEY (organization_id) REFERENCES organizations(id),
    CONSTRAINT fk_fx_revaluations_period FOREIGN KEY (period_id) REFERENCES tax_periods(id)
);

CREATE INDEX idx_fx_revaluations_org ON fx_revaluations(organization_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE fx_revaluations;