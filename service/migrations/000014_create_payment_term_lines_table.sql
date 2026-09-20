-- +goose Up
SELECT 'up SQL query';
CREATE TABLE payment_term_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  payment_term_id BIGINT NOT NULL,
  sequence INT DEFAULT 10,
  value_type TEXT CHECK (value_type IN ('percent','fixed','balance')),
  value NUMERIC(12,4),
  days_after INT DEFAULT 0,
  day_of_month INT,
  discount_pct NUMERIC(8,4),
  discount_days INT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_payment_term_lines_payment_term_id
    FOREIGN KEY (payment_term_id)
    REFERENCES payment_terms(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE payment_term_lines;
