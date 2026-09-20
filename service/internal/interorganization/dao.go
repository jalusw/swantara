package interorganization

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type DropshipLinkDAO interface {
	dao.CRUD[DropshipLink]
	ListByPurchaseOrder(ctx context.Context, orderID uint64) ([]*DropshipLink, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, link *DropshipLink) (*DropshipLink, error)
}

type dropshipLinkDAO struct {
	dao.Base[DropshipLink]
	db *gorm.DB
}

func NewDropshipLinkDAO(db *gorm.DB) DropshipLinkDAO {
	return dropshipLinkDAO{Base: dao.NewBase[DropshipLink](db), db: db}
}

func (d dropshipLinkDAO) ListByPurchaseOrder(ctx context.Context, orderID uint64) ([]*DropshipLink, error) {
	var links []DropshipLink
	if err := d.db.WithContext(ctx).
		Joins("JOIN purchase_order_lines ON purchase_order_lines.id = dropship_links.purchase_order_line_id").
		Where("purchase_order_lines.order_id = ? AND dropship_links.deleted_at IS NULL", orderID).
		Find(&links).Error; err != nil {
		return nil, err
	}
	items := make([]*DropshipLink, len(links))
	for i := range links {
		items[i] = &links[i]
	}
	return items, nil
}

func (d dropshipLinkDAO) UpdateTx(ctx context.Context, tx *gorm.DB, link *DropshipLink) (*DropshipLink, error) {
	if err := tx.WithContext(ctx).Save(link).Error; err != nil {
		return nil, err
	}
	return link, nil
}

type InterorganizationRuleDAO interface {
	dao.CRUD[InterorganizationRule]
}

type interorganizationRuleDAO struct {
	dao.Base[InterorganizationRule]
}

func NewInterorganizationRuleDAO(db *gorm.DB) InterorganizationRuleDAO {
	return interorganizationRuleDAO{Base: dao.NewBase[InterorganizationRule](db)}
}

type InterorganizationTransactionDAO interface {
	dao.CRUD[InterorganizationTransaction]
}

type interorganizationTransactionDAO struct {
	dao.Base[InterorganizationTransaction]
}

func NewInterorganizationTransactionDAO(db *gorm.DB) InterorganizationTransactionDAO {
	return interorganizationTransactionDAO{Base: dao.NewBase[InterorganizationTransaction](db)}
}

type ConsolidationRunDAO interface {
	dao.CRUD[ConsolidationRun]
	UpdateTx(ctx context.Context, tx *gorm.DB, run *ConsolidationRun) (*ConsolidationRun, error)
}

type consolidationRunDAO struct {
	dao.Base[ConsolidationRun]
	db *gorm.DB
}

func NewConsolidationRunDAO(db *gorm.DB) ConsolidationRunDAO {
	return consolidationRunDAO{Base: dao.NewBase[ConsolidationRun](db), db: db}
}

func (d consolidationRunDAO) UpdateTx(ctx context.Context, tx *gorm.DB, run *ConsolidationRun) (*ConsolidationRun, error) {
	if err := tx.WithContext(ctx).Save(run).Error; err != nil {
		return nil, err
	}
	return run, nil
}

type ConsolidationEliminationDAO interface {
	dao.CRUD[ConsolidationElimination]
	ListByRun(ctx context.Context, runID uint64) ([]*ConsolidationElimination, error)
	CreateTx(ctx context.Context, tx *gorm.DB, elimination *ConsolidationElimination) (*ConsolidationElimination, error)
}

type consolidationEliminationDAO struct {
	dao.Base[ConsolidationElimination]
	db *gorm.DB
}

func NewConsolidationEliminationDAO(db *gorm.DB) ConsolidationEliminationDAO {
	return consolidationEliminationDAO{Base: dao.NewBase[ConsolidationElimination](db), db: db}
}

func (d consolidationEliminationDAO) ListByRun(ctx context.Context, runID uint64) ([]*ConsolidationElimination, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "consolidation_run_id", Operator: query.Equal, Value: runID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d consolidationEliminationDAO) CreateTx(ctx context.Context, tx *gorm.DB, elimination *ConsolidationElimination) (*ConsolidationElimination, error) {
	if err := tx.WithContext(ctx).Create(elimination).Error; err != nil {
		return nil, err
	}
	return elimination, nil
}
