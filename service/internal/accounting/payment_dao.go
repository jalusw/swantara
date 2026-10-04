package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type PaymentDAO interface {
	dao.CRUD[Payment]
	CreateWithAllocationsTx(ctx context.Context, tx *gorm.DB, payment *Payment, allocations []*PaymentAllocation) (*Payment, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, payment *Payment) (*Payment, error)
	ListPostedByContact(ctx context.Context, contactID uint64) ([]*Payment, error)
}

type paymentDAO struct {
	dao.Base[Payment]
	db *gorm.DB
}

func NewPaymentDAO(db *gorm.DB) PaymentDAO {
	return paymentDAO{Base: dao.NewBase[Payment](db), db: db}
}

func (d paymentDAO) UpdateTx(ctx context.Context, tx *gorm.DB, payment *Payment) (*Payment, error) {
	if err := tx.WithContext(ctx).Save(payment).Error; err != nil {
		return nil, err
	}
	return payment, nil
}

func (d paymentDAO) CreateWithAllocationsTx(ctx context.Context, tx *gorm.DB, payment *Payment, allocations []*PaymentAllocation) (*Payment, error) {
	if err := tx.WithContext(ctx).Create(payment).Error; err != nil {
		return nil, err
	}
	for _, allocation := range allocations {
		allocation.PaymentID = payment.ID
		if err := tx.WithContext(ctx).Create(allocation).Error; err != nil {
			return nil, err
		}
	}
	return payment, nil
}

func (d paymentDAO) ListPostedByContact(ctx context.Context, contactID uint64) ([]*Payment, error) {
	var entities []Payment
	if err := d.db.WithContext(ctx).Where("contact_id = ? AND state = ?", contactID, PaymentStatePosted).Find(&entities).Error; err != nil {
		return nil, err
	}
	items := make([]*Payment, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

type PaymentAllocationDAO interface {
	dao.CRUD[PaymentAllocation]
	DeleteTx(ctx context.Context, tx *gorm.DB, id uint64) error
	ListByPayment(ctx context.Context, paymentID uint64) ([]*PaymentAllocation, error)
	ListByInvoice(ctx context.Context, invoiceID uint64) ([]*PaymentAllocation, error)
}

type paymentAllocationDAO struct {
	dao.Base[PaymentAllocation]
	db *gorm.DB
}

func NewPaymentAllocationDAO(db *gorm.DB) PaymentAllocationDAO {
	return paymentAllocationDAO{Base: dao.NewBase[PaymentAllocation](db), db: db}
}

func (d paymentAllocationDAO) DeleteTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&PaymentAllocation{}, id).Error
}

func (d paymentAllocationDAO) ListByPayment(ctx context.Context, paymentID uint64) ([]*PaymentAllocation, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "payment_id", Operator: query.Equal, Value: paymentID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d paymentAllocationDAO) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*PaymentAllocation, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "invoice_id", Operator: query.Equal, Value: invoiceID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
