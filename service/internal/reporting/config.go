package reporting

import (
	"context"
	"encoding/json"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

const (
	configReportingJournalID      = "reporting.journal_id"
	configReportingFxGainLossAcct = "reporting.fx_gain_loss_account_id"
)

type ConfigLookup interface {
	List(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error)
}

type ConfigSource interface {
	JournalID(ctx context.Context, organizationID uint64) (uint64, error)
	FxGainLossAccountID(ctx context.Context, organizationID uint64) (uint64, error)
}

type configSource struct {
	configs ConfigLookup
}

func NewConfigSource(configs ConfigLookup) ConfigSource {
	return configSource{configs: configs}
}

func (s configSource) JournalID(ctx context.Context, organizationID uint64) (uint64, error) {
	return s.read(ctx, organizationID, configReportingJournalID)
}

func (s configSource) FxGainLossAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	return s.read(ctx, organizationID, configReportingFxGainLossAcct)
}

func (s configSource) read(ctx context.Context, organizationID uint64, key string) (uint64, error) {
	page, err := s.configs.List(ctx, &query.Query{Filters: []query.Filter{{Field: "key", Operator: query.Equal, Value: key}}})
	if err != nil {
		return 0, err
	}
	for _, config := range page.Items {
		if config.Key != key {
			continue
		}
		if config.OrganizationID != nil && *config.OrganizationID != organizationID {
			continue
		}
		if len(config.Value) == 0 {
			return 0, ErrConfigMissing
		}
		var value uint64
		if err := json.Unmarshal(config.Value, &value); err != nil {
			return 0, err
		}
		if value == 0 {
			return 0, ErrConfigMissing
		}
		return value, nil
	}
	return 0, ErrConfigMissing
}
