package accounting

import (
	"context"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type InvoiceDAO interface {
	dao.CRUD[Invoice]
	CreateWithLinesTx(ctx context.Context, tx *gorm.DB, invoice *Invoice, lines []*InvoiceLine, taxes []*InvoiceTax) (*Invoice, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, invoice *Invoice) (*Invoice, error)
	ListOpenByContact(ctx context.Context, contactID uint64) ([]*Invoice, error)
	ListOpenByOrganization(ctx context.Context, organizationID uint64) ([]*Invoice, error)
	FindByOrigin(ctx context.Context, originInvoiceID uint64) (*Invoice, error)
	FindBySaleOrder(ctx context.Context, orderID uint64) (*Invoice, error)
	FindByPurchaseOrder(ctx context.Context, orderID uint64) (*Invoice, error)
	ListOverdue(ctx context.Context, asOf *time.Time) ([]*Invoice, error)
}

type invoiceDAO struct {
	dao.Base[Invoice]
	db *gorm.DB
}

func NewInvoiceDAO(db *gorm.DB) InvoiceDAO {
	return invoiceDAO{Base: dao.NewBase[Invoice](db), db: db}
}

func (d invoiceDAO) UpdateTx(ctx context.Context, tx *gorm.DB, invoice *Invoice) (*Invoice, error) {
	if err := tx.WithContext(ctx).Save(invoice).Error; err != nil {
		return nil, err
	}
	return invoice, nil
}

func (d invoiceDAO) ListOpenByContact(ctx context.Context, contactID uint64) ([]*Invoice, error) {
	var entities []Invoice
	if err := d.db.WithContext(ctx).
		Where("contact_id = ? AND state = ? AND payment_state NOT IN ?", contactID, InvoiceStatePosted, []string{PaymentStatePaid, PaymentStateBadDebt}).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	items := make([]*Invoice, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

func (d invoiceDAO) FindByOrigin(ctx context.Context, originInvoiceID uint64) (*Invoice, error) {
	var entity Invoice
	if err := d.db.WithContext(ctx).Where("origin_invoice_id = ?", originInvoiceID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (d invoiceDAO) FindBySaleOrder(ctx context.Context, orderID uint64) (*Invoice, error) {
	var entity Invoice
	query := d.db.WithContext(ctx).
		Joins("JOIN invoice_lines ON invoice_lines.invoice_id = invoices.id").
		Joins("JOIN sale_order_lines ON sale_order_lines.id = invoice_lines.sale_line_id").
		Where("sale_order_lines.order_id = ? AND invoices.state = ?", orderID, InvoiceStatePosted).
		Order("invoices.id DESC").
		Limit(1).
		First(&entity)
	if err := query.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (d invoiceDAO) FindByPurchaseOrder(ctx context.Context, orderID uint64) (*Invoice, error) {
	var entity Invoice
	query := d.db.WithContext(ctx).
		Joins("JOIN invoice_lines ON invoice_lines.invoice_id = invoices.id").
		Joins("JOIN purchase_order_lines ON purchase_order_lines.id = invoice_lines.purchase_line_id").
		Where("purchase_order_lines.order_id = ? AND invoices.state = ?", orderID, InvoiceStatePosted).
		Order("invoices.id DESC").
		Limit(1).
		First(&entity)
	if err := query.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (d invoiceDAO) ListOverdue(ctx context.Context, asOf *time.Time) ([]*Invoice, error) {
	date := time.Now().UTC()
	if asOf != nil {
		date = *asOf
	}
	var entities []Invoice
	if err := d.db.WithContext(ctx).
		Where("state = ? AND payment_state NOT IN ? AND due_date IS NOT NULL AND due_date < ?", InvoiceStatePosted, []string{PaymentStatePaid, PaymentStateBadDebt}, date).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	items := make([]*Invoice, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

func (d invoiceDAO) ListOpenByOrganization(ctx context.Context, organizationID uint64) ([]*Invoice, error) {
	var entities []Invoice
	if err := d.db.WithContext(ctx).
		Where("organization_id = ? AND type = ? AND state = ? AND payment_state NOT IN ?", organizationID, InvoiceTypeCustomerInvoice, InvoiceStatePosted, []string{PaymentStatePaid, PaymentStateBadDebt}).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	items := make([]*Invoice, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

func (d invoiceDAO) CreateWithLinesTx(ctx context.Context, tx *gorm.DB, invoice *Invoice, lines []*InvoiceLine, taxes []*InvoiceTax) (*Invoice, error) {
	if err := tx.WithContext(ctx).Create(invoice).Error; err != nil {
		return nil, err
	}
	for _, line := range lines {
		line.InvoiceID = invoice.ID
		if err := tx.WithContext(ctx).Create(line).Error; err != nil {
			return nil, err
		}
	}
	for _, tax := range taxes {
		tax.InvoiceID = invoice.ID
		if err := tx.WithContext(ctx).Create(tax).Error; err != nil {
			return nil, err
		}
	}
	return invoice, nil
}

type InvoiceLineDAO interface {
	dao.CRUD[InvoiceLine]
	ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceLine, error)
}

type invoiceLineDAO struct {
	dao.Base[InvoiceLine]
	db *gorm.DB
}

func NewInvoiceLineDAO(db *gorm.DB) InvoiceLineDAO {
	return invoiceLineDAO{Base: dao.NewBase[InvoiceLine](db), db: db}
}

func (d invoiceLineDAO) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "invoice_id", Operator: query.Equal, Value: invoiceID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type InvoiceTaxDAO interface {
	dao.CRUD[InvoiceTax]
	ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceTax, error)
	SumTaxByPeriod(ctx context.Context, organizationID uint64, invoiceType string, start, end time.Time) (float64, error)
}

type invoiceTaxDAO struct {
	dao.Base[InvoiceTax]
	db *gorm.DB
}

func NewInvoiceTaxDAO(db *gorm.DB) InvoiceTaxDAO {
	return invoiceTaxDAO{Base: dao.NewBase[InvoiceTax](db), db: db}
}

func (d invoiceTaxDAO) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceTax, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "invoice_id", Operator: query.Equal, Value: invoiceID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d invoiceTaxDAO) SumTaxByPeriod(ctx context.Context, organizationID uint64, invoiceType string, start, end time.Time) (float64, error) {
	var total float64
	if err := d.db.WithContext(ctx).
		Model(&InvoiceTax{}).
		Joins("JOIN invoices ON invoices.id = invoice_taxes.invoice_id").
		Where("invoices.organization_id = ? AND invoices.type = ? AND invoices.state = ?", organizationID, invoiceType, InvoiceStatePosted).
		Where("invoices.invoice_date >= ? AND invoices.invoice_date <= ?", start, end).
		Select("COALESCE(SUM(invoice_taxes.amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
