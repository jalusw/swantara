-- +goose Up
CREATE TABLE accrual_lines (
    id BIGSERIAL PRIMARY KEY,
    accrual_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    name TEXT,
    debit NUMERIC(18,4) NOT NULL DEFAULT 0,
    credit NUMERIC(18,4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_accrual_lines_accrual FOREIGN KEY (accrual_id) REFERENCES accruals(id),
    CONSTRAINT fk_accrual_lines_account FOREIGN KEY (account_id) REFERENCES accounts(id)
);

CREATE INDEX idx_accrual_lines_accrual ON accrual_lines(accrual_id);

-- +goose Down
DROP TABLE accrual_lines;