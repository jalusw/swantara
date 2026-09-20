-- +goose Up
CREATE TABLE job_runs (
  id BIGSERIAL PRIMARY KEY,
  job_type TEXT NOT NULL,
  status TEXT CHECK (status IN ('running','completed','failed')),
  processed INT DEFAULT 0,
  cursor BIGINT,
  error_message TEXT,
  started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_job_runs_type_status ON job_runs(job_type, status);

-- +goose Down
DROP TABLE job_runs;