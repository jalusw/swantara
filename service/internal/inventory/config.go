package inventory

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/systemconfig"
)

const configInboundCostJournalID = "inventory.inbound_cost_journal_id"

type InboundCostConfigSource interface {
	JournalID(ctx context.Context, organizationID uint64) (uint64, error)
}

type inboundCostConfigSource struct {
	configs systemconfig.Lookup
}

func NewInboundCostConfigSource(configs systemconfig.Lookup) InboundCostConfigSource {
	return inboundCostConfigSource{configs: configs}
}

func (s inboundCostConfigSource) JournalID(ctx context.Context, organizationID uint64) (uint64, error) {
	var value uint64
	if err := systemconfig.Read(ctx, s.configs, organizationID, configInboundCostJournalID, &value); err != nil {
		return 0, err
	}
	return value, nil
}
