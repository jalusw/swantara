package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type stubPaymentBatchService struct {
	listFunc    func(ctx context.Context, q *query.Query) (*query.Page[procurement.PaymentBatch], error)
	getFunc     func(ctx context.Context, id uint64) (*procurement.PaymentBatch, []*procurement.PaymentBatchLine, error)
	createFunc  func(ctx context.Context, orgID, supplierID uint64, req procurement.PaymentBatchRequest) (*procurement.PaymentBatch, []*procurement.PaymentBatchLine, error)
	confirmFunc func(ctx context.Context, id uint64) (*procurement.PaymentBatch, error)
}

func (m stubPaymentBatchService) ListPaymentBatches(ctx context.Context, q *query.Query) (*query.Page[procurement.PaymentBatch], error) {
	return m.listFunc(ctx, q)
}

func (m stubPaymentBatchService) GetPaymentBatch(ctx context.Context, id uint64) (*procurement.PaymentBatch, []*procurement.PaymentBatchLine, error) {
	return m.getFunc(ctx, id)
}

func (m stubPaymentBatchService) CreatePaymentBatch(ctx context.Context, orgID, supplierID uint64, req procurement.PaymentBatchRequest) (*procurement.PaymentBatch, []*procurement.PaymentBatchLine, error) {
	return m.createFunc(ctx, orgID, supplierID, req)
}

func (m stubPaymentBatchService) ConfirmBatch(ctx context.Context, id uint64) (*procurement.PaymentBatch, error) {
	return m.confirmFunc(ctx, id)
}

func paymentBatchTestApp(t *testing.T, svc procurement.PaymentBatchService) *fiber.App {
	t.Helper()
	return procurementTestApp(t, func(api fiber.Router, guards httpx.RouteGuards) {
		NewPaymentBatchHandler(svc).Register(api, guards)
	})
}

func samplePaymentBatch() *procurement.PaymentBatch {
	return &procurement.PaymentBatch{
		Base:           model.Base{ID: 1},
		OrganizationID: ptrUint64(10),
		JournalID:      3,
		ContactID:      20,
		State:          "draft",
		TotalAmount:    5000,
		PaymentCount:   2,
	}
}

func samplePaymentBatchLines() []*procurement.PaymentBatchLine {
	return []*procurement.PaymentBatchLine{{Base: model.Base{ID: 1}, BatchID: 1, OrderID: 7, Amount: 5000}}
}

func okBatchService() stubPaymentBatchService {
	return stubPaymentBatchService{
		listFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.PaymentBatch], error) {
			return &query.Page[procurement.PaymentBatch]{Items: []*procurement.PaymentBatch{samplePaymentBatch()}, Count: 1}, nil
		},
		getFunc: func(_ context.Context, _ uint64) (*procurement.PaymentBatch, []*procurement.PaymentBatchLine, error) {
			return samplePaymentBatch(), samplePaymentBatchLines(), nil
		},
		createFunc: func(_ context.Context, _, _ uint64, _ procurement.PaymentBatchRequest) (*procurement.PaymentBatch, []*procurement.PaymentBatchLine, error) {
			return samplePaymentBatch(), samplePaymentBatchLines(), nil
		},
		confirmFunc: func(_ context.Context, _ uint64) (*procurement.PaymentBatch, error) {
			return samplePaymentBatch(), nil
		},
	}
}

func TestPaymentBatchHandler_List_Get(t *testing.T) {
	t.Run("lists batches", func(t *testing.T) {
		resp, err := doRequest(paymentBatchTestApp(t, okBatchService()), http.MethodGet, "/payment-batches/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects invalid query and missing tenant and reports failure", func(t *testing.T) {
		resp, err := doRequest(paymentBatchTestApp(t, okBatchService()), http.MethodGet, "/payment-batches/?size=abc", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		app := procurementTestAppNoTenant(t, func(api fiber.Router, guards httpx.RouteGuards) {
			NewPaymentBatchHandler(okBatchService()).Register(api, guards)
		})
		resp, err = doRequest(app, http.MethodGet, "/payment-batches/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)

		failing := okBatchService()
		failing.listFunc = func(_ context.Context, _ *query.Query) (*query.Page[procurement.PaymentBatch], error) {
			return nil, errors.New("db down")
		}
		resp, err = doRequest(paymentBatchTestApp(t, failing), http.MethodGet, "/payment-batches/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		resp, err := doRequest(paymentBatchTestApp(t, okBatchService()), http.MethodGet, "/payment-batches/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(paymentBatchTestApp(t, okBatchService()), http.MethodPost, "/payment-batches/abc/confirm", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("returns batch", func(t *testing.T) {
		resp, err := doRequest(paymentBatchTestApp(t, okBatchService()), http.MethodGet, "/payment-batches/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})
}

func TestPaymentBatchHandler_Write(t *testing.T) {
	validBody := `{"supplier_id":20,"journal_id":3,"orders":[{"order_id":7,"amount":5000}]}`

	t.Run("creates batch", func(t *testing.T) {
		resp, err := doRequest(paymentBatchTestApp(t, okBatchService()), http.MethodPost, "/payment-batches/", validBody)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)
	})

	t.Run("rejects invalid body and date and missing tenant", func(t *testing.T) {
		resp, err := doRequest(paymentBatchTestApp(t, okBatchService()), http.MethodPost, "/payment-batches/", `{"supplier_id":0}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(paymentBatchTestApp(t, okBatchService()), http.MethodPost, "/payment-batches/", `{"supplier_id":20,"journal_id":3,"date":"yesterday","orders":[{"order_id":7,"amount":5000}]}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		app := procurementTestAppNoTenant(t, func(api fiber.Router, guards httpx.RouteGuards) {
			NewPaymentBatchHandler(okBatchService()).Register(api, guards)
		})
		resp, err = doRequest(app, http.MethodPost, "/payment-batches/", validBody)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)
	})

	t.Run("confirms batch", func(t *testing.T) {
		resp, err := doRequest(paymentBatchTestApp(t, okBatchService()), http.MethodPost, "/payment-batches/1/confirm", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	batchErrors := []struct {
		name   string
		svcErr error
		status int
	}{
		{name: "not found", svcErr: procurement.ErrPaymentBatchNotFound, status: http.StatusNotFound},
		{name: "no orders", svcErr: procurement.ErrPaymentBatchNoOrders, status: http.StatusUnprocessableEntity},
		{name: "mixed vendors", svcErr: procurement.ErrPaymentBatchMixedVendors, status: http.StatusUnprocessableEntity},
		{name: "invalid order state", svcErr: procurement.ErrPaymentBatchInvalidOrderState, status: http.StatusUnprocessableEntity},
		{name: "invalid amount", svcErr: procurement.ErrPaymentBatchInvalidAmount, status: http.StatusUnprocessableEntity},
		{name: "invalid state", svcErr: procurement.ErrPaymentBatchInvalidState, status: http.StatusUnprocessableEntity},
		{name: "unexpected", svcErr: errors.New("db down"), status: http.StatusInternalServerError},
	}

	for _, tt := range batchErrors {
		t.Run("get maps "+tt.name, func(t *testing.T) {
			svc := okBatchService()
			svc.getFunc = func(_ context.Context, _ uint64) (*procurement.PaymentBatch, []*procurement.PaymentBatchLine, error) {
				return nil, nil, tt.svcErr
			}
			resp, err := doRequest(paymentBatchTestApp(t, svc), http.MethodGet, "/payment-batches/1", "")
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, tt.status)
		})

		t.Run("create maps "+tt.name, func(t *testing.T) {
			svc := okBatchService()
			svc.createFunc = func(_ context.Context, _, _ uint64, _ procurement.PaymentBatchRequest) (*procurement.PaymentBatch, []*procurement.PaymentBatchLine, error) {
				return nil, nil, tt.svcErr
			}
			resp, err := doRequest(paymentBatchTestApp(t, svc), http.MethodPost, "/payment-batches/", validBody)
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, tt.status)
		})

		t.Run("confirm maps "+tt.name, func(t *testing.T) {
			svc := okBatchService()
			svc.confirmFunc = func(_ context.Context, _ uint64) (*procurement.PaymentBatch, error) {
				return nil, tt.svcErr
			}
			resp, err := doRequest(paymentBatchTestApp(t, svc), http.MethodPost, "/payment-batches/1/confirm", "")
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, tt.status)
		})
	}
}
