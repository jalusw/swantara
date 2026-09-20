-- +goose Up
SELECT 'up SQL query';
CREATE TABLE pipeline_stages (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT,
  sequence INT,
  is_won BOOLEAN DEFAULT false,
  probability NUMERIC(5,2),
  organization_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_pipeline_stages_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_pipeline_stages_organization_id ON pipeline_stages(organization_id);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE pipeline_stages;
