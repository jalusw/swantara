-- +goose Up
CREATE TABLE kpi_account_summaries (
  organization_id BIGINT NOT NULL,
  tax_period_id BIGINT NOT NULL,
  account_id BIGINT NOT NULL,
  opening_debit NUMERIC(18,4) NOT NULL DEFAULT 0,
  opening_credit NUMERIC(18,4) NOT NULL DEFAULT 0,
  period_debit NUMERIC(18,4) NOT NULL DEFAULT 0,
  period_credit NUMERIC(18,4) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  PRIMARY KEY (organization_id, tax_period_id, account_id)
);

CREATE INDEX idx_kpi_summary_period_account ON kpi_account_summaries(tax_period_id, account_id);

-- +goose Down
DROP TABLE kpi_account_summaries;