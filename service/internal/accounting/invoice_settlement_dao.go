package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type InvoiceCreditApplicationDAO interface {
	dao.CRUD[InvoiceCreditApplication]
	CreateTx(ctx context.Context, tx *gorm.DB, application *InvoiceCreditApplication) (*InvoiceCreditApplication, error)
	ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceCreditApplication, error)
}

type invoiceCreditApplicationDAO struct {
	dao.Base[InvoiceCreditApplication]
	db *gorm.DB
}

func NewInvoiceCreditApplicationDAO(db *gorm.DB) InvoiceCreditApplicationDAO {
	return invoiceCreditApplicationDAO{Base: dao.NewBase[InvoiceCreditApplication](db), db: db}
}

func (d invoiceCreditApplicationDAO) CreateTx(ctx context.Context, tx *gorm.DB, application *InvoiceCreditApplication) (*InvoiceCreditApplication, error) {
	if err := tx.WithContext(ctx).Create(application).Error; err != nil {
		return nil, err
	}
	return application, nil
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
	CreateTx(ctx context.Context, tx *gorm.DB, settlement *InvoiceContraSettlement) (*InvoiceContraSettlement, error)
	ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceContraSettlement, error)
}

type invoiceContraSettlementDAO struct {
	dao.Base[InvoiceContraSettlement]
	db *gorm.DB
}

func NewInvoiceContraSettlementDAO(db *gorm.DB) InvoiceContraSettlementDAO {
	return invoiceContraSettlementDAO{Base: dao.NewBase[InvoiceContraSettlement](db), db: db}
}

func (d invoiceContraSettlementDAO) CreateTx(ctx context.Context, tx *gorm.DB, settlement *InvoiceContraSettlement) (*InvoiceContraSettlement, error) {
	if err := tx.WithContext(ctx).Create(settlement).Error; err != nil {
		return nil, err
	}
	return settlement, nil
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
	CreateTx(ctx context.Context, tx *gorm.DB, link *DownPaymentLink) (*DownPaymentLink, error)
	ListByFinal(ctx context.Context, finalInvoiceID uint64) ([]*DownPaymentLink, error)
}

type downPaymentLinkDAO struct {
	dao.Base[DownPaymentLink]
	db *gorm.DB
}

func NewDownPaymentLinkDAO(db *gorm.DB) DownPaymentLinkDAO {
	return downPaymentLinkDAO{Base: dao.NewBase[DownPaymentLink](db), db: db}
}

func (d downPaymentLinkDAO) CreateTx(ctx context.Context, tx *gorm.DB, link *DownPaymentLink) (*DownPaymentLink, error) {
	if err := tx.WithContext(ctx).Create(link).Error; err != nil {
		return nil, err
	}
	return link, nil
}

func (d downPaymentLinkDAO) ListByFinal(ctx context.Context, finalInvoiceID uint64) ([]*DownPaymentLink, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "final_invoice_id", Operator: query.Equal, Value: finalInvoiceID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
