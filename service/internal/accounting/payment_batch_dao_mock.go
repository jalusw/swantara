package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type PaymentBatchDAOMock struct {
	dao.CRUDMock[PaymentBatch]
	CreateTxFunc func(ctx context.Context, tx *gorm.DB, batch *PaymentBatch) (*PaymentBatch, error)
	UpdateTxFunc func(ctx context.Context, tx *gorm.DB, batch *PaymentBatch) (*PaymentBatch, error)
}

func (m PaymentBatchDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, batch *PaymentBatch) (*PaymentBatch, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, batch)
	}
	return m.Create(ctx, batch)
}

func (m PaymentBatchDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, batch *PaymentBatch) (*PaymentBatch, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, batch)
	}
	return batch, nil
}

type PaymentBatchLineDAOMock struct {
	dao.CRUDMock[PaymentBatchLine]
	CreateBatchLinesTxFunc func(ctx context.Context, tx *gorm.DB, lines []*PaymentBatchLine, batchID uint64) error
	ListByBatchFunc        func(ctx context.Context, batchID uint64) ([]*PaymentBatchLine, error)
}

func (m PaymentBatchLineDAOMock) CreateBatchLinesTx(ctx context.Context, tx *gorm.DB, lines []*PaymentBatchLine, batchID uint64) error {
	if m.CreateBatchLinesTxFunc != nil {
		return m.CreateBatchLinesTxFunc(ctx, tx, lines, batchID)
	}
	return nil
}

func (m PaymentBatchLineDAOMock) ListByBatch(ctx context.Context, batchID uint64) ([]*PaymentBatchLine, error) {
	if m.ListByBatchFunc != nil {
		return m.ListByBatchFunc(ctx, batchID)
	}
	return nil, nil
}
