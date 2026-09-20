package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type InvoiceInstallmentDAO interface {
	dao.CRUD[InvoiceInstallment]
	CreateManyTx(ctx context.Context, tx *gorm.DB, installments []*InvoiceInstallment) error
	ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceInstallment, error)
}

type invoiceInstallmentDAO struct {
	dao.Base[InvoiceInstallment]
	db *gorm.DB
}

func NewInvoiceInstallmentDAO(db *gorm.DB) InvoiceInstallmentDAO {
	return invoiceInstallmentDAO{Base: dao.NewBase[InvoiceInstallment](db), db: db}
}

func (d invoiceInstallmentDAO) CreateManyTx(ctx context.Context, tx *gorm.DB, installments []*InvoiceInstallment) error {
	for _, installment := range installments {
		if err := tx.WithContext(ctx).Create(installment).Error; err != nil {
			return err
		}
	}
	return nil
}

func (d invoiceInstallmentDAO) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceInstallment, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "invoice_id", Operator: query.Equal, Value: invoiceID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
