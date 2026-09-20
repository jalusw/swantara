package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type ReconcileRuleDAO interface {
	dao.CRUD[ReconcileRule]
	ListByOrganization(ctx context.Context, organizationID uint64) ([]*ReconcileRule, error)
}

type reconcileRuleDAO struct {
	dao.Base[ReconcileRule]
	db *gorm.DB
}

func NewReconcileRuleDAO(db *gorm.DB) ReconcileRuleDAO {
	return reconcileRuleDAO{Base: dao.NewBase[ReconcileRule](db), db: db}
}

func (d reconcileRuleDAO) ListByOrganization(ctx context.Context, organizationID uint64) ([]*ReconcileRule, error) {
	page, err := d.List(ctx, &query.Query{
		Filters: []query.Filter{
			{Field: "organization_id", Operator: query.Equal, Value: organizationID},
			{Field: "active", Operator: query.Equal, Value: true},
		},
		Sorts: []query.Sort{{Field: "sequence", Direction: query.Ascending}},
	})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type ReconcileRuleMatchDAO interface {
	dao.CRUD[ReconcileRuleMatch]
}

type reconcileRuleMatchDAO struct {
	dao.Base[ReconcileRuleMatch]
	db *gorm.DB
}

func NewReconcileRuleMatchDAO(db *gorm.DB) ReconcileRuleMatchDAO {
	return reconcileRuleMatchDAO{Base: dao.NewBase[ReconcileRuleMatch](db), db: db}
}
