-- +goose Up
SELECT 'up SQL query';
CREATE TABLE attachments (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  owner_type TEXT,
  owner_id BIGINT,
  filename TEXT,
  mime_type TEXT,
  byte_size BIGINT,
  storage_url TEXT,
  checksum TEXT,
  uploaded_by BIGINT,
  uploaded_at TIMESTAMP,
  organization_id BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_attachments_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_attachments_organization_id ON attachments(organization_id);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE attachments;
