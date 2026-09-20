package procurement

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type PurchaseRequestDAOMock struct {
	dao.CRUDMock[PurchaseRequest]
	CreateWithLinesFunc func(ctx context.Context, requisition *PurchaseRequest, lines []*PurchaseRequestLine) (*PurchaseRequest, error)
	UpdateTxFunc        func(ctx context.Context, tx *gorm.DB, requisition *PurchaseRequest) (*PurchaseRequest, error)
}

func (m PurchaseRequestDAOMock) CreateWithLines(ctx context.Context, requisition *PurchaseRequest, lines []*PurchaseRequestLine) (*PurchaseRequest, error) {
	if m.CreateWithLinesFunc != nil {
		return m.CreateWithLinesFunc(ctx, requisition, lines)
	}
	return requisition, nil
}

func (m PurchaseRequestDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, requisition *PurchaseRequest) (*PurchaseRequest, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, requisition)
	}
	return requisition, nil
}

type PurchaseRequestLineDAOMock struct {
	dao.CRUDMock[PurchaseRequestLine]
	ListByRequestFunc func(ctx context.Context, requestID uint64) ([]*PurchaseRequestLine, error)
	ReplaceLinesFunc  func(ctx context.Context, requestID uint64, lines []*PurchaseRequestLine) error
}

func (m PurchaseRequestLineDAOMock) ListByRequest(ctx context.Context, requestID uint64) ([]*PurchaseRequestLine, error) {
	if m.ListByRequestFunc != nil {
		return m.ListByRequestFunc(ctx, requestID)
	}
	return []*PurchaseRequestLine{}, nil
}

func (m PurchaseRequestLineDAOMock) ReplaceLines(ctx context.Context, requestID uint64, lines []*PurchaseRequestLine) error {
	if m.ReplaceLinesFunc != nil {
		return m.ReplaceLinesFunc(ctx, requestID, lines)
	}
	return nil
}

type PurchaseOrderDAOMock struct {
	dao.CRUDMock[PurchaseOrder]
	CreateWithLinesFunc func(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error)
	UpdateTxFunc        func(ctx context.Context, tx *gorm.DB, order *PurchaseOrder) (*PurchaseOrder, error)
}

func (m PurchaseOrderDAOMock) CreateWithLines(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error) {
	if m.CreateWithLinesFunc != nil {
		return m.CreateWithLinesFunc(ctx, order, lines)
	}
	return order, nil
}

func (m PurchaseOrderDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, order *PurchaseOrder) (*PurchaseOrder, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, order)
	}
	return order, nil
}

type PurchaseOrderLineDAOMock struct {
	dao.CRUDMock[PurchaseOrderLine]
	ListByOrderFunc  func(ctx context.Context, orderID uint64) ([]*PurchaseOrderLine, error)
	ReplaceLinesFunc func(ctx context.Context, orderID uint64, lines []*PurchaseOrderLine) error
	UpdateTxFunc     func(ctx context.Context, tx *gorm.DB, line *PurchaseOrderLine) (*PurchaseOrderLine, error)
}

func (m PurchaseOrderLineDAOMock) ListByOrder(ctx context.Context, orderID uint64) ([]*PurchaseOrderLine, error) {
	if m.ListByOrderFunc != nil {
		return m.ListByOrderFunc(ctx, orderID)
	}
	return []*PurchaseOrderLine{}, nil
}

func (m PurchaseOrderLineDAOMock) ReplaceLines(ctx context.Context, orderID uint64, lines []*PurchaseOrderLine) error {
	if m.ReplaceLinesFunc != nil {
		return m.ReplaceLinesFunc(ctx, orderID, lines)
	}
	return nil
}

func (m PurchaseOrderLineDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, line *PurchaseOrderLine) (*PurchaseOrderLine, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, line)
	}
	return line, nil
}

type SupplierQuoteRequestDAOMock struct {
	dao.CRUDMock[SupplierQuoteRequest]
	CreateWithLinesFunc func(ctx context.Context, quoteRequest *SupplierQuoteRequest, lines []*SupplierQuoteRequestLine) (*SupplierQuoteRequest, error)
	UpdateTxFunc        func(ctx context.Context, tx *gorm.DB, quoteRequest *SupplierQuoteRequest) (*SupplierQuoteRequest, error)
}

func (m SupplierQuoteRequestDAOMock) CreateWithLines(ctx context.Context, quoteRequest *SupplierQuoteRequest, lines []*SupplierQuoteRequestLine) (*SupplierQuoteRequest, error) {
	if m.CreateWithLinesFunc != nil {
		return m.CreateWithLinesFunc(ctx, quoteRequest, lines)
	}
	return quoteRequest, nil
}

func (m SupplierQuoteRequestDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, quoteRequest *SupplierQuoteRequest) (*SupplierQuoteRequest, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, quoteRequest)
	}
	return quoteRequest, nil
}

type SupplierQuoteRequestLineDAOMock struct {
	dao.CRUDMock[SupplierQuoteRequestLine]
	ListByRFQFunc func(ctx context.Context, rfqID uint64) ([]*SupplierQuoteRequestLine, error)
}

func (m SupplierQuoteRequestLineDAOMock) ListByRFQ(ctx context.Context, rfqID uint64) ([]*SupplierQuoteRequestLine, error) {
	if m.ListByRFQFunc != nil {
		return m.ListByRFQFunc(ctx, rfqID)
	}
	return []*SupplierQuoteRequestLine{}, nil
}

type SupplierQuoteDAOMock struct {
	dao.CRUDMock[SupplierQuote]
	CreateWithLinesFunc func(ctx context.Context, quote *SupplierQuote, lines []*SupplierQuoteLine) (*SupplierQuote, error)
	UpdateTxFunc        func(ctx context.Context, tx *gorm.DB, quote *SupplierQuote) (*SupplierQuote, error)
}

func (m SupplierQuoteDAOMock) CreateWithLines(ctx context.Context, quote *SupplierQuote, lines []*SupplierQuoteLine) (*SupplierQuote, error) {
	if m.CreateWithLinesFunc != nil {
		return m.CreateWithLinesFunc(ctx, quote, lines)
	}
	return quote, nil
}

func (m SupplierQuoteDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, quote *SupplierQuote) (*SupplierQuote, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, quote)
	}
	return quote, nil
}

type SupplierQuoteLineDAOMock struct {
	dao.CRUDMock[SupplierQuoteLine]
	ListByQuoteFunc func(ctx context.Context, quoteID uint64) ([]*SupplierQuoteLine, error)
}

func (m SupplierQuoteLineDAOMock) ListByQuote(ctx context.Context, quoteID uint64) ([]*SupplierQuoteLine, error) {
	if m.ListByQuoteFunc != nil {
		return m.ListByQuoteFunc(ctx, quoteID)
	}
	return []*SupplierQuoteLine{}, nil
}

type CurrencyRateDAOMock struct {
	dao.CRUDMock[CurrencyRate]
	FindByPairFunc     func(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64) (*CurrencyRate, error)
	FindByPairDateFunc func(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64, rateDate time.Time) (*CurrencyRate, error)
}

func (m CurrencyRateDAOMock) FindByPair(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64) (*CurrencyRate, error) {
	if m.FindByPairFunc != nil {
		return m.FindByPairFunc(ctx, fromCurrency, toCurrency, orgID)
	}
	return nil, nil
}

func (m CurrencyRateDAOMock) FindByPairDate(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64, rateDate time.Time) (*CurrencyRate, error) {
	if m.FindByPairDateFunc != nil {
		return m.FindByPairDateFunc(ctx, fromCurrency, toCurrency, orgID, rateDate)
	}
	return nil, nil
}

type SupplyAgreementDAOMock struct {
	dao.CRUDMock[SupplyAgreement]
	CreateWithLinesFunc func(ctx context.Context, agreement *SupplyAgreement, lines []*SupplyAgreementLine) (*SupplyAgreement, error)
}

func (m SupplyAgreementDAOMock) CreateWithLines(ctx context.Context, agreement *SupplyAgreement, lines []*SupplyAgreementLine) (*SupplyAgreement, error) {
	if m.CreateWithLinesFunc != nil {
		return m.CreateWithLinesFunc(ctx, agreement, lines)
	}
	agreement.ID = 1
	return agreement, nil
}

type SupplyAgreementLineDAOMock struct {
	dao.CRUDMock[SupplyAgreementLine]
	ListByAgreementFunc func(ctx context.Context, agreementID uint64) ([]*SupplyAgreementLine, error)
}

func (m SupplyAgreementLineDAOMock) ListByAgreement(ctx context.Context, agreementID uint64) ([]*SupplyAgreementLine, error) {
	if m.ListByAgreementFunc != nil {
		return m.ListByAgreementFunc(ctx, agreementID)
	}
	return []*SupplyAgreementLine{}, nil
}

type SupplierScorecardDAOMock struct {
	dao.CRUDMock[SupplierScorecard]
	FindByVendorAndPeriodFunc func(ctx context.Context, supplierID uint64, periodStart, periodEnd *time.Time) (*SupplierScorecard, error)
}

func (m SupplierScorecardDAOMock) FindByVendorAndPeriod(ctx context.Context, supplierID uint64, periodStart, periodEnd *time.Time) (*SupplierScorecard, error) {
	if m.FindByVendorAndPeriodFunc != nil {
		return m.FindByVendorAndPeriodFunc(ctx, supplierID, periodStart, periodEnd)
	}
	return nil, nil
}

type PurchaseCreditMemoDAOMock struct {
	dao.CRUDMock[PurchaseCreditMemo]
}

type PurchaseDebitMemoDAOMock struct {
	dao.CRUDMock[PurchaseDebitMemo]
}

type PaymentBatchDAOMock struct {
	dao.CRUDMock[PaymentBatch]
	CreateWithLinesFunc func(ctx context.Context, batch *PaymentBatch, lines []*PaymentBatchLine) (*PaymentBatch, error)
}

func (m PaymentBatchDAOMock) CreateWithLines(ctx context.Context, batch *PaymentBatch, lines []*PaymentBatchLine) (*PaymentBatch, error) {
	if m.CreateWithLinesFunc != nil {
		return m.CreateWithLinesFunc(ctx, batch, lines)
	}
	batch.ID = 1
	return batch, nil
}

type PaymentBatchLineDAOMock struct {
	dao.CRUDMock[PaymentBatchLine]
	ListByBatchFunc func(ctx context.Context, batchID uint64) ([]*PaymentBatchLine, error)
}

func (m PaymentBatchLineDAOMock) ListByBatch(ctx context.Context, batchID uint64) ([]*PaymentBatchLine, error) {
	if m.ListByBatchFunc != nil {
		return m.ListByBatchFunc(ctx, batchID)
	}
	return []*PaymentBatchLine{}, nil
}
