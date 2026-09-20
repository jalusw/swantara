-- +goose Up
CREATE TABLE shifts (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    organization_id BIGINT NOT NULL REFERENCES organizations(id),
    name VARCHAR(100) NOT NULL,
    start_time VARCHAR(5) NOT NULL,
    end_time VARCHAR(5) NOT NULL
);

CREATE INDEX idx_shifts_organization_id ON shifts (organization_id) WHERE deleted_at IS NULL;

CREATE TABLE shift_assignments (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    employee_id BIGINT NOT NULL REFERENCES employees(id),
    shift_id BIGINT NOT NULL REFERENCES shifts(id),
    date DATE NOT NULL
);

CREATE UNIQUE INDEX idx_shift_assignments_employee_date ON shift_assignments (employee_id, date) WHERE deleted_at IS NULL;
CREATE INDEX idx_shift_assignments_shift_id ON shift_assignments (shift_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS shift_assignments;
DROP TABLE IF EXISTS shifts;
