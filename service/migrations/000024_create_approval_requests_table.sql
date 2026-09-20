-- +goose Up
SELECT 'up SQL query';
CREATE TABLE approval_requests (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  owner_type TEXT,
  owner_id BIGINT,
  requested_by BIGINT,
  state TEXT CHECK (state IN ('pending','approved','refused')),
  organization_id BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_approval_requests_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_approval_requests_organization_id ON approval_requests(organization_id);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE approval_requests;
