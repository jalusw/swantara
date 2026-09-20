package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/pos"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func paymentAccountTestApp(t *testing.T, svc *pos.PaymentAccountService, withOrg bool) *fiber.App {
	t.Helper()
	return paymentAccountTestAppGuards(t, svc, withOrg, passthroughGuards())
}

func paymentAccountTestAppGuards(t *testing.T, svc *pos.PaymentAccountService, withOrg bool, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	if withOrg {
		app.Use(func(c fiber.Ctx) error {
			c.Locals(httpx.LocalOrganizationID, uint64(10))
			return c.Next()
		})
	}
	NewPaymentAccountHandler(svc).Register(app, guards)
	return app
}

func bareGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error { return c.Next() },
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error { return c.Next() }
		},
	}
}

func accountSvc(accounts dao.CRUDMock[reference.POSPaymentAccount]) *pos.PaymentAccountService {
	return pos.NewPaymentAccountService(accounts)
}

func TestPaymentAccountHandler_List(t *testing.T) {
	t.Run("lists accounts", func(t *testing.T) {
		svc := accountSvc(dao.CRUDMock[reference.POSPaymentAccount]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
				return &query.Page[reference.POSPaymentAccount]{Items: []*reference.POSPaymentAccount{{Base: model.Base{ID: 1}, OrganizationID: 10, Method: "cash", AccountID: 100}}, Count: 1}, nil
			},
		})
		resp, err := doRequest(paymentAccountTestApp(t, svc, true), http.MethodGet, "/pos/payment-accounts/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects missing org and reports failure", func(t *testing.T) {
		svc := accountSvc(dao.CRUDMock[reference.POSPaymentAccount]{})
		resp, err := doRequest(paymentAccountTestAppGuards(t, svc, false, bareGuards()), http.MethodGet, "/pos/payment-accounts/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)

		failing := accountSvc(dao.CRUDMock[reference.POSPaymentAccount]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
				return nil, errors.New("db down")
			},
		})
		resp, err = doRequest(paymentAccountTestApp(t, failing, true), http.MethodGet, "/pos/payment-accounts/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}

func TestPaymentAccountHandler_Upsert(t *testing.T) {
	t.Run("creates and updates", func(t *testing.T) {
		created := accountSvc(dao.CRUDMock[reference.POSPaymentAccount]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
				return &query.Page[reference.POSPaymentAccount]{}, nil
			},
			CreateFunc: func(_ context.Context, a *reference.POSPaymentAccount) (*reference.POSPaymentAccount, error) {
				a.ID = 2
				return a, nil
			},
		})
		resp, err := doRequest(paymentAccountTestApp(t, created, true), http.MethodPut, "/pos/payment-accounts/cash", `{"account_id":100}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)

		updated := accountSvc(dao.CRUDMock[reference.POSPaymentAccount]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
				return &query.Page[reference.POSPaymentAccount]{Items: []*reference.POSPaymentAccount{{Base: model.Base{ID: 1}, OrganizationID: 10, Method: "cash", AccountID: 100}}, Count: 1}, nil
			},
			UpdateFunc: func(_ context.Context, a *reference.POSPaymentAccount) (*reference.POSPaymentAccount, error) {
				return a, nil
			},
		})
		resp, err = doRequest(paymentAccountTestApp(t, updated, true), http.MethodPut, "/pos/payment-accounts/cash", `{"account_id":200}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects invalid input and missing org", func(t *testing.T) {
		svc := accountSvc(dao.CRUDMock[reference.POSPaymentAccount]{})
		resp, err := doRequest(paymentAccountTestApp(t, svc, true), http.MethodPut, "/pos/payment-accounts/cash", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(paymentAccountTestAppGuards(t, svc, false, bareGuards()), http.MethodPut, "/pos/payment-accounts/cash", `{"account_id":100}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("maps service errors", func(t *testing.T) {
		conflict := accountSvc(dao.CRUDMock[reference.POSPaymentAccount]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
				return &query.Page[reference.POSPaymentAccount]{}, nil
			},
			CreateFunc: func(_ context.Context, _ *reference.POSPaymentAccount) (*reference.POSPaymentAccount, error) {
				return nil, &pgconn.PgError{Code: "23505"}
			},
		})
		resp, err := doRequest(paymentAccountTestApp(t, conflict, true), http.MethodPut, "/pos/payment-accounts/cash", `{"account_id":100}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusConflict)

		failing := accountSvc(dao.CRUDMock[reference.POSPaymentAccount]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
				return nil, errors.New("db down")
			},
		})
		resp, err = doRequest(paymentAccountTestApp(t, failing, true), http.MethodPut, "/pos/payment-accounts/cash", `{"account_id":100}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}

func TestPaymentAccountHandler_Delete(t *testing.T) {
	t.Run("deletes mapping", func(t *testing.T) {
		svc := accountSvc(dao.CRUDMock[reference.POSPaymentAccount]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
				return &query.Page[reference.POSPaymentAccount]{Items: []*reference.POSPaymentAccount{{Base: model.Base{ID: 1}}}, Count: 1}, nil
			},
		})
		resp, err := doRequest(paymentAccountTestApp(t, svc, true), http.MethodDelete, "/pos/payment-accounts/cash", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNoContent)
	})

	t.Run("rejects missing org and maps errors", func(t *testing.T) {
		svc := accountSvc(dao.CRUDMock[reference.POSPaymentAccount]{})
		resp, err := doRequest(paymentAccountTestAppGuards(t, svc, false, bareGuards()), http.MethodDelete, "/pos/payment-accounts/cash", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(paymentAccountTestApp(t, svc, true), http.MethodDelete, "/pos/payment-accounts/cash", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		failing := accountSvc(dao.CRUDMock[reference.POSPaymentAccount]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
				return nil, errors.New("db down")
			},
		})
		resp, err = doRequest(paymentAccountTestApp(t, failing, true), http.MethodDelete, "/pos/payment-accounts/cash", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}
