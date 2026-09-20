-- +goose Up
CREATE TABLE webhook_subscriptions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  organization_id BIGINT,
  url TEXT NOT NULL,
  secret TEXT NOT NULL,
  enabled BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_webhook_subscriptions_org ON webhook_subscriptions(organization_id)
  WHERE enabled = TRUE;

CREATE TABLE webhook_deliveries (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  event_id BIGINT NOT NULL REFERENCES integration_events(id) ON DELETE CASCADE,
  subscription_id BIGINT REFERENCES webhook_subscriptions(id) ON DELETE SET NULL,
  organization_id BIGINT,
  status TEXT CHECK (status IN ('delivered','failed')),
  attempt INT DEFAULT 0,
  response_code INT,
  error TEXT,
  delivered_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_webhook_deliveries_event ON webhook_deliveries(event_id, status);

-- +goose Down
DROP TABLE webhook_deliveries;
DROP TABLE webhook_subscriptions;