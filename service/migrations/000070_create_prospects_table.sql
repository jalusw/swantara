-- +goose Up
SELECT 'up SQL query';
CREATE TABLE prospects (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT NOT NULL,
  type TEXT CHECK (type IN ('lead','opportunity')),
  contact_id BIGINT,
  contact_name TEXT,
  email TEXT,
  phone TEXT,
  job_position TEXT,
  stage_id BIGINT,
  expected_revenue NUMERIC(18,4),
  probability NUMERIC(5,2),
  priority SMALLINT DEFAULT 0,
  salesperson_id BIGINT,
  sales_group_id BIGINT,
  source TEXT,
  medium TEXT,
  campaign TEXT,
  lost_reason TEXT,
  expected_close DATE,
  closed_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_prospects_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_prospects_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_prospects_stage_id
    FOREIGN KEY (stage_id)
    REFERENCES pipeline_stages(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_prospects_salesperson_id
    FOREIGN KEY (salesperson_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_prospects_sales_group_id
    FOREIGN KEY (sales_group_id)
    REFERENCES sales_groups(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE prospects;
