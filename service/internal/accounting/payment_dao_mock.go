package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type PaymentDAOMock struct {
	dao.CRUDMock[Payment]
	CreateWithAllocationsTxFunc func(ctx context.Context, tx *gorm.DB, payment *Payment, allocations []*PaymentAllocation) (*Payment, error)
	UpdateTxFunc                func(ctx context.Context, tx *gorm.DB, payment *Payment) (*Payment, error)
	ListPostedByContactFunc     func(ctx context.Context, contactID uint64) ([]*Payment, error)
}

func (m PaymentDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, payment *Payment) (*Payment, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, payment)
	}
	return m.Update(ctx, payment)
}

func (m PaymentDAOMock) CreateWithAllocationsTx(ctx context.Context, tx *gorm.DB, payment *Payment, allocations []*PaymentAllocation) (*Payment, error) {
	if m.CreateWithAllocationsTxFunc != nil {
		return m.CreateWithAllocationsTxFunc(ctx, tx, payment, allocations)
	}
	return payment, nil
}

func (m PaymentDAOMock) ListPostedByContact(ctx context.Context, contactID uint64) ([]*Payment, error) {
	if m.ListPostedByContactFunc != nil {
		return m.ListPostedByContactFunc(ctx, contactID)
	}
	return nil, nil
}

type PdcInstrumentDAOMock struct {
	dao.CRUDMock[PdcInstrument]
	UpdateTxFunc func(ctx context.Context, tx *gorm.DB, instrument *PdcInstrument) (*PdcInstrument, error)
	ListDueFunc  func(ctx context.Context, asOf *time.Time) ([]*PdcInstrument, error)
}

func (m PdcInstrumentDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, instrument *PdcInstrument) (*PdcInstrument, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, instrument)
	}
	return instrument, nil
}

func (m PdcInstrumentDAOMock) ListDue(ctx context.Context, asOf *time.Time) ([]*PdcInstrument, error) {
	if m.ListDueFunc != nil {
		return m.ListDueFunc(ctx, asOf)
	}
	return nil, nil
}

type PaymentAllocationDAOMock struct {
	dao.CRUDMock[PaymentAllocation]
	DeleteTxFunc      func(ctx context.Context, tx *gorm.DB, id uint64) error
	ListByPaymentFunc func(ctx context.Context, paymentID uint64) ([]*PaymentAllocation, error)
	ListByInvoiceFunc func(ctx context.Context, invoiceID uint64) ([]*PaymentAllocation, error)
}

func (m PaymentAllocationDAOMock) DeleteTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	if m.DeleteTxFunc != nil {
		return m.DeleteTxFunc(ctx, tx, id)
	}
	return m.Delete(ctx, id)
}

func (m PaymentAllocationDAOMock) ListByPayment(ctx context.Context, paymentID uint64) ([]*PaymentAllocation, error) {
	if m.ListByPaymentFunc != nil {
		return m.ListByPaymentFunc(ctx, paymentID)
	}
	return nil, nil
}

func (m PaymentAllocationDAOMock) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*PaymentAllocation, error) {
	if m.ListByInvoiceFunc != nil {
		return m.ListByInvoiceFunc(ctx, invoiceID)
	}
	return nil, nil
}
