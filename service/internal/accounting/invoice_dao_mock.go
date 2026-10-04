package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type InvoiceDAOMock struct {
	dao.CRUDMock[Invoice]
	CreateWithLinesTxFunc      func(ctx context.Context, tx *gorm.DB, invoice *Invoice, lines []*InvoiceLine, taxes []*InvoiceTax) (*Invoice, error)
	UpdateTxFunc               func(ctx context.Context, tx *gorm.DB, invoice *Invoice) (*Invoice, error)
	ListOpenByContactFunc      func(ctx context.Context, contactID uint64) ([]*Invoice, error)
	ListOpenByOrganizationFunc func(ctx context.Context, organizationID uint64) ([]*Invoice, error)
	FindByOriginFunc           func(ctx context.Context, originInvoiceID uint64) (*Invoice, error)
	FindBySaleOrderFunc        func(ctx context.Context, orderID uint64) (*Invoice, error)
	FindByPurchaseOrderFunc    func(ctx context.Context, orderID uint64) (*Invoice, error)
	ListOverdueFunc            func(ctx context.Context, asOf *time.Time) ([]*Invoice, error)
}

func (m InvoiceDAOMock) CreateWithLinesTx(ctx context.Context, tx *gorm.DB, invoice *Invoice, lines []*InvoiceLine, taxes []*InvoiceTax) (*Invoice, error) {
	if m.CreateWithLinesTxFunc != nil {
		return m.CreateWithLinesTxFunc(ctx, tx, invoice, lines, taxes)
	}
	return invoice, nil
}

func (m InvoiceDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, invoice *Invoice) (*Invoice, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, invoice)
	}
	return invoice, nil
}

func (m InvoiceDAOMock) ListOpenByContact(ctx context.Context, contactID uint64) ([]*Invoice, error) {
	if m.ListOpenByContactFunc != nil {
		return m.ListOpenByContactFunc(ctx, contactID)
	}
	return nil, nil
}

func (m InvoiceDAOMock) ListOpenByOrganization(ctx context.Context, organizationID uint64) ([]*Invoice, error) {
	if m.ListOpenByOrganizationFunc != nil {
		return m.ListOpenByOrganizationFunc(ctx, organizationID)
	}
	return nil, nil
}

func (m InvoiceDAOMock) FindByOrigin(ctx context.Context, originInvoiceID uint64) (*Invoice, error) {
	if m.FindByOriginFunc != nil {
		return m.FindByOriginFunc(ctx, originInvoiceID)
	}
	return nil, nil
}

func (m InvoiceDAOMock) FindBySaleOrder(ctx context.Context, orderID uint64) (*Invoice, error) {
	if m.FindBySaleOrderFunc != nil {
		return m.FindBySaleOrderFunc(ctx, orderID)
	}
	return nil, nil
}

func (m InvoiceDAOMock) FindByPurchaseOrder(ctx context.Context, orderID uint64) (*Invoice, error) {
	if m.FindByPurchaseOrderFunc != nil {
		return m.FindByPurchaseOrderFunc(ctx, orderID)
	}
	return nil, nil
}

func (m InvoiceDAOMock) ListOverdue(ctx context.Context, asOf *time.Time) ([]*Invoice, error) {
	if m.ListOverdueFunc != nil {
		return m.ListOverdueFunc(ctx, asOf)
	}
	return nil, nil
}

type InvoiceLineDAOMock struct {
	dao.CRUDMock[InvoiceLine]
	ListByInvoiceFunc func(ctx context.Context, invoiceID uint64) ([]*InvoiceLine, error)
}

func (m InvoiceLineDAOMock) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceLine, error) {
	if m.ListByInvoiceFunc != nil {
		return m.ListByInvoiceFunc(ctx, invoiceID)
	}
	return nil, nil
}

type InvoiceInstallmentDAOMock struct {
	dao.CRUDMock[InvoiceInstallment]
	CreateManyTxFunc  func(ctx context.Context, tx *gorm.DB, installments []*InvoiceInstallment) error
	ListByInvoiceFunc func(ctx context.Context, invoiceID uint64) ([]*InvoiceInstallment, error)
}

func (m InvoiceInstallmentDAOMock) CreateManyTx(ctx context.Context, tx *gorm.DB, installments []*InvoiceInstallment) error {
	if m.CreateManyTxFunc != nil {
		return m.CreateManyTxFunc(ctx, tx, installments)
	}
	return nil
}

func (m InvoiceInstallmentDAOMock) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceInstallment, error) {
	if m.ListByInvoiceFunc != nil {
		return m.ListByInvoiceFunc(ctx, invoiceID)
	}
	return nil, nil
}

type InvoiceCreditApplicationDAOMock struct {
	dao.CRUDMock[InvoiceCreditApplication]
	CreateTxFunc      func(ctx context.Context, tx *gorm.DB, application *InvoiceCreditApplication) (*InvoiceCreditApplication, error)
	ListByInvoiceFunc func(ctx context.Context, invoiceID uint64) ([]*InvoiceCreditApplication, error)
}

func (m InvoiceCreditApplicationDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, application *InvoiceCreditApplication) (*InvoiceCreditApplication, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, application)
	}
	return m.Create(ctx, application)
}

func (m InvoiceCreditApplicationDAOMock) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceCreditApplication, error) {
	if m.ListByInvoiceFunc != nil {
		return m.ListByInvoiceFunc(ctx, invoiceID)
	}
	return nil, nil
}

type InvoiceContraSettlementDAOMock struct {
	dao.CRUDMock[InvoiceContraSettlement]
	CreateTxFunc      func(ctx context.Context, tx *gorm.DB, settlement *InvoiceContraSettlement) (*InvoiceContraSettlement, error)
	ListByInvoiceFunc func(ctx context.Context, invoiceID uint64) ([]*InvoiceContraSettlement, error)
}

func (m InvoiceContraSettlementDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, settlement *InvoiceContraSettlement) (*InvoiceContraSettlement, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, settlement)
	}
	return m.Create(ctx, settlement)
}

func (m InvoiceContraSettlementDAOMock) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceContraSettlement, error) {
	if m.ListByInvoiceFunc != nil {
		return m.ListByInvoiceFunc(ctx, invoiceID)
	}
	return nil, nil
}

type DownPaymentLinkDAOMock struct {
	dao.CRUDMock[DownPaymentLink]
	CreateTxFunc    func(ctx context.Context, tx *gorm.DB, link *DownPaymentLink) (*DownPaymentLink, error)
	ListByFinalFunc func(ctx context.Context, finalInvoiceID uint64) ([]*DownPaymentLink, error)
}

func (m DownPaymentLinkDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, link *DownPaymentLink) (*DownPaymentLink, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, link)
	}
	return m.Create(ctx, link)
}

func (m DownPaymentLinkDAOMock) ListByFinal(ctx context.Context, finalInvoiceID uint64) ([]*DownPaymentLink, error) {
	if m.ListByFinalFunc != nil {
		return m.ListByFinalFunc(ctx, finalInvoiceID)
	}
	return nil, nil
}

type InvoiceTaxDAOMock struct {
	dao.CRUDMock[InvoiceTax]
	ListByInvoiceFunc  func(ctx context.Context, invoiceID uint64) ([]*InvoiceTax, error)
	SumTaxByPeriodFunc func(ctx context.Context, organizationID uint64, invoiceType string, start, end time.Time) (float64, error)
}

func (m InvoiceTaxDAOMock) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*InvoiceTax, error) {
	if m.ListByInvoiceFunc != nil {
		return m.ListByInvoiceFunc(ctx, invoiceID)
	}
	return nil, nil
}

func (m InvoiceTaxDAOMock) SumTaxByPeriod(ctx context.Context, organizationID uint64, invoiceType string, start, end time.Time) (float64, error) {
	if m.SumTaxByPeriodFunc != nil {
		return m.SumTaxByPeriodFunc(ctx, organizationID, invoiceType, start, end)
	}
	return 0, nil
}
