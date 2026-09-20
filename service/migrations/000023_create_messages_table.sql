-- +goose Up
SELECT 'up SQL query';
CREATE TABLE messages (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  owner_type TEXT,
  owner_id BIGINT,
  author_id BIGINT,
  body TEXT,
  message_type TEXT,
  organization_id BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_messages_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE INDEX idx_messages_organization_id ON messages(organization_id);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE messages;
