package procurement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/crosscutting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestPriceLines(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() PurchaseOrderService
		lines     []*PurchaseOrderLine
		expectErr bool
		errTarget error
	}{
		{
			name: "RejectsZeroQty",
			setup: func() PurchaseOrderService {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				return svc
			},
			lines:     []*PurchaseOrderLine{{QtyOrdered: 0, ItemID: helper.Ptr(uint64(200)), UnitPrice: 10}},
			expectErr: true,
			errTarget: ErrPurchaseOrderLineQty,
		},
		{
			name: "RejectsDiscountOutOfRange",
			setup: func() PurchaseOrderService {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				return svc
			},
			lines: func() []*PurchaseOrderLine {
				l := purchaseLine(1)
				l.DiscountPct = 110
				return []*PurchaseOrderLine{l}
			}(),
			expectErr: true,
			errTarget: ErrPurchaseOrderLineDiscount,
		},
		{
			name: "RejectsUnpricedDescriptionLine",
			setup: func() PurchaseOrderService {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				return svc
			},
			lines:     []*PurchaseOrderLine{{QtyOrdered: 1, UnitPrice: 0}},
			expectErr: true,
			errTarget: ErrPurchaseOrderNoOffer,
		},
		{
			name: "AcceptsPricedDescriptionLine",
			setup: func() PurchaseOrderService {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				return svc
			},
			lines:     []*PurchaseOrderLine{{QtyOrdered: 1, UnitPrice: 50}},
			expectErr: false,
		},
		{
			name: "PropagatesResolverError",
			setup: func() PurchaseOrderService {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				svc.resolver = inventory.ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
						return inventory.ResolvedItem{}, errors.New("db down")
					},
				}
				return svc
			},
			lines:     []*PurchaseOrderLine{{QtyOrdered: 1, ItemID: helper.Ptr(uint64(200)), UnitPrice: 10}},
			expectErr: true,
		},
		{
			name: "RejectsWhenOfferFails",
			setup: func() PurchaseOrderService {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				svc.offers = OfferEngineMock{
					BestOfferForSupplierFunc: func(_ context.Context, _ uint64, _ uint64, _ amount.Amount, _ time.Time) (*products.SupplierProduct, error) {
						return nil, ErrPurchaseOrderNoOffer
					},
				}
				return svc
			},
			lines:     []*PurchaseOrderLine{{QtyOrdered: 1, ItemID: helper.Ptr(uint64(200)), UnitPrice: 0}},
			expectErr: true,
			errTarget: ErrPurchaseOrderNoOffer,
		},
		{
			name: "RejectsNilOffer",
			setup: func() PurchaseOrderService {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				svc.offers = OfferEngineMock{
					BestOfferForSupplierFunc: func(_ context.Context, _ uint64, _ uint64, _ amount.Amount, _ time.Time) (*products.SupplierProduct, error) {
						return nil, nil
					},
				}
				return svc
			},
			lines:     []*PurchaseOrderLine{{QtyOrdered: 1, ItemID: helper.Ptr(uint64(200)), UnitPrice: 0}},
			expectErr: true,
			errTarget: ErrPurchaseOrderNoOffer,
		},
		{
			name: "RejectsOfferWithoutPrice",
			setup: func() PurchaseOrderService {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				svc.offers = OfferEngineMock{
					BestOfferForSupplierFunc: func(_ context.Context, _ uint64, _ uint64, _ amount.Amount, _ time.Time) (*products.SupplierProduct, error) {
						return &products.SupplierProduct{ItemID: 200}, nil
					},
				}
				return svc
			},
			lines:     []*PurchaseOrderLine{{QtyOrdered: 1, ItemID: helper.Ptr(uint64(200)), UnitPrice: 0}},
			expectErr: true,
			errTarget: ErrPurchaseOrderNoOffer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			order := draftOrder()
			date := time.Now()
			err := svc.priceLines(context.Background(), order, &date, tt.lines)
			if tt.expectErr {
				helper.AssertError(t, err, true, tt.errTarget)
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateVendor(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() PurchaseOrderService
		expectErr bool
		errTarget error
	}{
		{
			name: "RejectsMissingContact",
			setup: func() PurchaseOrderService {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				svc.contacts = contacts.ContactDAOMock{
					CRUDMock: dao.CRUDMock[contacts.Contact]{
						FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
							return nil, nil
						},
					},
				}
				return svc
			},
			expectErr: true,
			errTarget: ErrPurchaseOrderVendor,
		},
		{
			name: "PropagatesSupplierLookupError",
			setup: func() PurchaseOrderService {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				svc.suppliers = SupplierProfileLookupMock{
					FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
						return nil, errors.New("db down")
					},
				}
				return svc
			},
			expectErr: true,
		},
		{
			name: "RejectsMissingSupplier",
			setup: func() PurchaseOrderService {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				svc.suppliers = SupplierProfileLookupMock{
					FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
						return nil, nil
					},
				}
				return svc
			},
			expectErr: true,
			errTarget: ErrPurchaseOrderVendorNotSupplier,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			order := draftOrder()
			err := svc.validateVendor(context.Background(), order)
			helper.AssertError(t, err, tt.expectErr, tt.errTarget)
		})
	}
}

func TestFindOrder(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(orders *PurchaseOrderDAOMock)
		expectErr bool
		errTarget error
	}{
		{
			name: "PropagatesError",
			setup: func(orders *PurchaseOrderDAOMock) {
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) {
						return nil, errors.New("db down")
					},
				}
			},
			expectErr: true,
		},
		{
			name: "RejectsMissingOrder",
			setup: func(orders *PurchaseOrderDAOMock) {
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) {
						return nil, nil
					},
				}
			},
			expectErr: true,
			errTarget: ErrPurchaseOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orders, _, _, _, _, _, _, _ := testPurchaseOrderService()
			tt.setup(orders)
			_, err := svc.findOrder(context.Background(), 1)
			helper.AssertError(t, err, tt.expectErr, tt.errTarget)
		})
	}
}

func TestRequestApproval(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(svc *PurchaseOrderService, approvals *ApprovalEngineMock)
		expectErr bool
		errTarget error
	}{
		{
			name: "PropagatesApproverLookupError",
			setup: func(svc *PurchaseOrderService, approvals *ApprovalEngineMock) {
				svc.configs = NewSystemConfigSource(ConfigLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
						return nil, errors.New("db down")
					},
				})
			},
			expectErr: true,
		},
		{
			name: "PropagatesApprovalCreateError",
			setup: func(svc *PurchaseOrderService, approvals *ApprovalEngineMock) {
				svc.configs = NewSystemConfigSource(ConfigLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
						return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
							{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Key: configApproverIDs, Value: []byte(`[7]`)},
						}}, nil
					},
				})
				approvals.CreateFunc = func(_ context.Context, _ uint64, _ string, _ uint64, _ uint64, _ []uint64) (*crosscutting.ApprovalRequest, error) {
					return nil, errors.New("create failed")
				}
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _, approvals, _, _, _, _, _ := testPurchaseOrderService()
			tt.setup(&svc, approvals)
			order := draftOrder()
			err := svc.requestApproval(context.Background(), order, 9)
			helper.AssertError(t, err, tt.expectErr, tt.errTarget)
		})
	}
}

func TestFindQuote(t *testing.T) {
	tests := []struct {
		name      string
		expectErr bool
		errTarget error
	}{
		{
			name:      "PropagatesError",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testRFQService(SupplierQuoteRequestDAOMock{}, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{
				CRUDMock: dao.CRUDMock[SupplierQuote]{
					FindFunc: func(_ context.Context, _ uint64) (*SupplierQuote, error) {
						return nil, errors.New("db down")
					},
				},
			}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)
			_, err := svc.findQuote(context.Background(), 1)
			helper.AssertError(t, err, tt.expectErr, tt.errTarget)
		})
	}
}

func TestRFQCancel(t *testing.T) {
	tests := []struct {
		name         string
		quoteRequest SupplierQuoteRequest
		expectErr    bool
		errTarget    error
	}{
		{
			name:         "RejectsDoneRFQ",
			quoteRequest: SupplierQuoteRequest{Base: model.Base{ID: 1}, State: QuoteRequestStateDone},
			expectErr:    true,
			errTarget:    ErrQuoteRequestState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quote_requests := SupplierQuoteRequestDAOMock{
				CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
					FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
						return &tt.quoteRequest, nil
					},
				},
			}
			svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)
			_, err := svc.Cancel(context.Background(), 1)
			helper.AssertError(t, err, tt.expectErr, tt.errTarget)
		})
	}
}

func TestRFQCreate(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(svc *SupplierQuoteRequestService)
		quoteRequest *SupplierQuoteRequest
		lines        []*SupplierQuoteRequestLine
		expectErr    bool
		errTarget    error
	}{
		{
			name:         "RejectsMissingRequester",
			setup:        func(svc *SupplierQuoteRequestService) {},
			quoteRequest: &SupplierQuoteRequest{OrganizationID: helper.Ptr(uint64(1)), RequesterID: 0},
			lines:        []*SupplierQuoteRequestLine{{Qty: 1}},
			expectErr:    true,
			errTarget:    ErrRFQRequester,
		},
		{
			name: "RejectsUnknownRequester",
			setup: func(svc *SupplierQuoteRequestService) {
				svc.contacts = contacts.ContactDAOMock{
					CRUDMock: dao.CRUDMock[contacts.Contact]{
						FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
							return nil, nil
						},
					},
				}
			},
			quoteRequest: &SupplierQuoteRequest{OrganizationID: helper.Ptr(uint64(1)), RequesterID: 5},
			lines:        []*SupplierQuoteRequestLine{{Qty: 1}},
			expectErr:    true,
			errTarget:    ErrRFQRequester,
		},
		{
			name: "PropagatesRequesterError",
			setup: func(svc *SupplierQuoteRequestService) {
				svc.contacts = contacts.ContactDAOMock{
					CRUDMock: dao.CRUDMock[contacts.Contact]{
						FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
							return nil, errors.New("db down")
						},
					},
				}
			},
			quoteRequest: &SupplierQuoteRequest{OrganizationID: helper.Ptr(uint64(1)), RequesterID: 5},
			lines:        []*SupplierQuoteRequestLine{{Qty: 1}},
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testRFQService(SupplierQuoteRequestDAOMock{}, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)
			tt.setup(&svc)
			_, err := svc.Create(context.Background(), tt.quoteRequest, tt.lines)
			helper.AssertError(t, err, tt.expectErr, tt.errTarget)
		})
	}
}

func TestCreateSupplierBill(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(svc *PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, bills *BillEngineMock)
		expectErr bool
		errTarget error
	}{
		{
			name: "FallsBackToExpenseAccount",
			setup: func(svc *PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, bills *BillEngineMock) {
				order := draftOrder()
				order.State = PurchaseOrderStateConfirmed
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, o *PurchaseOrder) (*PurchaseOrder, error) {
						return o, nil
					},
				}
				line := purchaseLine(5)
				line.QtyReceived = 5
				lines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*PurchaseOrderLine, error) {
					return []*PurchaseOrderLine{line}, nil
				}
				svc.resolver = inventory.ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
						return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockInputAccountID: 0}}, nil
					},
				}
				svc.expense = ExpenseAccountEngineMock{
					ResolveExpenseAccountFunc: func(_ context.Context, _ uint64) (uint64, error) {
						return 6000, nil
					},
				}
				bills.CreateSupplierBillFunc = func(_ context.Context, req accounting.CreateSupplierBillRequest) (*accounting.Invoice, error) {
					return &accounting.Invoice{Base: model.Base{ID: 500}, ContactID: req.ContactID}, nil
				}
			},
			expectErr: false,
		},
		{
			name: "RejectsWhenNothingToBill",
			setup: func(svc *PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, bills *BillEngineMock) {
				order := draftOrder()
				order.State = PurchaseOrderStateConfirmed
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				line := purchaseLine(5)
				line.QtyReceived = 5
				line.QtyBilled = 5
				lines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*PurchaseOrderLine, error) {
					return []*PurchaseOrderLine{line}, nil
				}
			},
			expectErr: true,
			errTarget: ErrPurchaseOrderNothingToBill,
		},
		{
			name: "PropagatesResolverError",
			setup: func(svc *PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, bills *BillEngineMock) {
				order := draftOrder()
				order.State = PurchaseOrderStateConfirmed
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				line := purchaseLine(5)
				line.QtyReceived = 5
				lines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*PurchaseOrderLine, error) {
					return []*PurchaseOrderLine{line}, nil
				}
				svc.resolver = inventory.ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
						return inventory.ResolvedItem{}, errors.New("db down")
					},
				}
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orders, lines, _, _, _, bills, _, _ := testPurchaseOrderService()
			tt.setup(&svc, orders, lines, bills)
			_, err := svc.CreateSupplierBill(context.Background(), 1, 90, time.Now(), false)
			helper.AssertError(t, err, tt.expectErr, tt.errTarget)
		})
	}
}

func TestReceive(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(svc *PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, approvals *ApprovalEngineMock)
		expectErr bool
		errTarget error
	}{
		{
			name: "RejectsNonConfirmableState",
			setup: func(svc *PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, approvals *ApprovalEngineMock) {
				order := draftOrder()
				order.State = PurchaseOrderStateDraft
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
			},
			expectErr: true,
			errTarget: ErrPurchaseOrderState,
		},
		{
			name: "RejectsWithoutLines",
			setup: func(svc *PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, approvals *ApprovalEngineMock) {
				order := draftOrder()
				order.State = PurchaseOrderStateSent
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				lines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*PurchaseOrderLine, error) {
					return []*PurchaseOrderLine{}, nil
				}
			},
			expectErr: true,
			errTarget: ErrPurchaseOrderNoLines,
		},
		{
			name: "BlocksUnapprovedAboveThreshold",
			setup: func(svc *PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, approvals *ApprovalEngineMock) {
				order := draftOrder()
				order.State = PurchaseOrderStateSent
				order.AmountTotal = 1000000
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				svc.configs = NewSystemConfigSource(ConfigLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
						return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
							{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Key: configApprovalThreshold, Value: []byte(`1000`)},
						}}, nil
					},
				})
				approvals.IsApprovedFunc = func(_ context.Context, _ string, _ uint64) (bool, error) { return false, nil }
			},
			expectErr: true,
			errTarget: ErrPurchaseOrderApprovalPending,
		},
		{
			name: "RejectsWhenOrganizationMissing",
			setup: func(svc *PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, approvals *ApprovalEngineMock) {
				order := draftOrder()
				order.State = PurchaseOrderStateSent
				order.OrganizationID = nil
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
			},
			expectErr: true,
			errTarget: ErrPurchaseOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orders, lines, approvals, _, _, _, _, _ := testPurchaseOrderService()
			tt.setup(&svc, orders, lines, approvals)
			_, err := svc.Receive(context.Background(), 1, 90, time.Now())
			helper.AssertError(t, err, tt.expectErr, tt.errTarget)
		})
	}
}

func TestPaySupplierBill(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(svc *PurchaseOrderService, orders *PurchaseOrderDAOMock, openInvoices *OpenInvoiceLookupMock)
		expectErr bool
		errTarget error
	}{
		{
			name: "PropagatesOpenInvoiceError",
			setup: func(svc *PurchaseOrderService, orders *PurchaseOrderDAOMock, openInvoices *OpenInvoiceLookupMock) {
				order := draftOrder()
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				openInvoices.ListOpenByContactFunc = func(_ context.Context, _ uint64) ([]*accounting.Invoice, error) {
					return nil, errors.New("db down")
				}
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orders, _, _, _, _, _, openInvoices, _ := testPurchaseOrderService()
			tt.setup(&svc, orders, openInvoices)
			_, err := svc.PaySupplierBill(context.Background(), 1, 90, time.Now())
			helper.AssertError(t, err, tt.expectErr, tt.errTarget)
		})
	}
}

func TestFindOrCreateShipment(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(order *PurchaseOrder)
		expectErr bool
		errTarget error
	}{
		{
			name: "RejectsWithoutName",
			setup: func(order *PurchaseOrder) {
				order.Name = nil
			},
			expectErr: true,
			errTarget: ErrPurchaseOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
			order := draftOrder()
			tt.setup(order)
			_, err := svc.findOrCreateShipment(context.Background(), order)
			helper.AssertError(t, err, tt.expectErr, tt.errTarget)
		})
	}
}

func TestCreateFromRequest(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(svc *PurchaseOrderService)
		expectErr bool
		errTarget error
	}{
		{
			name: "PropagatesFindError",
			setup: func(svc *PurchaseOrderService) {
				svc.requisitions = &PurchaseRequestDAOMock{
					CRUDMock: dao.CRUDMock[PurchaseRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) {
							return nil, errors.New("db down")
						},
					},
				}
			},
			expectErr: true,
		},
		{
			name: "RejectsMissingRequisition",
			setup: func(svc *PurchaseOrderService) {
				svc.requisitions = &PurchaseRequestDAOMock{
					CRUDMock: dao.CRUDMock[PurchaseRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) {
							return nil, nil
						},
					},
				}
			},
			expectErr: true,
			errTarget: ErrRequisitionNotFound,
		},
		{
			name: "RejectsWithoutLines",
			setup: func(svc *PurchaseOrderService) {
				requisition := &PurchaseRequest{Base: model.Base{ID: 1}, State: RequestStateApproved}
				svc.requisitions = &PurchaseRequestDAOMock{
					CRUDMock: dao.CRUDMock[PurchaseRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) { return requisition, nil },
					},
				}
				svc.linesDAO = &PurchaseRequestLineDAOMock{
					ListByRequestFunc: func(_ context.Context, _ uint64) ([]*PurchaseRequestLine, error) {
						return []*PurchaseRequestLine{}, nil
					},
				}
			},
			expectErr: true,
			errTarget: ErrRequisitionNoLines,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
			tt.setup(&svc)
			_, err := svc.CreateFromRequest(context.Background(), 1, &PurchaseOrder{})
			helper.AssertError(t, err, tt.expectErr, tt.errTarget)
		})
	}
}
