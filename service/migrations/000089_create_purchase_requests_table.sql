-- +goose Up
SELECT 'up SQL query';
CREATE TABLE purchase_requests (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT,
  requester_id BIGINT,
  department_id BIGINT,
  state TEXT CHECK (state IN ('draft','confirmed','approved','done','cancelled')),
  needed_by DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_purchase_requests_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_purchase_requests_requester_id
    FOREIGN KEY (requester_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE purchase_requests;
