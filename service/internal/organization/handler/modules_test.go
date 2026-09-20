package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/organization"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func moduleOrgs() dao.CRUDMock[reference.Organization] {
	return dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return sampleOrganization(), nil
		},
	}
}

func TestOrganizationHandler_ListModules_ReturnsAllActiveByDefault(t *testing.T) {
	orgs := moduleOrgs()
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodGet, "/organizations/1/modules", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var envelope struct {
		Data struct {
			Modules []ModuleResponse `json:"modules"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Data.Modules) != len(organization.OrgModules) {
		t.Fatalf("modules = %d, want %d", len(envelope.Data.Modules), len(organization.OrgModules))
	}
	for _, module := range envelope.Data.Modules {
		if !module.Active {
			t.Errorf("module %q inactive, want active by default", module.ModuleID)
		}
	}
}

func TestOrganizationHandler_ListModules_ReturnsNotFoundForMissingOrg(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodGet, "/organizations/9/modules", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOrganizationHandler_UpdateModule_TogglesActive(t *testing.T) {
	orgs := moduleOrgs()
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	svc.SetModuleStore(dao.CRUDMock[reference.OrganizationModule]{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodPut, "/organizations/1/modules", `{"module_id":"hr","active":false}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var envelope struct {
		Data struct {
			Module ModuleResponse `json:"module"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Module.ModuleID != "hr" || envelope.Data.Module.Active {
		t.Errorf("module = %+v, want hr inactive", envelope.Data.Module)
	}
}

func TestOrganizationHandler_UpdateModule_RejectsUnknownModule(t *testing.T) {
	orgs := moduleOrgs()
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	svc.SetModuleStore(dao.CRUDMock[reference.OrganizationModule]{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodPut, "/organizations/1/modules", `{"module_id":"nope","active":false}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}
