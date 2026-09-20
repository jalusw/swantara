package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

func memoTestSvc(orders procurement.PurchaseOrderDAOMock, bills procurement.BillEngineMock) procurement.PurchaseOrderService {
	return procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders: orders,
		Bills:  bills,
	})
}

func confirmedOrder() *procurement.PurchaseOrder {
	order := samplePurchaseOrder()
	order.State = procurement.PurchaseOrderStateConfirmed
	return order
}

func TestPurchaseOrderHandler_CreateCreditMemo(t *testing.T) {
	validBody := `{"journal_id":5,"amount":500,"reason":"return"}`

	t.Run("creates credit memo", func(t *testing.T) {
		orders := procurement.PurchaseOrderDAOMock{
			CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
					return confirmedOrder(), nil
				},
			},
		}
		svc := memoTestSvc(orders, procurement.BillEngineMock{})
		app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/credit-memo", validBody)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)
	})

	t.Run("rejects invalid id body and date", func(t *testing.T) {
		svc := memoTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.BillEngineMock{})
		app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, svc)

		resp, err := doRequest(app, http.MethodPost, "/purchase-orders/abc/credit-memo", validBody)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/purchase-orders/1/credit-memo", `{"journal_id":0}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/purchase-orders/1/credit-memo", `{"journal_id":5,"amount":500,"reason":"return","date":"yesterday"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("maps service errors", func(t *testing.T) {
		cases := []struct {
			name   string
			find   func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error)
			bills  procurement.BillEngineMock
			status int
		}{
			{
				name:   "order not found",
				find:   func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) { return nil, nil },
				status: http.StatusNotFound,
			},
			{
				name: "wrong state",
				find: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
					return samplePurchaseOrder(), nil
				},
				status: http.StatusConflict,
			},
			{
				name: "bill failure",
				find: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
					return confirmedOrder(), nil
				},
				bills: procurement.BillEngineMock{
					CreateCreditNoteFunc: func(_ context.Context, _ accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
						return nil, errors.New("bill down")
					},
				},
				status: http.StatusInternalServerError,
			},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				orders := procurement.PurchaseOrderDAOMock{
					CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: tt.find},
				}
				svc := memoTestSvc(orders, tt.bills)
				app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, svc)
				resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/credit-memo", validBody)
				if err != nil {
					t.Fatal(err)
				}
				helper.AssertStatus(t, resp.StatusCode, tt.status)
			})
		}
	})
}

func TestPurchaseOrderHandler_CreateDebitMemo(t *testing.T) {
	validBody := `{"journal_id":5,"amount":500,"reason":"shortage"}`

	t.Run("creates debit memo", func(t *testing.T) {
		orders := procurement.PurchaseOrderDAOMock{
			CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
					return confirmedOrder(), nil
				},
			},
		}
		svc := memoTestSvc(orders, procurement.BillEngineMock{})
		app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/debit-memo", validBody)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)
	})

	t.Run("rejects invalid id body and date", func(t *testing.T) {
		svc := memoTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.BillEngineMock{})
		app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, svc)

		resp, err := doRequest(app, http.MethodPost, "/purchase-orders/abc/debit-memo", validBody)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/purchase-orders/1/debit-memo", `{"reason":"x"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/purchase-orders/1/debit-memo", `{"journal_id":5,"amount":500,"reason":"x","date":"yesterday"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("maps service errors", func(t *testing.T) {
		missing := procurement.PurchaseOrderDAOMock{
			CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) { return nil, nil },
			},
		}
		svc := memoTestSvc(missing, procurement.BillEngineMock{})
		app := orderTestApp(t, missing, procurement.PurchaseOrderLineDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/debit-memo", validBody)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		draft := procurement.PurchaseOrderDAOMock{
			CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
					return samplePurchaseOrder(), nil
				},
			},
		}
		svc = memoTestSvc(draft, procurement.BillEngineMock{})
		app = orderTestApp(t, draft, procurement.PurchaseOrderLineDAOMock{}, svc)
		resp, err = doRequest(app, http.MethodPost, "/purchase-orders/1/debit-memo", validBody)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusConflict)

		billed := procurement.PurchaseOrderDAOMock{
			CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
					return confirmedOrder(), nil
				},
			},
		}
		bills := procurement.BillEngineMock{
			CreateSupplierBillFunc: func(_ context.Context, _ accounting.CreateSupplierBillRequest) (*accounting.Invoice, error) {
				return nil, errors.New("bill down")
			},
		}
		svc = memoTestSvc(billed, bills)
		app = orderTestApp(t, billed, procurement.PurchaseOrderLineDAOMock{}, svc)
		resp, err = doRequest(app, http.MethodPost, "/purchase-orders/1/debit-memo", validBody)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}
