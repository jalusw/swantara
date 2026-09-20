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
)

func confirmedPOMock(state string) *PurchaseOrderDAOMock {
	return &PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) {
				return &PurchaseOrder{
					Base:           model.Base{ID: 1},
					OrganizationID: helper.Ptr(uint64(10)),
					SupplierID:     20,
					State:          state,
				}, nil
			},
		},
	}
}

func TestPurchaseOrderService_VendorMemos(t *testing.T) {
	ctx := context.Background()

	t.Run("credit memo success", func(t *testing.T) {
		svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
		svc.orders = confirmedPOMock(PurchaseOrderStateConfirmed)
		memo, err := svc.CreateVendorCreditMemo(ctx, 1, 5, time.Now(), 500, "return")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if memo.OrderID != 1 || memo.AmountTotal != 500 {
			t.Errorf("memo = %+v, want order 1 amount 500", memo)
		}
	})

	t.Run("credit memo failures", func(t *testing.T) {
		dbErr := errors.New("db down")
		cases := []struct {
			name   string
			find   func(_ context.Context, _ uint64) (*PurchaseOrder, error)
			billEr error
			wantEr error
		}{
			{name: "order missing", find: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return nil, nil }, wantEr: ErrPurchaseOrderNotFound},
			{name: "lookup error", find: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return nil, dbErr }, wantEr: dbErr},
			{name: "no org", find: func(_ context.Context, _ uint64) (*PurchaseOrder, error) {
				return &PurchaseOrder{Base: model.Base{ID: 1}, State: PurchaseOrderStateConfirmed}, nil
			}, wantEr: ErrPurchaseOrderNotFound},
			{name: "wrong state", find: func(_ context.Context, _ uint64) (*PurchaseOrder, error) {
				return &PurchaseOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: PurchaseOrderStateDraft}, nil
			}, wantEr: ErrPurchaseOrderState},
			{name: "bill error", find: func(_ context.Context, _ uint64) (*PurchaseOrder, error) {
				return &PurchaseOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), SupplierID: 20, State: PurchaseOrderStateDone}, nil
			}, billEr: dbErr, wantEr: dbErr},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc, _, _, _, _, _, bills, _, _ := testPurchaseOrderService()
				svc.orders = &PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[PurchaseOrder]{FindFunc: tt.find}}
				if tt.billEr != nil {
					bills.CreateCreditNoteFunc = func(_ context.Context, _ accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
						return nil, tt.billEr
					}
				}
				_, err := svc.CreateVendorCreditMemo(ctx, 1, 5, time.Now(), 500, "return")
				helper.AssertError(t, err, true, tt.wantEr)
			})
		}
	})
}
