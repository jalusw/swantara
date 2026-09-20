package interorganization

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type DropshipLinkDAOMock struct {
	dao.CRUDMock[DropshipLink]
	ListByPurchaseOrderFunc func(ctx context.Context, orderID uint64) ([]*DropshipLink, error)
	UpdateTxFunc            func(ctx context.Context, tx *gorm.DB, link *DropshipLink) (*DropshipLink, error)
}

func (m DropshipLinkDAOMock) ListByPurchaseOrder(ctx context.Context, orderID uint64) ([]*DropshipLink, error) {
	if m.ListByPurchaseOrderFunc != nil {
		return m.ListByPurchaseOrderFunc(ctx, orderID)
	}
	return nil, nil
}

func (m DropshipLinkDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, link *DropshipLink) (*DropshipLink, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, link)
	}
	return link, nil
}

type InterorganizationRuleDAOMock struct {
	dao.CRUDMock[InterorganizationRule]
}

type InterorganizationTransactionDAOMock struct {
	dao.CRUDMock[InterorganizationTransaction]
}

type ConsolidationRunDAOMock struct {
	dao.CRUDMock[ConsolidationRun]
	UpdateTxFunc func(ctx context.Context, tx *gorm.DB, run *ConsolidationRun) (*ConsolidationRun, error)
}

func (m ConsolidationRunDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, run *ConsolidationRun) (*ConsolidationRun, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, run)
	}
	return run, nil
}

type ConsolidationEliminationDAOMock struct {
	dao.CRUDMock[ConsolidationElimination]
	ListByRunFunc func(ctx context.Context, runID uint64) ([]*ConsolidationElimination, error)
	CreateTxFunc  func(ctx context.Context, tx *gorm.DB, elimination *ConsolidationElimination) (*ConsolidationElimination, error)
}

func (m ConsolidationEliminationDAOMock) ListByRun(ctx context.Context, runID uint64) ([]*ConsolidationElimination, error) {
	if m.ListByRunFunc != nil {
		return m.ListByRunFunc(ctx, runID)
	}
	return nil, nil
}

func (m ConsolidationEliminationDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, elimination *ConsolidationElimination) (*ConsolidationElimination, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, elimination)
	}
	return elimination, nil
}
