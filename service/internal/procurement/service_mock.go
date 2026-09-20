package procurement

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type OrderCreatorMock struct {
	CreateFunc func(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error)
}

func (m OrderCreatorMock) Create(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, order, lines)
	}
	return order, nil
}

type CurrencyConverterMock struct {
	ConvertFunc func(ctx context.Context, amount amount.Amount, fromCurrency, toCurrency string, orgID *uint64, rateDate time.Time) (amount.Amount, error)
	HasRateFunc func(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64) (bool, error)
}

func (m CurrencyConverterMock) Convert(ctx context.Context, amt amount.Amount, fromCurrency, toCurrency string, orgID *uint64, rateDate time.Time) (amount.Amount, error) {
	if m.ConvertFunc != nil {
		return m.ConvertFunc(ctx, amt, fromCurrency, toCurrency, orgID, rateDate)
	}
	return amt, nil
}

func (m CurrencyConverterMock) HasRate(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64) (bool, error) {
	if m.HasRateFunc != nil {
		return m.HasRateFunc(ctx, fromCurrency, toCurrency, orgID)
	}
	return fromCurrency == toCurrency, nil
}

type PurchaseOrderServiceTestDeps struct {
	Orders           PurchaseOrderDAOMock
	Lines            PurchaseOrderLineDAOMock
	Requisitions     PurchaseRequestDAOMock
	RequisitionLines PurchaseRequestLineDAOMock
	Agreements       SupplyAgreementDAOMock
	AgreementLines   SupplyAgreementLineDAOMock
	CreditMemos      PurchaseCreditMemoDAOMock
	DebitMemos       PurchaseDebitMemoDAOMock
	Batches          PaymentBatchDAOMock
	BatchLines       PaymentBatchLineDAOMock
	Contacts         contacts.ContactDAOMock
	Suppliers        SupplierProfileLookupMock
	Warehouses       inventory.WarehouseDAOMock
	Locations        inventory.StockLocationDAOMock
	Shipments        inventory.ShipmentDAOMock
	Movements        inventory.StockMovementDAOMock
	Resolver         inventory.ItemResolverMock
	Expense          ExpenseAccountEngineMock
	Taxes            dao.CRUDMock[reference.Tax]
	Offers           OfferEngineMock
	Configs          ConfigLookupMock
	Approvals        ApprovalEngineMock
	Receive          ReceiveEngineMock
	Bills            BillEngineMock
	OpenInvoices     OpenInvoiceLookupMock
	Payments         OutboundPaymentEngineMock
	Quality          QualityEngineMock
	Sequences        sequence.DAOMock
}

func NewTestPurchaseOrderService(deps PurchaseOrderServiceTestDeps) PurchaseOrderService {
	sequences := deps.Sequences
	if sequences.ReserveFunc == nil {
		sequences = sequence.DAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
				return &sequence.Reservation{Value: 1, Number: "PO/00001"}, nil
			},
		}
	}
	contactsDAO := deps.Contacts
	if contactsDAO.FindFunc == nil {
		contactsDAO = contacts.ContactDAOMock{
			CRUDMock: dao.CRUDMock[contacts.Contact]{
				FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
					return &contacts.Contact{Base: model.Base{ID: 10}}, nil
				},
			},
		}
	}
	warehouses := deps.Warehouses
	if warehouses.FindFunc == nil {
		warehouses = inventory.WarehouseDAOMock{
			CRUDMock: dao.CRUDMock[reference.Warehouse]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
					return &reference.Warehouse{Base: model.Base{ID: 20}, OrganizationID: helper.Ptr(uint64(10))}, nil
				},
			},
		}
	}
	expense := deps.Expense
	if expense.ResolveExpenseAccountFunc == nil {
		expense = ExpenseAccountEngineMock{
			ResolveExpenseAccountFunc: func(_ context.Context, _ uint64) (uint64, error) {
				return 6000, nil
			},
		}
	}
	return NewPurchaseOrderService(
		deps.Orders,
		deps.Lines,
		deps.Requisitions,
		deps.RequisitionLines,
		deps.Agreements,
		deps.AgreementLines,
		deps.CreditMemos,
		deps.DebitMemos,
		deps.Batches,
		deps.BatchLines,
		sequence.NewSequenceService(sequences),
		NewSystemConfigSource(deps.Configs),
		deps.Offers,
		contactsDAO,
		deps.Suppliers,
		warehouses,
		deps.Locations,
		deps.Shipments,
		deps.Movements,
		deps.Resolver,
		expense,
		deps.Taxes,
		deps.Receive,
		deps.Approvals,
		deps.Bills,
		deps.OpenInvoices,
		deps.Payments,
		deps.Quality,
		CurrencyConverterMock{},
	)
}

type PurchaseRequestServiceTestDeps struct {
	Requisitions PurchaseRequestDAOMock
	Lines        PurchaseRequestLineDAOMock
	Contacts     contacts.ContactDAOMock
	Sequences    sequence.DAOMock
}

func NewTestPurchaseRequestService(deps PurchaseRequestServiceTestDeps) PurchaseRequestService {
	sequences := deps.Sequences
	if sequences.ReserveFunc == nil {
		sequences = sequence.DAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
				return &sequence.Reservation{Value: 1, Number: "REQ/00001"}, nil
			},
		}
	}
	contactsDAO := deps.Contacts
	if contactsDAO.FindFunc == nil {
		contactsDAO = contacts.ContactDAOMock{
			CRUDMock: dao.CRUDMock[contacts.Contact]{
				FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
					return &contacts.Contact{Base: model.Base{ID: 5}}, nil
				},
			},
		}
	}
	return NewPurchaseRequestService(
		deps.Requisitions,
		deps.Lines,
		sequence.NewSequenceService(sequences),
		contactsDAO,
	)
}

type SupplierQuoteRequestServiceTestDeps struct {
	QuoteRequests    SupplierQuoteRequestDAOMock
	RFQLines         SupplierQuoteRequestLineDAOMock
	Quotes           SupplierQuoteDAOMock
	QuoteLines       SupplierQuoteLineDAOMock
	Requisitions     PurchaseRequestDAOMock
	RequisitionLines PurchaseRequestLineDAOMock
	Contacts         contacts.ContactDAOMock
	Sequences        sequence.DAOMock
	Orders           OrderCreator
}

func NewTestSupplierQuoteRequestService(deps SupplierQuoteRequestServiceTestDeps) SupplierQuoteRequestService {
	sequences := deps.Sequences
	if sequences.ReserveFunc == nil {
		sequences = sequence.DAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
				return &sequence.Reservation{Value: 1, Number: "QuoteRequest/00001"}, nil
			},
		}
	}
	contactsDAO := deps.Contacts
	if contactsDAO.FindFunc == nil {
		contactsDAO = contacts.ContactDAOMock{
			CRUDMock: dao.CRUDMock[contacts.Contact]{
				FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
					return &contacts.Contact{Base: model.Base{ID: 5}}, nil
				},
			},
		}
	}
	orders := deps.Orders
	if orders == nil {
		orders = OrderCreatorMock{}
	}
	return NewSupplierQuoteRequestService(
		deps.QuoteRequests,
		deps.RFQLines,
		deps.Quotes,
		deps.QuoteLines,
		deps.Requisitions,
		deps.RequisitionLines,
		sequence.NewSequenceService(sequences),
		contactsDAO,
		orders,
	)
}

func NewTestApprovalConfig(organizationID uint64, threshold float64, approverIDs []uint64) ConfigLookupMock {
	items := []*reference.SystemConfig{}
	if threshold > 0 {
		items = append(items, &reference.SystemConfig{
			Base:           model.Base{ID: 1},
			OrganizationID: &organizationID,
			Key:            configApprovalThreshold,
			Value:          []byte(fmt.Sprintf("%f", threshold)),
		})
	}
	if len(approverIDs) > 0 {
		value, _ := json.Marshal(approverIDs)
		items = append(items, &reference.SystemConfig{
			Base:           model.Base{ID: 2},
			OrganizationID: &organizationID,
			Key:            configApproverIDs,
			Value:          value,
		})
	}
	return ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: items, Count: int64(len(items))}, nil
		},
	}
}
