-- +goose Up
SELECT 'up SQL query';
CREATE TABLE approval_steps (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  request_id BIGINT,
  approver_id BIGINT,
  sequence INT,
  decision TEXT CHECK (decision IN ('pending','approved','refused')),
  decided_at TIMESTAMP,
  comment TEXT,
  organization_id BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_approval_steps_request_id
    FOREIGN KEY (request_id)
    REFERENCES approval_requests(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_approval_steps_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_approval_steps_organization_id ON approval_steps(organization_id);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE approval_steps;
