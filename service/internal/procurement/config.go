package procurement

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/systemconfig"
)

const (
	configApprovalThreshold = "purchase.approval_threshold"
	configApproverIDs       = "purchase.approver_ids"
)

type PurchaseConfigSource interface {
	ApprovalThreshold(ctx context.Context, organizationID uint64) (float64, error)
	ApproverIDs(ctx context.Context, organizationID uint64) ([]uint64, error)
}

type systemConfigSource struct {
	configs systemconfig.Lookup
}

func NewSystemConfigSource(configs systemconfig.Lookup) PurchaseConfigSource {
	return systemConfigSource{configs: configs}
}

func (s systemConfigSource) ApprovalThreshold(ctx context.Context, organizationID uint64) (float64, error) {
	var value float64
	if err := systemconfig.Read(ctx, s.configs, organizationID, configApprovalThreshold, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (s systemConfigSource) ApproverIDs(ctx context.Context, organizationID uint64) ([]uint64, error) {
	var value []uint64
	if err := systemconfig.Read(ctx, s.configs, organizationID, configApproverIDs, &value); err != nil {
		return nil, err
	}
	return value, nil
}
