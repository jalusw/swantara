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
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

func costCenterTestApp(t *testing.T, costCenters dao.CRUDMock[procurement.CostCenter]) *fiber.App {
	t.Helper()
	return procurementTestApp(t, func(api fiber.Router, guards httpx.RouteGuards) {
		NewCostCenterHandler(procurement.NewCostCenterService(costCenters)).Register(api, guards)
	})
}

func sampleCostCenter(orgID uint64) *procurement.CostCenter {
	return &procurement.CostCenter{
		Base:           model.Base{ID: 1},
		OrganizationID: &orgID,
		Name:           ptrString("Marketing"),
		Code:           ptrString("MKT"),
		Active:         true,
	}
}

func TestCostCenterHandler_CRUD(t *testing.T) {
	t.Run("lists cost centers", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.CostCenter], error) {
				return &query.Page[procurement.CostCenter]{Items: []*procurement.CostCenter{sampleCostCenter(10)}, Count: 1}, nil
			},
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodGet, "/cost-centers/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects invalid query", func(t *testing.T) {
		resp, err := doRequest(costCenterTestApp(t, dao.CRUDMock[procurement.CostCenter]{}), http.MethodGet, "/cost-centers/?size=abc", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("rejects missing tenant on list", func(t *testing.T) {
		app := procurementTestAppNoTenant(t, func(api fiber.Router, guards httpx.RouteGuards) {
			NewCostCenterHandler(procurement.NewCostCenterService(dao.CRUDMock[procurement.CostCenter]{})).Register(api, guards)
		})
		resp, err := doRequest(app, http.MethodGet, "/cost-centers/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)
	})

	t.Run("reports list failure", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.CostCenter], error) {
				return nil, errors.New("db down")
			},
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodGet, "/cost-centers/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			resp, err := doRequest(costCenterTestApp(t, dao.CRUDMock[procurement.CostCenter]{}), method, "/cost-centers/abc", "")
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

	t.Run("returns cost center", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.CostCenter, error) {
				return sampleCostCenter(10), nil
			},
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodGet, "/cost-centers/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("returns not found for missing or foreign", func(t *testing.T) {
		for _, orgID := range []uint64{0, 11} {
			var cc *procurement.CostCenter
			if orgID != 0 {
				cc = sampleCostCenter(orgID)
			}
			mock := dao.CRUDMock[procurement.CostCenter]{
				FindFunc: func(_ context.Context, _ uint64) (*procurement.CostCenter, error) {
					return cc, nil
				},
			}
			resp, err := doRequest(costCenterTestApp(t, mock), http.MethodGet, "/cost-centers/1", "")
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("reports lookup failure", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.CostCenter, error) {
				return nil, errors.New("db down")
			},
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodGet, "/cost-centers/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("creates cost center", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			CreateFunc: func(_ context.Context, cc *procurement.CostCenter) (*procurement.CostCenter, error) {
				cc.ID = 5
				return cc, nil
			},
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodPost, "/cost-centers/", `{"name":"Marketing","code":"MKT"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)
	})

	t.Run("rejects invalid create body", func(t *testing.T) {
		resp, err := doRequest(costCenterTestApp(t, dao.CRUDMock[procurement.CostCenter]{}), http.MethodPost, "/cost-centers/", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("reports create failure", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			CreateFunc: func(_ context.Context, _ *procurement.CostCenter) (*procurement.CostCenter, error) {
				return nil, errors.New("db down")
			},
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodPost, "/cost-centers/", `{"name":"Marketing","code":"MKT"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("updates cost center", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.CostCenter, error) {
				return sampleCostCenter(10), nil
			},
			UpdateFunc: func(_ context.Context, cc *procurement.CostCenter) (*procurement.CostCenter, error) {
				return cc, nil
			},
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodPut, "/cost-centers/1", `{"name":"Sales","active":false}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects update for missing", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.CostCenter, error) {
				return nil, nil
			},
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodPut, "/cost-centers/1", `{"name":"Sales"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
	})

	t.Run("reports update failure", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.CostCenter, error) {
				return sampleCostCenter(10), nil
			},
			UpdateFunc: func(_ context.Context, _ *procurement.CostCenter) (*procurement.CostCenter, error) {
				return nil, errors.New("db down")
			},
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodPut, "/cost-centers/1", `{"name":"Sales"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("deletes cost center", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.CostCenter, error) {
				return sampleCostCenter(10), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error { return nil },
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodDelete, "/cost-centers/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects delete for foreign", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.CostCenter, error) {
				return sampleCostCenter(11), nil
			},
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodDelete, "/cost-centers/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
	})

	t.Run("reports delete failure", func(t *testing.T) {
		mock := dao.CRUDMock[procurement.CostCenter]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.CostCenter, error) {
				return sampleCostCenter(10), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error { return errors.New("db down") },
		}
		resp, err := doRequest(costCenterTestApp(t, mock), http.MethodDelete, "/cost-centers/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}
