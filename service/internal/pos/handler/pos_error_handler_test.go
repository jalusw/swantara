package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/pos"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func writeErrorStatus(path string, fn func(c fiber.Ctx) error) int {
	app := fiber.New()
	app.Post(path, func(c fiber.Ctx) error {
		return fn(c)
	})
	resp, err := doRequest(app, http.MethodPost, path, "")
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

func TestWritePOSError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "config not found", err: pos.ErrConfigNotFound, status: http.StatusUnprocessableEntity},
		{name: "session not found", err: pos.ErrSessionNotFound, status: http.StatusNotFound},
		{name: "session state", err: pos.ErrSessionState, status: http.StatusConflict},
		{name: "session open", err: pos.ErrSessionOpen, status: http.StatusConflict},
		{name: "session reconciliation", err: pos.ErrSessionReconciliation, status: http.StatusConflict},
		{name: "order not found", err: pos.ErrOrderNotFound, status: http.StatusNotFound},
		{name: "order no lines", err: pos.ErrOrderNoLines, status: http.StatusUnprocessableEntity},
		{name: "order item", err: pos.ErrOrderProduct, status: http.StatusUnprocessableEntity},
		{name: "order qty", err: pos.ErrOrderQty, status: http.StatusUnprocessableEntity},
		{name: "order discount", err: pos.ErrOrderDiscount, status: http.StatusUnprocessableEntity},
		{name: "order tax", err: pos.ErrOrderTax, status: http.StatusUnprocessableEntity},
		{name: "order stock unavailable", err: pos.ErrOrderStockUnavailable, status: http.StatusConflict},
		{name: "order no stock", err: pos.ErrOrderNoStock, status: http.StatusUnprocessableEntity},
		{name: "payment missing", err: pos.ErrPaymentMissing, status: http.StatusUnprocessableEntity},
		{name: "payment total", err: pos.ErrPaymentTotal, status: http.StatusUnprocessableEntity},
		{name: "order invoiced", err: pos.ErrOrderInvoiced, status: http.StatusConflict},
		{name: "config warehouse", err: pos.ErrConfigWarehouse, status: http.StatusUnprocessableEntity},
		{name: "config journal", err: pos.ErrConfigJournal, status: http.StatusUnprocessableEntity},
		{name: "config price_book", err: pos.ErrConfigPriceBook, status: http.StatusUnprocessableEntity},
		{name: "no cashier", err: pos.ErrNoCashier, status: http.StatusUnprocessableEntity},
		{name: "order state", err: pos.ErrOrderState, status: http.StatusConflict},
		{name: "order cost", err: pos.ErrOrderCost, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writePOSError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}

func TestPOSConfigHandler_List_ErrorPaths(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		h := NewPOSConfigHandler(pos.NewPOSConfigService(dao.CRUDMock[reference.POSConfig]{}))
		app := fiber.New()
		h.Register(app, tenantNoOrganizationGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/configs", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		h := NewPOSConfigHandler(pos.NewPOSConfigService(dao.CRUDMock[reference.POSConfig]{}))
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/configs?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		h := NewPOSConfigHandler(pos.NewPOSConfigService(dao.CRUDMock[reference.POSConfig]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSConfig], error) {
				return nil, errors.New("boom")
			},
		}))
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/configs", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSConfigHandler_Create_ErrorPaths(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		h := NewPOSConfigHandler(pos.NewPOSConfigService(dao.CRUDMock[reference.POSConfig]{}))
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/configs", `{"name":"Counter"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		h := NewPOSConfigHandler(pos.NewPOSConfigService(dao.CRUDMock[reference.POSConfig]{
			CreateFunc: func(_ context.Context, _ *reference.POSConfig) (*reference.POSConfig, error) {
				return nil, errors.New("boom")
			},
		}))
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/configs", `{"name":"Counter A","warehouse_id":4,"journal_id":2,"price_book_id":3}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSConfigHandler_Get_ErrorPaths(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		h := NewPOSConfigHandler(pos.NewPOSConfigService(dao.CRUDMock[reference.POSConfig]{}))
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/configs/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := NewPOSConfigHandler(pos.NewPOSConfigService(dao.CRUDMock[reference.POSConfig]{}))
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/configs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		h := NewPOSConfigHandler(pos.NewPOSConfigService(dao.CRUDMock[reference.POSConfig]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
				return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(99))}, nil
			},
		}))
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/configs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		h := NewPOSConfigHandler(pos.NewPOSConfigService(dao.CRUDMock[reference.POSConfig]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
				return nil, errors.New("boom")
			},
		}))
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/configs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSSessionHandler_List_ErrorPaths(t *testing.T) {
	t.Run("invalid query", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/sessions?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{
			ListInOrganizationFunc: func(_ context.Context, _ *query.Query, _ uint64) (*query.Page[pos.POSSession], error) {
				return nil, errors.New("boom")
			},
		}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/sessions", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSSessionHandler_Get_ErrorPaths(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/sessions/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/sessions/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("breakdown error", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
				return &pos.POSSession{Base: model.Base{ID: 1}, ConfigID: 1, State: pos.SessionStateOpened}, nil
			},
		}}
		configs := dao.CRUDMock[reference.POSConfig]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
				return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1))}, nil
			},
		}
		payments := pos.POSPaymentDAOMock{
			SumBySessionFunc: func(_ context.Context, _ uint64) ([]pos.PaymentMethodTotal, error) {
				return nil, errors.New("boom")
			},
		}
		svc := testPOSHandlerService(configs, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, payments)
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/sessions/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSSessionHandler_Open_ErrorPaths(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/sessions", `{"cashier_id":7}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		configs := dao.CRUDMock[reference.POSConfig]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
				return nil, errors.New("boom")
			},
		}
		svc := testPOSHandlerService(configs, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/sessions", `{"config_id":1,"cashier_id":7,"opening_balance":100}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSSessionHandler_StartClosing(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
				return &pos.POSSession{Base: model.Base{ID: 1}, ConfigID: 1, State: pos.SessionStateOpened}, nil
			},
			UpdateFunc: func(_ context.Context, s *pos.POSSession) (*pos.POSSession, error) { return s, nil },
		}}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/sessions/1/closing", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/sessions/abc/closing", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/sessions/1/closing", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("wrong state", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
				return &pos.POSSession{Base: model.Base{ID: 1}, State: pos.SessionStateClosed}, nil
			},
		}}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/sessions/1/closing", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
				return nil, errors.New("boom")
			},
		}}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/sessions/1/closing", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSSessionHandler_Close_ErrorPaths(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/sessions/abc/close", `{"closing_balance":250}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/sessions/1/close", `{"closing_balance":-5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/sessions/1/close", `{"closing_balance":250}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
				return nil, errors.New("boom")
			},
		}}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSSessionHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/sessions/1/close", `{"closing_balance":250}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSOrderHandler_List_ErrorPaths(t *testing.T) {
	t.Run("invalid query", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/orders?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		orders := pos.POSOrderDAOMock{
			ListInOrganizationFunc: func(_ context.Context, _ *query.Query, _ uint64) (*query.Page[pos.POSOrder], error) {
				return nil, errors.New("boom")
			},
		}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, orders, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/orders", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSOrderHandler_Get_ErrorPaths(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/orders/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/orders/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("lines error", func(t *testing.T) {
		organizationID := uint64(1)
		config := reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: &organizationID}
		sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
				return &pos.POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
			},
		}}
		configs := dao.CRUDMock[reference.POSConfig]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
		}
		orders := pos.POSOrderDAOMock{CRUDMock: dao.CRUDMock[pos.POSOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSOrder, error) {
				return &pos.POSOrder{Base: model.Base{ID: 10}, SessionID: 1, State: pos.OrderStateDone}, nil
			},
		}}
		lines := pos.POSOrderLineDAOMock{
			ListByOrderFunc: func(_ context.Context, _ uint64) ([]*pos.POSOrderLine, error) {
				return nil, errors.New("boom")
			},
		}
		svc := testPOSHandlerService(configs, sessions, orders, lines, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/orders/10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("payments error", func(t *testing.T) {
		organizationID := uint64(1)
		config := reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: &organizationID}
		sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
				return &pos.POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
			},
		}}
		configs := dao.CRUDMock[reference.POSConfig]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
		}
		orders := pos.POSOrderDAOMock{CRUDMock: dao.CRUDMock[pos.POSOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSOrder, error) {
				return &pos.POSOrder{Base: model.Base{ID: 10}, SessionID: 1, State: pos.OrderStateDone}, nil
			},
		}}
		lines := pos.POSOrderLineDAOMock{}
		payments := pos.POSPaymentDAOMock{
			ListByOrderFunc: func(_ context.Context, _ uint64) ([]*pos.POSPayment, error) {
				return nil, errors.New("boom")
			},
		}
		svc := testPOSHandlerService(configs, sessions, orders, lines, payments)
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodGet, "/pos/orders/10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSOrderHandler_Invoice_ErrorPaths(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders/abc/invoice", `{"journal_id":2,"date":"2026-08-14"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders/1/invoice", `{"journal_id":2}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders/1/invoice", `{"journal_id":2,"date":"bogus"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("order not found", func(t *testing.T) {
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders/1/invoice", `{"journal_id":2,"date":"2026-08-14"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		orders := pos.POSOrderDAOMock{CRUDMock: dao.CRUDMock[pos.POSOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSOrder, error) {
				return nil, errors.New("boom")
			},
		}}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, orders, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders/1/invoice", `{"journal_id":2,"date":"2026-08-14"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSOrderHandler_Refund_ErrorPaths(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders/abc/refund", `{"journal_id":2,"date":"2026-08-14"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders/1/refund", `{"journal_id":2}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders/1/refund", `{"journal_id":2,"date":"bogus"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("order not found", func(t *testing.T) {
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders/1/refund", `{"journal_id":2,"date":"2026-08-14"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		orders := pos.POSOrderDAOMock{CRUDMock: dao.CRUDMock[pos.POSOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSOrder, error) {
				return nil, errors.New("boom")
			},
		}}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, orders, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders/1/refund", `{"journal_id":2,"date":"2026-08-14"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPOSOrderHandler_Sell_ErrorPaths(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, pos.POSSessionDAOMock{}, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders", `{"session_id":1}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("session not open", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
				return &pos.POSSession{Base: model.Base{ID: 1}, State: pos.SessionStateClosed}, nil
			},
		}}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders", `{"session_id":1,"lines":[{"item_id":100,"qty":2}],"payments":[{"method":"cash","amount":220}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
			FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
				return nil, errors.New("boom")
			},
		}}
		svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
		h := NewPOSOrderHandler(svc)
		app := fiber.New()
		h.Register(app, passthroughGuards())

		resp, err := doRequest(app, http.MethodPost, "/pos/orders", `{"session_id":1,"lines":[{"item_id":100,"qty":2}],"payments":[{"method":"cash","amount":220}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func tenantNoOrganizationGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			return c.Next()
		},
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error {
				return c.Next()
			}
		},
	}
}
