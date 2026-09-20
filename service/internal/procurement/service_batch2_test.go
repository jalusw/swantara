package procurement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestPurchaseOrderService_VendorDebitMemo(t *testing.T) {
	ctx := context.Background()

	t.Run("debit memo success", func(t *testing.T) {
		svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
		svc.orders = confirmedPOMock(PurchaseOrderStateDone)
		memo, err := svc.CreateVendorDebitMemo(ctx, 1, 5, time.Now(), 250, "shortage")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if memo.OrderID != 1 || memo.AmountTotal != 250 {
			t.Errorf("memo = %+v, want order 1 amount 250", memo)
		}
	})

	t.Run("debit memo failures", func(t *testing.T) {
		dbErr := errors.New("db down")
		cases := []struct {
			name   string
			find   func(_ context.Context, _ uint64) (*PurchaseOrder, error)
			billEr error
			wantEr error
		}{
			{name: "order missing", find: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return nil, nil }, wantEr: ErrPurchaseOrderNotFound},
			{name: "wrong state", find: func(_ context.Context, _ uint64) (*PurchaseOrder, error) {
				return &PurchaseOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: PurchaseOrderStateDraft}, nil
			}, wantEr: ErrPurchaseOrderState},
			{name: "bill error", find: func(_ context.Context, _ uint64) (*PurchaseOrder, error) {
				return &PurchaseOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), SupplierID: 20, State: PurchaseOrderStateConfirmed}, nil
			}, billEr: dbErr, wantEr: dbErr},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc, _, _, _, _, _, bills, _, _ := testPurchaseOrderService()
				svc.orders = &PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[PurchaseOrder]{FindFunc: tt.find}}
				if tt.billEr != nil {
					bills.CreateSupplierBillFunc = func(_ context.Context, _ accounting.CreateSupplierBillRequest) (*accounting.Invoice, error) {
						return nil, tt.billEr
					}
				}
				_, err := svc.CreateVendorDebitMemo(ctx, 1, 5, time.Now(), 250, "shortage")
				helper.AssertError(t, err, true, tt.wantEr)
			})
		}
	})
}

func batchOrderMock(supplierID uint64, state string) *PurchaseOrderDAOMock {
	return &PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) {
				return &PurchaseOrder{
					Base:           model.Base{ID: 1},
					OrganizationID: helper.Ptr(uint64(10)),
					SupplierID:     supplierID,
					State:          state,
				}, nil
			},
		},
	}
}

func TestPurchaseOrderService_PaymentBatchFlow(t *testing.T) {
	ctx := context.Background()

	t.Run("creates batch", func(t *testing.T) {
		svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
		svc.orders = batchOrderMock(20, PurchaseOrderStateConfirmed)
		batch, lines, err := svc.CreatePaymentBatch(ctx, 10, 20, PaymentBatchRequest{
			JournalID: 3,
			Date:      time.Now(),
			Orders:    []PaymentBatchRequestLine{{OrderID: 1, Amount: 5000}},
		})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if batch.TotalAmount != 5000 || len(lines) != 1 {
			t.Errorf("batch = %+v lines = %d, want total 5000 one line", batch, len(lines))
		}
	})

	t.Run("create failures", func(t *testing.T) {
		dbErr := errors.New("db down")
		cases := []struct {
			name       string
			supplierID uint64
			state      string
			orders     []PaymentBatchRequestLine
			findErr    error
			wantEr     error
		}{
			{name: "no orders", orders: nil, wantEr: ErrPaymentBatchNoOrders},
			{name: "mixed vendors", supplierID: 21, state: PurchaseOrderStateConfirmed, orders: []PaymentBatchRequestLine{{OrderID: 1, Amount: 100}}, wantEr: ErrPaymentBatchMixedVendors},
			{name: "bad state", supplierID: 20, state: PurchaseOrderStateDraft, orders: []PaymentBatchRequestLine{{OrderID: 1, Amount: 100}}, wantEr: ErrPaymentBatchInvalidOrderState},
			{name: "bad amount", supplierID: 20, state: PurchaseOrderStateConfirmed, orders: []PaymentBatchRequestLine{{OrderID: 1, Amount: 0}}, wantEr: ErrPaymentBatchInvalidAmount},
			{name: "lookup error", findErr: dbErr, orders: []PaymentBatchRequestLine{{OrderID: 1, Amount: 100}}, wantEr: dbErr},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				if tt.findErr != nil {
					svc.orders = &PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[PurchaseOrder]{
						FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return nil, tt.findErr },
					}}
				} else if tt.orders != nil {
					svc.orders = batchOrderMock(tt.supplierID, tt.state)
				}
				_, _, err := svc.CreatePaymentBatch(ctx, 10, 20, PaymentBatchRequest{
					JournalID: 3,
					Date:      time.Now(),
					Orders:    tt.orders,
				})
				helper.AssertError(t, err, true, tt.wantEr)
			})
		}
	})

	t.Run("confirms batch", func(t *testing.T) {
		svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
		svc.orders = batchOrderMock(20, PurchaseOrderStateConfirmed)
		svc.batches = &PaymentBatchDAOMock{
			CRUDMock: dao.CRUDMock[PaymentBatch]{
				FindFunc: func(_ context.Context, _ uint64) (*PaymentBatch, error) {
					return &PaymentBatch{Base: model.Base{ID: 1}, State: PaymentBatchStateDraft}, nil
				},
			},
		}
		svc.batchLines = &PaymentBatchLineDAOMock{
			ListByBatchFunc: func(_ context.Context, _ uint64) ([]*PaymentBatchLine, error) {
				return []*PaymentBatchLine{{Base: model.Base{ID: 1}, BatchID: 1, OrderID: 1, Amount: 5000}}, nil
			},
		}
		got, err := svc.ConfirmBatch(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != PaymentBatchStatePosted {
			t.Errorf("state = %q, want posted", got.State)
		}
	})

	t.Run("confirm failures", func(t *testing.T) {
		dbErr := errors.New("db down")
		cases := []struct {
			name    string
			batch   *PaymentBatch
			findErr error
			wantEr  error
		}{
			{name: "missing", batch: nil, wantEr: ErrPaymentBatchNotFound},
			{name: "lookup error", findErr: dbErr, wantEr: dbErr},
			{name: "bad state", batch: &PaymentBatch{Base: model.Base{ID: 1}, State: PaymentBatchStatePosted}, wantEr: ErrPaymentBatchInvalidState},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
				svc.batches = &PaymentBatchDAOMock{
					CRUDMock: dao.CRUDMock[PaymentBatch]{
						FindFunc: func(_ context.Context, _ uint64) (*PaymentBatch, error) { return tt.batch, tt.findErr },
					},
				}
				_, err := svc.ConfirmBatch(ctx, 1)
				helper.AssertError(t, err, true, tt.wantEr)
			})
		}
	})

	t.Run("gets batch and lists", func(t *testing.T) {
		svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
		svc.batches = &PaymentBatchDAOMock{
			CRUDMock: dao.CRUDMock[PaymentBatch]{
				FindFunc: func(_ context.Context, _ uint64) (*PaymentBatch, error) {
					return &PaymentBatch{Base: model.Base{ID: 1}, State: PaymentBatchStateDraft}, nil
				},
			},
		}
		batch, lines, err := svc.GetPaymentBatch(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if batch.ID != 1 || len(lines) != 0 {
			t.Errorf("batch = %+v lines = %d", batch, len(lines))
		}

		svc.batches = &PaymentBatchDAOMock{
			CRUDMock: dao.CRUDMock[PaymentBatch]{
				FindFunc: func(_ context.Context, _ uint64) (*PaymentBatch, error) { return nil, nil },
			},
		}
		_, _, err = svc.GetPaymentBatch(ctx, 2)
		helper.AssertError(t, err, true, ErrPaymentBatchNotFound)

		svc.batches = &PaymentBatchDAOMock{
			CRUDMock: dao.CRUDMock[PaymentBatch]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[PaymentBatch], error) {
					return &query.Page[PaymentBatch]{Items: []*PaymentBatch{{Base: model.Base{ID: 1}}}, Count: 1}, nil
				},
			},
		}
		page, err := svc.ListPaymentBatches(ctx, &query.Query{})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if page.Count != 1 {
			t.Errorf("count = %d, want 1", page.Count)
		}
	})
}

func agreementForOrder(state string, endDate *time.Time, qtyLimit, amountLimit float64) *SupplyAgreement {
	return &SupplyAgreement{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(10)),
		SupplierID:     20,
		State:          state,
		EndDate:        endDate,
		QtyLimit:       qtyLimit,
		AmountLimit:    amountLimit,
	}
}

func TestPurchaseOrderService_CreateFromAgreement(t *testing.T) {
	ctx := context.Background()
	future := time.Now().AddDate(0, 3, 0)
	past := time.Now().AddDate(0, -1, 0)

	newSvc := func(agreement *SupplyAgreement, findErr error, lines []*SupplyAgreementLine) PurchaseOrderService {
		svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
		svc.agreements = &SupplyAgreementDAOMock{
			CRUDMock: dao.CRUDMock[SupplyAgreement]{
				FindFunc: func(_ context.Context, _ uint64) (*SupplyAgreement, error) { return agreement, findErr },
			},
		}
		svc.agreementLines = &SupplyAgreementLineDAOMock{
			ListByAgreementFunc: func(_ context.Context, _ uint64) ([]*SupplyAgreementLine, error) { return lines, nil },
		}
		return svc
	}
	twoLines := []*SupplyAgreementLine{
		{Base: model.Base{ID: 1}, Qty: 2, UnitPrice: 1000},
		{Base: model.Base{ID: 2}, Qty: 1, UnitPrice: 500},
	}

	t.Run("creates order with overrides", func(t *testing.T) {
		svc := newSvc(agreementForOrder(SupplyAgreementStateActive, &future, 0, 0), nil, twoLines)
		order, err := svc.CreateFromAgreement(ctx, 1, &CreateFromAgreementOverrides{
			SupplierID:   helper.Ptr(uint64(21)),
			WarehouseID:  helper.Ptr(uint64(9)),
			QtyOverrides: map[uint64]float64{1: 5},
		})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if order.SupplierID != 21 {
			t.Errorf("supplier = %d, want override 21", order.SupplierID)
		}
	})

	t.Run("creates order without overrides", func(t *testing.T) {
		svc := newSvc(agreementForOrder(SupplyAgreementStateActive, nil, 0, 0), nil, twoLines[:1])
		order, err := svc.CreateFromAgreement(ctx, 1, nil)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if order.State != PurchaseOrderStateDraft {
			t.Errorf("state = %q, want draft", order.State)
		}
	})

	t.Run("failures", func(t *testing.T) {
		dbErr := errors.New("db down")
		cases := []struct {
			name      string
			agreement *SupplyAgreement
			findErr   error
			lines     []*SupplyAgreementLine
			wantEr    error
		}{
			{name: "lookup error", findErr: dbErr, wantEr: dbErr},
			{name: "missing", agreement: nil, wantEr: ErrAgreementNotFound},
			{name: "not active", agreement: agreementForOrder(SupplyAgreementStateDraft, &future, 0, 0), lines: twoLines, wantEr: ErrAgreementNotActive},
			{name: "expired", agreement: agreementForOrder(SupplyAgreementStateActive, &past, 0, 0), lines: twoLines, wantEr: ErrAgreementExpired},
			{name: "no lines", agreement: agreementForOrder(SupplyAgreementStateActive, &future, 0, 0), wantEr: ErrAgreementNoLines},
			{name: "all zero qty", agreement: agreementForOrder(SupplyAgreementStateActive, &future, 0, 0), lines: []*SupplyAgreementLine{{Base: model.Base{ID: 1}, Qty: 0}}, wantEr: ErrPurchaseOrderNothingToBill},
			{name: "qty exceeded", agreement: agreementForOrder(SupplyAgreementStateActive, &future, 1, 0), lines: twoLines, wantEr: ErrAgreementQtyExceeded},
			{name: "amount exceeded", agreement: agreementForOrder(SupplyAgreementStateActive, &future, 0, 100), lines: twoLines, wantEr: ErrAgreementAmountExceeded},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := newSvc(tt.agreement, tt.findErr, tt.lines)
				_, err := svc.CreateFromAgreement(ctx, 1, nil)
				helper.AssertError(t, err, true, tt.wantEr)
			})
		}
	})
}
