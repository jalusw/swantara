-- +goose Up
SELECT 'up SQL query';
CREATE TABLE tax_rule_tax_maps (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  tax_rule_id BIGINT,
  src_tax_id BIGINT,
  dest_tax_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_tax_rule_tax_maps_tax_rule_id
    FOREIGN KEY (tax_rule_id)
    REFERENCES tax_rules(id)
    ON DELETE CASCADE
    ON UPDATE NO ACTION,
  CONSTRAINT fk_tax_rule_tax_maps_src_tax_id
    FOREIGN KEY (src_tax_id)
    REFERENCES taxes(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION,
  CONSTRAINT fk_tax_rule_tax_maps_dest_tax_id
    FOREIGN KEY (dest_tax_id)
    REFERENCES taxes(id)
    ON DELETE RESTRICT
    ON UPDATE NO ACTION
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE tax_rule_tax_maps;
