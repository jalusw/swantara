-- +goose Up
SELECT 'up SQL query';
CREATE TABLE supplier_quote_request_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  quote_request_id BIGINT,
  item_id BIGINT,
  description TEXT,
  qty NUMERIC(18,4),
  unit_id BIGINT,
  needed_by DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_supplier_quote_request_lines_quote_request_id
    FOREIGN KEY (quote_request_id)
    REFERENCES supplier_quote_requests(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_supplier_quote_request_lines_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_supplier_quote_request_lines_unit_id
    FOREIGN KEY (unit_id)
    REFERENCES units(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TABLE supplier_quotes (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  quote_request_id BIGINT,
  supplier_id BIGINT NOT NULL,
  currency_code CHAR(3),
  state TEXT CHECK (state IN ('draft','submitted','accepted','rejected')),
  quote_date DATE,
  valid_until DATE,
  notes TEXT,
  amount_untaxed NUMERIC(18,4) DEFAULT 0,
  amount_tax NUMERIC(18,4) DEFAULT 0,
  amount_total NUMERIC(18,4) DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_supplier_quotes_quote_request_id
    FOREIGN KEY (quote_request_id)
    REFERENCES supplier_quote_requests(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_supplier_quotes_supplier_id
    FOREIGN KEY (supplier_id)
    REFERENCES contacts(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_supplier_quotes_currency_code
    FOREIGN KEY (currency_code)
    REFERENCES currencies(code)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

CREATE TABLE supplier_quote_lines (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  supplier_quote_id BIGINT,
  quote_request_line_id BIGINT,
  item_id BIGINT,
  description TEXT,
  qty NUMERIC(18,4),
  unit_price NUMERIC(18,4),
  discount_pct NUMERIC(8,4) DEFAULT 0,
  price_subtotal NUMERIC(18,4) DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_supplier_quote_lines_supplier_quote_id
    FOREIGN KEY (supplier_quote_id)
    REFERENCES supplier_quotes(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_supplier_quote_lines_quote_request_line_id
    FOREIGN KEY (quote_request_line_id)
    REFERENCES supplier_quote_request_lines(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_supplier_quote_lines_item_id
    FOREIGN KEY (item_id)
    REFERENCES item_variants(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE supplier_quote_lines;
DROP TABLE supplier_quotes;
DROP TABLE supplier_quote_request_lines;