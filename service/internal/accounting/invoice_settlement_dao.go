package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type InvoiceCreditApplicationDAO interface {
	dao.CRUD[InvoiceCreditApplication]
	ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceCreditApplication, error)
}

type invoiceCreditApplicationDAO struct {
	dao.Base[InvoiceCreditApplication]
	db *gorm.DB
}

func NewInvoiceCreditApplicationDAO(db *gorm.DB) InvoiceCreditApplicationDAO {
	return invoiceCreditApplicationDAO{Base: dao.NewBase[InvoiceCreditApplication](db), db: db}
}

func (d invoiceCreditApplicationDAO) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceCreditApplication, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "invoice_id", Operator: query.Equal, Value: invoiceID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type InvoiceContraSettlementDAO interface {
	dao.CRUD[InvoiceContraSettlement]
	ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceContraSettlement, error)
}

type invoiceContraSettlementDAO struct {
	dao.Base[InvoiceContraSettlement]
	db *gorm.DB
}

func NewInvoiceContraSettlementDAO(db *gorm.DB) InvoiceContraSettlementDAO {
	return invoiceContraSettlementDAO{Base: dao.NewBase[InvoiceContraSettlement](db), db: db}
}

func (d invoiceContraSettlementDAO) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceContraSettlement, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "customer_invoice_id", Operator: query.Equal, Value: invoiceID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type DownPaymentLinkDAO interface {
	dao.CRUD[DownPaymentLink]
	ListByFinal(ctx context.Context, finalInvoiceID uint64) ([]*DownPaymentLink, error)
}

type downPaymentLinkDAO struct {
	dao.Base[DownPaymentLink]
	db *gorm.DB
}

func NewDownPaymentLinkDAO(db *gorm.DB) DownPaymentLinkDAO {
	return downPaymentLinkDAO{Base: dao.NewBase[DownPaymentLink](db), db: db}
}

func (d downPaymentLinkDAO) ListByFinal(ctx context.Context, finalInvoiceID uint64) ([]*DownPaymentLink, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "final_invoice_id", Operator: query.Equal, Value: finalInvoiceID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
