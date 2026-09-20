package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type stubAgreementOrderEngine struct {
	order *procurement.PurchaseOrder
	err   error
}

func (m stubAgreementOrderEngine) CreateFromAgreement(_ context.Context, _ uint64, _ *procurement.CreateFromAgreementOverrides) (*procurement.PurchaseOrder, error) {
	return m.order, m.err
}

func agreementTestService(agreements procurement.SupplyAgreementDAOMock, lines procurement.SupplyAgreementLineDAOMock) procurement.SupplyAgreementService {
	contactDAO := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 10}}, nil
			},
		},
	}
	seqDAO := sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "AGR/00001"}, nil
		},
	}
	return procurement.NewSupplyAgreementService(agreements, lines, sequence.NewSequenceService(seqDAO), contactDAO)
}

func agreementTestApp(t *testing.T, svc procurement.SupplyAgreementService, engine procurement.CreateFromAgreementEngine) *fiber.App {
	t.Helper()
	return procurementTestApp(t, func(api fiber.Router, guards httpx.RouteGuards) {
		NewSupplyAgreementHandler(svc, engine).Register(api, guards)
	})
}

func sampleAgreement() *procurement.SupplyAgreement {
	return &procurement.SupplyAgreement{
		Base:           model.Base{ID: 1},
		OrganizationID: ptrUint64(10),
		Name:           ptrString("AGR/00001"),
		SupplierID:     10,
		State:          procurement.SupplyAgreementStateDraft,
	}
}

func TestSupplyAgreementHandler_List_Get(t *testing.T) {
	t.Run("lists agreements", func(t *testing.T) {
		agreements := procurement.SupplyAgreementDAOMock{}
		agreements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplyAgreement], error) {
			return &query.Page[procurement.SupplyAgreement]{Items: []*procurement.SupplyAgreement{sampleAgreement()}, Count: 1}, nil
		}
		app := agreementTestApp(t, agreementTestService(agreements, procurement.SupplyAgreementLineDAOMock{}), stubAgreementOrderEngine{})
		resp, err := doRequest(app, http.MethodGet, "/supply-agreements/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects invalid query and missing tenant and reports failure", func(t *testing.T) {
		svc := agreementTestService(procurement.SupplyAgreementDAOMock{}, procurement.SupplyAgreementLineDAOMock{})
		resp, err := doRequest(agreementTestApp(t, svc, stubAgreementOrderEngine{}), http.MethodGet, "/supply-agreements/?size=abc", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		app := procurementTestAppNoTenant(t, func(api fiber.Router, guards httpx.RouteGuards) {
			NewSupplyAgreementHandler(svc, stubAgreementOrderEngine{}).Register(api, guards)
		})
		resp, err = doRequest(app, http.MethodGet, "/supply-agreements/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)

		failing := procurement.SupplyAgreementDAOMock{}
		failing.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplyAgreement], error) {
			return nil, errors.New("db down")
		}
		resp, err = doRequest(agreementTestApp(t, agreementTestService(failing, procurement.SupplyAgreementLineDAOMock{}), stubAgreementOrderEngine{}), http.MethodGet, "/supply-agreements/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		svc := agreementTestService(procurement.SupplyAgreementDAOMock{}, procurement.SupplyAgreementLineDAOMock{})
		cases := []struct {
			method string
			path   string
			body   string
		}{
			{method: http.MethodGet, path: "/supply-agreements/abc"},
			{method: http.MethodPost, path: "/supply-agreements/abc/activate"},
			{method: http.MethodPost, path: "/supply-agreements/abc/cancel"},
			{method: http.MethodPost, path: "/supply-agreements/abc/create-order", body: `{}`},
		}
		for _, tc := range cases {
			resp, err := doRequest(agreementTestApp(t, svc, stubAgreementOrderEngine{}), tc.method, tc.path, tc.body)
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

	t.Run("returns agreement with lines", func(t *testing.T) {
		agreements := procurement.SupplyAgreementDAOMock{}
		agreements.FindFunc = func(_ context.Context, _ uint64) (*procurement.SupplyAgreement, error) {
			return sampleAgreement(), nil
		}
		svc := agreementTestService(agreements, procurement.SupplyAgreementLineDAOMock{})
		resp, err := doRequest(agreementTestApp(t, svc, stubAgreementOrderEngine{}), http.MethodGet, "/supply-agreements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("returns not found and reports failures", func(t *testing.T) {
		missing := procurement.SupplyAgreementDAOMock{}
		missing.FindFunc = func(_ context.Context, _ uint64) (*procurement.SupplyAgreement, error) {
			return nil, nil
		}
		resp, err := doRequest(agreementTestApp(t, agreementTestService(missing, procurement.SupplyAgreementLineDAOMock{}), stubAgreementOrderEngine{}), http.MethodGet, "/supply-agreements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)

		failing := procurement.SupplyAgreementDAOMock{}
		failing.FindFunc = func(_ context.Context, _ uint64) (*procurement.SupplyAgreement, error) {
			return nil, errors.New("db down")
		}
		resp, err = doRequest(agreementTestApp(t, agreementTestService(failing, procurement.SupplyAgreementLineDAOMock{}), stubAgreementOrderEngine{}), http.MethodGet, "/supply-agreements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)

		linesFailing := procurement.SupplyAgreementLineDAOMock{}
		linesFailing.ListByAgreementFunc = func(_ context.Context, _ uint64) ([]*procurement.SupplyAgreementLine, error) {
			return nil, errors.New("db down")
		}
		found := procurement.SupplyAgreementDAOMock{}
		found.FindFunc = func(_ context.Context, _ uint64) (*procurement.SupplyAgreement, error) {
			return sampleAgreement(), nil
		}
		resp, err = doRequest(agreementTestApp(t, agreementTestService(found, linesFailing), stubAgreementOrderEngine{}), http.MethodGet, "/supply-agreements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}

func TestSupplyAgreementHandler_Write(t *testing.T) {
	validBody := `{"supplier_id":10,"lines":[{"qty":2,"unit_price":1000}]}`

	t.Run("creates agreement", func(t *testing.T) {
		svc := agreementTestService(procurement.SupplyAgreementDAOMock{}, procurement.SupplyAgreementLineDAOMock{})
		resp, err := doRequest(agreementTestApp(t, svc, stubAgreementOrderEngine{}), http.MethodPost, "/supply-agreements/", validBody)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)
	})

	t.Run("rejects invalid body and dates", func(t *testing.T) {
		svc := agreementTestService(procurement.SupplyAgreementDAOMock{}, procurement.SupplyAgreementLineDAOMock{})
		resp, err := doRequest(agreementTestApp(t, svc, stubAgreementOrderEngine{}), http.MethodPost, "/supply-agreements/", `{"supplier_id":0}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(agreementTestApp(t, svc, stubAgreementOrderEngine{}), http.MethodPost, "/supply-agreements/", `{"supplier_id":10,"start_date":"yesterday","lines":[{"qty":2,"unit_price":1000}]}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(agreementTestApp(t, svc, stubAgreementOrderEngine{}), http.MethodPost, "/supply-agreements/", `{"supplier_id":10,"end_date":"tomorrow","lines":[{"qty":2,"unit_price":1000}]}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(agreementTestApp(t, svc, stubAgreementOrderEngine{}), http.MethodPost, "/supply-agreements/", `{"supplier_id":10,"lines":[{"qty":2,"unit_price":1000,"needed_by":"someday"}]}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	agreementErrors := []struct {
		name   string
		svcErr error
		status int
	}{
		{name: "not found", svcErr: procurement.ErrAgreementNotFound, status: http.StatusNotFound},
		{name: "bad state", svcErr: procurement.ErrSupplyAgreementState, status: http.StatusUnprocessableEntity},
		{name: "not active", svcErr: procurement.ErrAgreementNotActive, status: http.StatusUnprocessableEntity},
		{name: "no lines", svcErr: procurement.ErrAgreementNoLines, status: http.StatusUnprocessableEntity},
		{name: "bad line qty", svcErr: procurement.ErrAgreementLineQty, status: http.StatusUnprocessableEntity},
		{name: "supplier missing", svcErr: procurement.ErrAgreementVendor, status: http.StatusNotFound},
		{name: "expired", svcErr: procurement.ErrAgreementExpired, status: http.StatusUnprocessableEntity},
		{name: "qty exceeded", svcErr: procurement.ErrAgreementQtyExceeded, status: http.StatusConflict},
		{name: "amount exceeded", svcErr: procurement.ErrAgreementAmountExceeded, status: http.StatusConflict},
		{name: "unexpected", svcErr: errors.New("db down"), status: http.StatusInternalServerError},
	}

	t.Run("activate maps errors", func(t *testing.T) {
		for _, tt := range agreementErrors {
			agreements := procurement.SupplyAgreementDAOMock{}
			agreements.FindFunc = func(_ context.Context, _ uint64) (*procurement.SupplyAgreement, error) {
				return nil, tt.svcErr
			}
			svc := agreementTestService(agreements, procurement.SupplyAgreementLineDAOMock{})
			resp, err := doRequest(agreementTestApp(t, svc, stubAgreementOrderEngine{}), http.MethodPost, "/supply-agreements/1/activate", "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Errorf("%s: status = %d, want %d", tt.name, resp.StatusCode, tt.status)
			}
		}
	})

	t.Run("activates and cancels", func(t *testing.T) {
		agreements := procurement.SupplyAgreementDAOMock{}
		agreements.FindFunc = func(_ context.Context, _ uint64) (*procurement.SupplyAgreement, error) {
			return sampleAgreement(), nil
		}
		svc := agreementTestService(agreements, procurement.SupplyAgreementLineDAOMock{})
		for _, action := range []string{"activate", "cancel"} {
			resp, err := doRequest(agreementTestApp(t, svc, stubAgreementOrderEngine{}), http.MethodPost, "/supply-agreements/1/"+action, "")
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("cancel maps not found", func(t *testing.T) {
		agreements := procurement.SupplyAgreementDAOMock{}
		agreements.FindFunc = func(_ context.Context, _ uint64) (*procurement.SupplyAgreement, error) {
			return nil, nil
		}
		svc := agreementTestService(agreements, procurement.SupplyAgreementLineDAOMock{})
		resp, err := doRequest(agreementTestApp(t, svc, stubAgreementOrderEngine{}), http.MethodPost, "/supply-agreements/1/cancel", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
	})

	t.Run("creates order from agreement", func(t *testing.T) {
		svc := agreementTestService(procurement.SupplyAgreementDAOMock{}, procurement.SupplyAgreementLineDAOMock{})
		engine := stubAgreementOrderEngine{order: samplePurchaseOrder()}
		resp, err := doRequest(agreementTestApp(t, svc, engine), http.MethodPost, "/supply-agreements/1/create-order", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)
	})

	t.Run("create order maps errors", func(t *testing.T) {
		svc := agreementTestService(procurement.SupplyAgreementDAOMock{}, procurement.SupplyAgreementLineDAOMock{})
		engine := stubAgreementOrderEngine{err: procurement.ErrAgreementNotFound}
		resp, err := doRequest(agreementTestApp(t, svc, engine), http.MethodPost, "/supply-agreements/1/create-order", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		failing := stubAgreementOrderEngine{err: errors.New("db down")}
		resp, err = doRequest(agreementTestApp(t, svc, failing), http.MethodPost, "/supply-agreements/1/create-order", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}
