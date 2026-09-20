-- +goose Up
CREATE TABLE fx_revaluation_lines (
    id BIGSERIAL PRIMARY KEY,
    revaluation_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    currency_code CHAR(3) NOT NULL,
    foreign_balance NUMERIC(18,4) NOT NULL DEFAULT 0,
    base_balance NUMERIC(18,4) NOT NULL DEFAULT 0,
    closing_rate NUMERIC(18,8) NOT NULL DEFAULT 0,
    gain_loss NUMERIC(18,4) NOT NULL DEFAULT 0,
    movement_id BIGINT,
    reversed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_fx_revaluation_lines_revaluation FOREIGN KEY (revaluation_id) REFERENCES fx_revaluations(id),
    CONSTRAINT fk_fx_revaluation_lines_account FOREIGN KEY (account_id) REFERENCES accounts(id)
);

CREATE INDEX idx_fx_revaluation_lines_open ON fx_revaluation_lines(account_id) WHERE reversed = FALSE;

-- +goose Down
DROP TABLE fx_revaluation_lines;