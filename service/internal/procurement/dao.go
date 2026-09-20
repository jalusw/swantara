package procurement

import (
	"context"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type PurchaseRequestDAO interface {
	dao.CRUD[PurchaseRequest]
	CreateWithLines(ctx context.Context, requisition *PurchaseRequest, lines []*PurchaseRequestLine) (*PurchaseRequest, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, requisition *PurchaseRequest) (*PurchaseRequest, error)
}

type purchaseRequestDAO struct {
	dao.Base[PurchaseRequest]
	db *gorm.DB
}

func NewPurchaseRequestDAO(db *gorm.DB) PurchaseRequestDAO {
	return purchaseRequestDAO{Base: dao.NewBase[PurchaseRequest](db), db: db}
}

func (d purchaseRequestDAO) CreateWithLines(ctx context.Context, requisition *PurchaseRequest, lines []*PurchaseRequestLine) (*PurchaseRequest, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(requisition).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.RequestID = requisition.ID
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return requisition, nil
}

func (d purchaseRequestDAO) UpdateTx(ctx context.Context, tx *gorm.DB, requisition *PurchaseRequest) (*PurchaseRequest, error) {
	if err := tx.WithContext(ctx).Save(requisition).Error; err != nil {
		return nil, err
	}
	return requisition, nil
}

type PurchaseRequestLineDAO interface {
	dao.CRUD[PurchaseRequestLine]
	ListByRequest(ctx context.Context, requestID uint64) ([]*PurchaseRequestLine, error)
	ReplaceLines(ctx context.Context, requestID uint64, lines []*PurchaseRequestLine) error
}

type purchaseRequestLineDAO struct {
	dao.Base[PurchaseRequestLine]
	db *gorm.DB
}

func NewPurchaseRequestLineDAO(db *gorm.DB) PurchaseRequestLineDAO {
	return purchaseRequestLineDAO{Base: dao.NewBase[PurchaseRequestLine](db), db: db}
}

func (d purchaseRequestLineDAO) ListByRequest(ctx context.Context, requestID uint64) ([]*PurchaseRequestLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "request_id", Operator: query.Equal, Value: requestID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d purchaseRequestLineDAO) ReplaceLines(ctx context.Context, requestID uint64, lines []*PurchaseRequestLine) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("request_id = ?", requestID).Delete(&PurchaseRequestLine{}).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.RequestID = requestID
			line.ID = 0
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type PurchaseOrderDAO interface {
	dao.CRUD[PurchaseOrder]
	CreateWithLines(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, order *PurchaseOrder) (*PurchaseOrder, error)
}

type purchaseOrderDAO struct {
	dao.Base[PurchaseOrder]
	db *gorm.DB
}

func NewPurchaseOrderDAO(db *gorm.DB) PurchaseOrderDAO {
	return purchaseOrderDAO{Base: dao.NewBase[PurchaseOrder](db), db: db}
}

func (d purchaseOrderDAO) CreateWithLines(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.OrderID = order.ID
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (d purchaseOrderDAO) UpdateTx(ctx context.Context, tx *gorm.DB, order *PurchaseOrder) (*PurchaseOrder, error) {
	if err := tx.WithContext(ctx).Save(order).Error; err != nil {
		return nil, err
	}
	return order, nil
}

type PurchaseOrderLineDAO interface {
	dao.CRUD[PurchaseOrderLine]
	ListByOrder(ctx context.Context, orderID uint64) ([]*PurchaseOrderLine, error)
	ReplaceLines(ctx context.Context, orderID uint64, lines []*PurchaseOrderLine) error
	UpdateTx(ctx context.Context, tx *gorm.DB, line *PurchaseOrderLine) (*PurchaseOrderLine, error)
}

type purchaseOrderLineDAO struct {
	dao.Base[PurchaseOrderLine]
	db *gorm.DB
}

func NewPurchaseOrderLineDAO(db *gorm.DB) PurchaseOrderLineDAO {
	return purchaseOrderLineDAO{Base: dao.NewBase[PurchaseOrderLine](db), db: db}
}

func (d purchaseOrderLineDAO) ListByOrder(ctx context.Context, orderID uint64) ([]*PurchaseOrderLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "order_id", Operator: query.Equal, Value: orderID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d purchaseOrderLineDAO) ReplaceLines(ctx context.Context, orderID uint64, lines []*PurchaseOrderLine) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", orderID).Delete(&PurchaseOrderLine{}).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.OrderID = orderID
			line.ID = 0
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (d purchaseOrderLineDAO) UpdateTx(ctx context.Context, tx *gorm.DB, line *PurchaseOrderLine) (*PurchaseOrderLine, error) {
	if err := tx.WithContext(ctx).Save(line).Error; err != nil {
		return nil, err
	}
	return line, nil
}

type SupplierQuoteRequestDAO interface {
	dao.CRUD[SupplierQuoteRequest]
	CreateWithLines(ctx context.Context, quoteRequest *SupplierQuoteRequest, lines []*SupplierQuoteRequestLine) (*SupplierQuoteRequest, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, quoteRequest *SupplierQuoteRequest) (*SupplierQuoteRequest, error)
}

type purchaseRFQDAO struct {
	dao.Base[SupplierQuoteRequest]
	db *gorm.DB
}

func NewSupplierQuoteRequestDAO(db *gorm.DB) SupplierQuoteRequestDAO {
	return purchaseRFQDAO{Base: dao.NewBase[SupplierQuoteRequest](db), db: db}
}

func (d purchaseRFQDAO) CreateWithLines(ctx context.Context, quoteRequest *SupplierQuoteRequest, lines []*SupplierQuoteRequestLine) (*SupplierQuoteRequest, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(quoteRequest).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.QuoteRequestID = quoteRequest.ID
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return quoteRequest, nil
}

func (d purchaseRFQDAO) UpdateTx(ctx context.Context, tx *gorm.DB, quoteRequest *SupplierQuoteRequest) (*SupplierQuoteRequest, error) {
	if err := tx.WithContext(ctx).Save(quoteRequest).Error; err != nil {
		return nil, err
	}
	return quoteRequest, nil
}

type SupplierQuoteRequestLineDAO interface {
	dao.CRUD[SupplierQuoteRequestLine]
	ListByRFQ(ctx context.Context, rfqID uint64) ([]*SupplierQuoteRequestLine, error)
}

type purchaseRFQLineDAO struct {
	dao.Base[SupplierQuoteRequestLine]
}

func NewSupplierQuoteRequestLineDAO(db *gorm.DB) SupplierQuoteRequestLineDAO {
	return purchaseRFQLineDAO{Base: dao.NewBase[SupplierQuoteRequestLine](db)}
}

func (d purchaseRFQLineDAO) ListByRFQ(ctx context.Context, rfqID uint64) ([]*SupplierQuoteRequestLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "quoteRequest_id", Operator: query.Equal, Value: rfqID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type SupplierQuoteDAO interface {
	dao.CRUD[SupplierQuote]
	CreateWithLines(ctx context.Context, quote *SupplierQuote, lines []*SupplierQuoteLine) (*SupplierQuote, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, quote *SupplierQuote) (*SupplierQuote, error)
}

type purchaseRFQQuoteDAO struct {
	dao.Base[SupplierQuote]
	db *gorm.DB
}

func NewSupplierQuoteDAO(db *gorm.DB) SupplierQuoteDAO {
	return purchaseRFQQuoteDAO{Base: dao.NewBase[SupplierQuote](db), db: db}
}

func (d purchaseRFQQuoteDAO) CreateWithLines(ctx context.Context, quote *SupplierQuote, lines []*SupplierQuoteLine) (*SupplierQuote, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(quote).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.SupplierQuoteID = quote.ID
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return quote, nil
}

func (d purchaseRFQQuoteDAO) UpdateTx(ctx context.Context, tx *gorm.DB, quote *SupplierQuote) (*SupplierQuote, error) {
	if err := tx.WithContext(ctx).Save(quote).Error; err != nil {
		return nil, err
	}
	return quote, nil
}

type SupplierQuoteLineDAO interface {
	dao.CRUD[SupplierQuoteLine]
	ListByQuote(ctx context.Context, quoteID uint64) ([]*SupplierQuoteLine, error)
}

type purchaseRFQQuoteLineDAO struct {
	dao.Base[SupplierQuoteLine]
}

func NewSupplierQuoteLineDAO(db *gorm.DB) SupplierQuoteLineDAO {
	return purchaseRFQQuoteLineDAO{Base: dao.NewBase[SupplierQuoteLine](db)}
}

func (d purchaseRFQQuoteLineDAO) ListByQuote(ctx context.Context, quoteID uint64) ([]*SupplierQuoteLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "supplier_quote_id", Operator: query.Equal, Value: quoteID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type CurrencyRateDAO interface {
	dao.CRUD[CurrencyRate]
	FindByPair(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64) (*CurrencyRate, error)
	FindByPairDate(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64, rateDate time.Time) (*CurrencyRate, error)
}

type currencyRateDAO struct {
	dao.Base[CurrencyRate]
	db *gorm.DB
}

func NewCurrencyRateDAO(db *gorm.DB) CurrencyRateDAO {
	return currencyRateDAO{Base: dao.NewBase[CurrencyRate](db), db: db}
}

func (d currencyRateDAO) FindByPair(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64) (*CurrencyRate, error) {
	return d.findLatest(ctx, fromCurrency, toCurrency, orgID, time.Now())
}

func (d currencyRateDAO) FindByPairDate(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64, rateDate time.Time) (*CurrencyRate, error) {
	return d.findLatest(ctx, fromCurrency, toCurrency, orgID, rateDate)
}

func (d currencyRateDAO) findLatest(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64, maxDate time.Time) (*CurrencyRate, error) {
	q := d.db.WithContext(ctx).
		Where("from_currency = ? AND to_currency = ? AND rate_date <= ?", fromCurrency, toCurrency, maxDate)
	if orgID != nil {
		q = q.Where("organization_id = ?", *orgID)
	} else {
		q = q.Where("organization_id IS NULL")
	}
	var rate CurrencyRate
	if err := q.Order("rate_date DESC").First(&rate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rate, nil
}

type SupplyAgreementDAO interface {
	dao.CRUD[SupplyAgreement]
	CreateWithLines(ctx context.Context, agreement *SupplyAgreement, lines []*SupplyAgreementLine) (*SupplyAgreement, error)
}

type supplyAgreementDAO struct {
	dao.Base[SupplyAgreement]
	db *gorm.DB
}

func NewSupplyAgreementDAO(db *gorm.DB) SupplyAgreementDAO {
	return supplyAgreementDAO{Base: dao.NewBase[SupplyAgreement](db), db: db}
}

func (d supplyAgreementDAO) CreateWithLines(ctx context.Context, agreement *SupplyAgreement, lines []*SupplyAgreementLine) (*SupplyAgreement, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(agreement).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.AgreementID = agreement.ID
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return agreement, nil
}

type SupplyAgreementLineDAO interface {
	dao.CRUD[SupplyAgreementLine]
	ListByAgreement(ctx context.Context, agreementID uint64) ([]*SupplyAgreementLine, error)
}

type supplyAgreementLineDAO struct {
	dao.Base[SupplyAgreementLine]
	db *gorm.DB
}

func NewSupplyAgreementLineDAO(db *gorm.DB) SupplyAgreementLineDAO {
	return supplyAgreementLineDAO{Base: dao.NewBase[SupplyAgreementLine](db), db: db}
}

func (d supplyAgreementLineDAO) ListByAgreement(ctx context.Context, agreementID uint64) ([]*SupplyAgreementLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "agreement_id", Operator: query.Equal, Value: agreementID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type SupplierScorecardDAO interface {
	dao.CRUD[SupplierScorecard]
	FindByVendorAndPeriod(ctx context.Context, supplierID uint64, periodStart, periodEnd *time.Time) (*SupplierScorecard, error)
}

type vendorScorecardDAO struct {
	dao.Base[SupplierScorecard]
	db *gorm.DB
}

func NewSupplierScorecardDAO(db *gorm.DB) SupplierScorecardDAO {
	return vendorScorecardDAO{Base: dao.NewBase[SupplierScorecard](db), db: db}
}

func (d vendorScorecardDAO) FindByVendorAndPeriod(ctx context.Context, supplierID uint64, periodStart, periodEnd *time.Time) (*SupplierScorecard, error) {
	q := d.db.WithContext(ctx).Where("supplier_id = ?", supplierID)
	if periodStart != nil {
		q = q.Where("period_start = ?", *periodStart)
	}
	if periodEnd != nil {
		q = q.Where("period_end = ?", *periodEnd)
	}
	var scorecard SupplierScorecard
	if err := q.First(&scorecard).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &scorecard, nil
}

type CostCenterDAO interface {
	dao.CRUD[CostCenter]
}

type costCenterDAO struct {
	dao.Base[CostCenter]
}

func NewCostCenterDAO(db *gorm.DB) CostCenterDAO {
	return costCenterDAO{Base: dao.NewBase[CostCenter](db)}
}

type PurchaseCreditMemoDAO interface {
	dao.CRUD[PurchaseCreditMemo]
}

type purchaseCreditMemoDAO struct {
	dao.Base[PurchaseCreditMemo]
}

func NewPurchaseCreditMemoDAO(db *gorm.DB) PurchaseCreditMemoDAO {
	return purchaseCreditMemoDAO{Base: dao.NewBase[PurchaseCreditMemo](db)}
}

type PurchaseDebitMemoDAO interface {
	dao.CRUD[PurchaseDebitMemo]
}

type purchaseDebitMemoDAO struct {
	dao.Base[PurchaseDebitMemo]
}

func NewPurchaseDebitMemoDAO(db *gorm.DB) PurchaseDebitMemoDAO {
	return purchaseDebitMemoDAO{Base: dao.NewBase[PurchaseDebitMemo](db)}
}

type PaymentBatchDAO interface {
	dao.CRUD[PaymentBatch]
	CreateWithLines(ctx context.Context, batch *PaymentBatch, lines []*PaymentBatchLine) (*PaymentBatch, error)
}

type paymentBatchDAO struct {
	dao.Base[PaymentBatch]
	db *gorm.DB
}

func NewPaymentBatchDAO(db *gorm.DB) PaymentBatchDAO {
	return paymentBatchDAO{Base: dao.NewBase[PaymentBatch](db), db: db}
}

func (d paymentBatchDAO) CreateWithLines(ctx context.Context, batch *PaymentBatch, lines []*PaymentBatchLine) (*PaymentBatch, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(batch).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.BatchID = batch.ID
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return batch, nil
}

type PaymentBatchLineDAO interface {
	dao.CRUD[PaymentBatchLine]
	ListByBatch(ctx context.Context, batchID uint64) ([]*PaymentBatchLine, error)
}

type paymentBatchLineDAO struct {
	dao.Base[PaymentBatchLine]
	db *gorm.DB
}

func NewPaymentBatchLineDAO(db *gorm.DB) PaymentBatchLineDAO {
	return paymentBatchLineDAO{Base: dao.NewBase[PaymentBatchLine](db), db: db}
}

func (d paymentBatchLineDAO) ListByBatch(ctx context.Context, batchID uint64) ([]*PaymentBatchLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "batch_id", Operator: query.Equal, Value: batchID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
