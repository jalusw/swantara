-- +goose Up
SELECT 'up SQL query';
CREATE TABLE reminder_actions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  contact_id BIGINT,
  invoice_id BIGINT,
  level_id BIGINT,
  sent_at TIMESTAMP,
  channel TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_reminder_actions_contact_id
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_reminder_actions_invoice_id
    FOREIGN KEY (invoice_id)
    REFERENCES invoices(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_reminder_actions_level_id
    FOREIGN KEY (level_id)
    REFERENCES reminder_levels(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE reminder_actions;
