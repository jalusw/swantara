package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/organization"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func moduleSvcWithStore(store dao.CRUDMock[reference.OrganizationModule]) organization.Service {
	svc := organizationTestSvc(
		dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
				return sampleOrganization(), nil
			},
		},
		handlerCurrencySearchMock{},
		handlerMemberProvisionerMock{},
	)
	svc.SetModuleStore(store)
	return svc
}

func TestOrganizationHandler_Modules_Errors(t *testing.T) {
	t.Run("list modules reports failure", func(t *testing.T) {
		orgs := dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
				return sampleOrganization(), nil
			},
		}
		svc := moduleSvcWithStore(dao.CRUDMock[reference.OrganizationModule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.OrganizationModule], error) {
				return nil, errors.New("db down")
			},
		})
		app := organizationHandlerTest(t, orgs, svc)
		resp, err := doRequest(app, http.MethodGet, "/organizations/1/modules", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("update rejects unknown module", func(t *testing.T) {
		orgs := dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
				return sampleOrganization(), nil
			},
		}
		svc := moduleSvcWithStore(dao.CRUDMock[reference.OrganizationModule]{})
		app := organizationHandlerTest(t, orgs, svc)
		resp, err := doRequest(app, http.MethodPut, "/organizations/1/modules", `{"module_id":"nope","active":true}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("update reports store failure", func(t *testing.T) {
		orgs := dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
				return sampleOrganization(), nil
			},
		}
		svc := moduleSvcWithStore(dao.CRUDMock[reference.OrganizationModule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.OrganizationModule], error) {
				return nil, errors.New("db down")
			},
		})
		app := organizationHandlerTest(t, orgs, svc)
		resp, err := doRequest(app, http.MethodPut, "/organizations/1/modules", `{"module_id":"crm","active":true}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("module store model", func(t *testing.T) {
		module := &reference.OrganizationModule{Base: model.Base{ID: 1}, OrganizationID: 10, ModuleID: "crm", Active: true}
		if module.ModuleID != "crm" {
			t.Errorf("module = %+v", module)
		}
		if !organization.IsKnownModule("crm") || organization.IsKnownModule("nope") {
			t.Error("IsKnownModule mismatch")
		}
	})
}
