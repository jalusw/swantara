package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
)

type ReconcileRuleDAOMock struct {
	dao.CRUDMock[ReconcileRule]
	ListByOrganizationFunc func(ctx context.Context, organizationID uint64) ([]*ReconcileRule, error)
}

func (m ReconcileRuleDAOMock) ListByOrganization(ctx context.Context, organizationID uint64) ([]*ReconcileRule, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, organizationID)
	}
	return nil, nil
}

type ReconcileRuleMatchDAOMock struct {
	dao.CRUDMock[ReconcileRuleMatch]
}
