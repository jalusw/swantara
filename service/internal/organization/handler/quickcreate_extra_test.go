package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/organization"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func quickCreateSvc() organization.Service {
	orgs := dao.CRUDMock[reference.Organization]{
		CreateFunc: func(_ context.Context, org *reference.Organization) (*reference.Organization, error) {
			org.ID = 11
			return org, nil
		},
	}
	currencies := handlerCurrencySearchMock{
		search: func(_ context.Context, _ string, value any) (*reference.Currency, error) {
			if value == "IDR" {
				return &reference.Currency{Code: "IDR"}, nil
			}
			return nil, nil
		},
	}
	return organizationTestSvc(orgs, currencies, handlerMemberProvisionerMock{})
}

func TestOrganizationHandler_QuickCreate(t *testing.T) {
	t.Run("creates organization", func(t *testing.T) {
		orgs := dao.CRUDMock[reference.Organization]{}
		app := organizationHandlerTest(t, orgs, quickCreateSvc())
		resp, err := doRequest(app, http.MethodPost, "/organizations/quick", `{"name":"Toko","country_code":"ID"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)
	})

	t.Run("rejects missing caller and invalid body", func(t *testing.T) {
		orgs := dao.CRUDMock[reference.Organization]{}
		app := organizationHandlerTestNoCaller(t, orgs, quickCreateSvc())
		resp, err := doRequest(app, http.MethodPost, "/organizations/quick", `{"name":"Toko","country_code":"ID"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)

		app = organizationHandlerTest(t, orgs, quickCreateSvc())
		resp, err = doRequest(app, http.MethodPost, "/organizations/quick", `{"name":""}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("maps service errors", func(t *testing.T) {
		orgs := dao.CRUDMock[reference.Organization]{}
		app := organizationHandlerTest(t, orgs, quickCreateSvc())
		resp, err := doRequest(app, http.MethodPost, "/organizations/quick", `{"name":"Toko","country_code":"XX"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		currencies := handlerCurrencySearchMock{
			search: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
				return nil, errors.New("db down")
			},
		}
		svc := organization.NewOrganizationService(dao.CRUDMock[reference.Organization]{}, currencies, handlerMemberProvisionerMock{})
		app = organizationHandlerTest(t, dao.CRUDMock[reference.Organization]{}, svc)
		resp, err = doRequest(app, http.MethodPost, "/organizations/quick", `{"name":"Toko","country_code":"ID"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}

func TestOrganizationHandler_Mocks(t *testing.T) {
	t.Run("currency search fallback", func(t *testing.T) {
		got, err := handlerCurrencySearchMock{}.Search(context.Background(), "code", "IDR")
		if err != nil {
			t.Errorf("Search = %v", err)
		}
		if got != nil {
			t.Errorf("Search = %+v, want nil", got)
		}
	})

	t.Run("provision fallback", func(t *testing.T) {
		if err := (handlerMemberProvisionerMock{}).ProvisionOwner(context.Background(), 1, 2); err != nil {
			t.Errorf("ProvisionOwner = %v", err)
		}
	})

	t.Run("write default error", func(t *testing.T) {
		orgs := dao.CRUDMock[reference.Organization]{}
		currencies := handlerCurrencySearchMock{
			search: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
				return &reference.Currency{Code: "IDR"}, nil
			},
		}
		members := handlerMemberProvisionerMock{
			provision: func(_ context.Context, _, _ uint64) error { return errors.New("provision down") },
		}
		svc := organization.NewOrganizationService(orgs, currencies, members)
		app := organizationHandlerTest(t, orgs, svc)
		resp, err := doRequest(app, http.MethodPost, "/organizations/quick", `{"name":"Toko","country_code":"ID"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("sample organization", func(t *testing.T) {
		org := sampleOrganization()
		if org.Name != "Acme" {
			t.Errorf("name = %q", org.Name)
		}
		_ = model.Base{}
	})
}
