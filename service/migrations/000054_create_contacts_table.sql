-- +goose Up
SELECT 'up SQL query';
CREATE TABLE contacts (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  name TEXT NOT NULL,
  display_name TEXT,
  is_organization BOOLEAN DEFAULT false,
  parent_id BIGINT,
  email TEXT,
  phone TEXT,
  mobile TEXT,
  website TEXT,
  tax_id TEXT,
  industry TEXT,
  currency_code CHAR(3),
  lang TEXT DEFAULT 'en',
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_contacts_organization_id
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_contacts_parent_id
    FOREIGN KEY (parent_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_contacts_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

ALTER TABLE users
  ADD CONSTRAINT fk_users_contact
    FOREIGN KEY (contact_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION;

-- +goose Down
SELECT 'down SQL query';
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_contact;
DROP TABLE contacts;
