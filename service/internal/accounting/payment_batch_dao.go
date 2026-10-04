package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type PaymentBatchDAO interface {
	dao.CRUD[PaymentBatch]
	CreateTx(ctx context.Context, tx *gorm.DB, batch *PaymentBatch) (*PaymentBatch, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, batch *PaymentBatch) (*PaymentBatch, error)
}

type paymentBatchDAO struct {
	dao.Base[PaymentBatch]
	db *gorm.DB
}

func NewPaymentBatchDAO(db *gorm.DB) PaymentBatchDAO {
	return paymentBatchDAO{Base: dao.NewBase[PaymentBatch](db), db: db}
}

func (d paymentBatchDAO) CreateTx(ctx context.Context, tx *gorm.DB, batch *PaymentBatch) (*PaymentBatch, error) {
	if err := tx.WithContext(ctx).Create(batch).Error; err != nil {
		return nil, err
	}
	return batch, nil
}

func (d paymentBatchDAO) UpdateTx(ctx context.Context, tx *gorm.DB, batch *PaymentBatch) (*PaymentBatch, error) {
	if err := tx.WithContext(ctx).Save(batch).Error; err != nil {
		return nil, err
	}
	return batch, nil
}

type PaymentBatchLineDAO interface {
	dao.CRUD[PaymentBatchLine]
	CreateBatchLinesTx(ctx context.Context, tx *gorm.DB, lines []*PaymentBatchLine, batchID uint64) error
	ListByBatch(ctx context.Context, batchID uint64) ([]*PaymentBatchLine, error)
}

type paymentBatchLineDAO struct {
	dao.Base[PaymentBatchLine]
	db *gorm.DB
}

func NewPaymentBatchLineDAO(db *gorm.DB) PaymentBatchLineDAO {
	return paymentBatchLineDAO{Base: dao.NewBase[PaymentBatchLine](db), db: db}
}

func (d paymentBatchLineDAO) CreateBatchLinesTx(ctx context.Context, tx *gorm.DB, lines []*PaymentBatchLine, batchID uint64) error {
	for _, line := range lines {
		line.BatchID = batchID
		if err := tx.WithContext(ctx).Create(line).Error; err != nil {
			return err
		}
	}
	return nil
}

func (d paymentBatchLineDAO) ListByBatch(ctx context.Context, batchID uint64) ([]*PaymentBatchLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "batch_id", Operator: query.Equal, Value: batchID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
