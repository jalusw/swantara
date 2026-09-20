package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type AccountFullReconcileDAOMock struct {
	dao.CRUDMock[AccountFullReconcile]
	CreateTxFunc func(ctx context.Context, tx *gorm.DB, reconcile *AccountFullReconcile) (*AccountFullReconcile, error)
}

func (m AccountFullReconcileDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, reconcile *AccountFullReconcile) (*AccountFullReconcile, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, reconcile)
	}
	return reconcile, nil
}

type AccountPartialReconcileDAOMock struct {
	dao.CRUDMock[AccountPartialReconcile]
	CreateTxFunc    func(ctx context.Context, tx *gorm.DB, reconcile *AccountPartialReconcile) (*AccountPartialReconcile, error)
	UpdateTxFunc    func(ctx context.Context, tx *gorm.DB, reconcile *AccountPartialReconcile) (*AccountPartialReconcile, error)
	TotalByLineFunc func(ctx context.Context, lineID uint64) (float64, error)
}

func (m AccountPartialReconcileDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, reconcile *AccountPartialReconcile) (*AccountPartialReconcile, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, reconcile)
	}
	return reconcile, nil
}

func (m AccountPartialReconcileDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, reconcile *AccountPartialReconcile) (*AccountPartialReconcile, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, reconcile)
	}
	return reconcile, nil
}

func (m AccountPartialReconcileDAOMock) TotalByLine(ctx context.Context, lineID uint64) (float64, error) {
	if m.TotalByLineFunc != nil {
		return m.TotalByLineFunc(ctx, lineID)
	}
	return 0, nil
}

type ReminderActionDAOMock struct {
	dao.CRUDMock[ReminderAction]
	FindByInvoiceLevelFunc func(ctx context.Context, invoiceID, levelID uint64) (*ReminderAction, error)
	ListByInvoiceFunc      func(ctx context.Context, invoiceID uint64) ([]*ReminderAction, error)
}

func (m ReminderActionDAOMock) FindByInvoiceLevel(ctx context.Context, invoiceID, levelID uint64) (*ReminderAction, error) {
	if m.FindByInvoiceLevelFunc != nil {
		return m.FindByInvoiceLevelFunc(ctx, invoiceID, levelID)
	}
	return nil, nil
}

func (m ReminderActionDAOMock) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*ReminderAction, error) {
	if m.ListByInvoiceFunc != nil {
		return m.ListByInvoiceFunc(ctx, invoiceID)
	}
	return nil, nil
}
