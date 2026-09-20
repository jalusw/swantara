package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

func supplierProductOffer(
	variants products.ItemVariantDAO,
	templates products.ItemDAO,
	supplierProducts products.SupplierProductDAO,
	contactDAO contacts.ContactDAO,
	supplierDAO contacts.SupplierProfileDAO,
) products.SupplierProductService {
	return supplierProductTestSvc(variants, templates, supplierProducts, contactDAO, supplierDAO)
}

func validSupplierProductDeps() (products.ItemVariantDAOMock, products.ItemDAOMock, contacts.ContactDAOMock, contacts.SupplierProfileDAOMock) {
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return sampleVariant(), nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	contactDAO := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10))}, nil
			},
		},
	}
	supplierDAO := contacts.SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
			return &contacts.SupplierProfile{Base: model.Base{ID: 7}, ContactID: 7, Active: true}, nil
		},
	}
	return variants, templates, contactDAO, supplierDAO
}

func TestSupplierProductHandler_List_ReturnsSupplierProducts(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[products.SupplierProduct], error) {
				return &query.Page[products.SupplierProduct]{Items: []*products.SupplierProduct{sampleSupplierProduct()}, Count: 1}, nil
			},
		},
	}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	resp, err := doRequest(app, http.MethodGet, "/supplier-products/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierProductHandler_List_ExportsCSV(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[products.SupplierProduct], error) {
				return &query.Page[products.SupplierProduct]{Items: []*products.SupplierProduct{sampleSupplierProduct()}, Count: 1}, nil
			},
		},
	}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	resp, err := doRequest(app, http.MethodGet, "/supplier-products/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierProductHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, products.SupplierProductDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/supplier-products/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_List_RequiresOrganization(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTestNoOrg(t, products.SupplierProductDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/supplier-products/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_List_ReturnsServerError(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[products.SupplierProduct], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, supplierProducts, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	resp, err := doRequest(app, http.MethodGet, "/supplier-products/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Get_ReturnsSupplierProduct(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return sampleSupplierProduct(), nil
			},
		},
	}
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, supplierProducts, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	resp, err := doRequest(app, http.MethodGet, "/supplier-products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, products.SupplierProductDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/supplier-products/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Get_RequiresOrganization(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTestNoOrg(t, products.SupplierProductDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/supplier-products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Get_ReturnsNotFound(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return nil, nil
			},
		},
	}
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, supplierProducts, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	resp, err := doRequest(app, http.MethodGet, "/supplier-products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Get_ReturnsServerError(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, supplierProducts, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	resp, err := doRequest(app, http.MethodGet, "/supplier-products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Create_CreatesSupplierProduct(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			CreateFunc: func(_ context.Context, offer *products.SupplierProduct) (*products.SupplierProduct, error) {
				offer.ID = 1
				return offer, nil
			},
		},
	}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"item_id":1,"supplier_id":7,"min_qty":1,"price":100,"priority":10}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Create_RejectsValidation(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, products.SupplierProductDAOMock{}, svc)

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Create_RequiresOrganization(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTestNoOrg(t, products.SupplierProductDAOMock{}, svc)

	body := `{"item_id":1,"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Create_ReturnsNotFoundForVariant(t *testing.T) {
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return nil, nil
			},
		},
	}
	supplierProducts := products.SupplierProductDAOMock{}
	svc := supplierProductOffer(variants, products.ItemDAOMock{}, supplierProducts, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"item_id":1,"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Create_ReturnsNotFoundForVendor(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{}
	variants, templates, _, supplierDAO := validSupplierProductDeps()
	contactDAO := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return nil, nil
			},
		},
	}
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"item_id":1,"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Create_RejectsNonSupplierVendor(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{}
	variants, templates, contactDAO, _ := validSupplierProductDeps()
	supplierDAO := contacts.SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
			return nil, nil
		},
	}
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"item_id":1,"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Create_RejectsInvalidValidity(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"item_id":1,"supplier_id":7,"valid_from":"2026-06-01T00:00:00Z","valid_to":"2026-01-01T00:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Create_RejectsNegativeMinQty(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"item_id":1,"supplier_id":7,"min_qty":-1}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Create_ReturnsServerError(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			CreateFunc: func(_ context.Context, _ *products.SupplierProduct) (*products.SupplierProduct, error) {
				return nil, errors.New("db down")
			},
		},
	}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"item_id":1,"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Update_UpdatesSupplierProduct(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return sampleSupplierProduct(), nil
			},
			UpdateFunc: func(_ context.Context, offer *products.SupplierProduct) (*products.SupplierProduct, error) {
				return offer, nil
			},
		},
	}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"supplier_id":7,"price":120,"priority":5}`
	resp, err := doRequest(app, http.MethodPut, "/supplier-products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Update_RejectsInvalidID(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, products.SupplierProductDAOMock{}, svc)

	body := `{"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPut, "/supplier-products/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Update_RequiresOrganization(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTestNoOrg(t, products.SupplierProductDAOMock{}, svc)

	body := `{"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPut, "/supplier-products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Update_RejectsValidation(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, products.SupplierProductDAOMock{}, svc)

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/supplier-products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Update_ReturnsNotFound(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return nil, nil
			},
		},
	}
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, supplierProducts, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPut, "/supplier-products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, supplierProducts, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPut, "/supplier-products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return sampleSupplierProduct(), nil
			},
			UpdateFunc: func(_ context.Context, _ *products.SupplierProduct) (*products.SupplierProduct, error) {
				return nil, errors.New("db down")
			},
		},
	}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPut, "/supplier-products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Update_RejectsNonSupplierVendor(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return sampleSupplierProduct(), nil
			},
		},
	}
	variants, templates, contactDAO, _ := validSupplierProductDeps()
	supplierDAO := contacts.SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
			return &contacts.SupplierProfile{Base: model.Base{ID: 7}, ContactID: 7, Active: false}, nil
		},
	}
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPut, "/supplier-products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Delete_DeletesSupplierProduct(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return sampleSupplierProduct(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, supplierProducts, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	resp, err := doRequest(app, http.MethodDelete, "/supplier-products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Delete_RejectsInvalidID(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, products.SupplierProductDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/supplier-products/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Delete_RequiresOrganization(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTestNoOrg(t, products.SupplierProductDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/supplier-products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Delete_ReturnsNotFound(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return nil, nil
			},
		},
	}
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, supplierProducts, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	resp, err := doRequest(app, http.MethodDelete, "/supplier-products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, supplierProducts, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	resp, err := doRequest(app, http.MethodDelete, "/supplier-products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierProductHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[products.SupplierProduct]{
			FindFunc: func(_ context.Context, _ uint64) (*products.SupplierProduct, error) {
				return sampleSupplierProduct(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return errors.New("db down")
			},
		},
	}
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, supplierProducts, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	resp, err := doRequest(app, http.MethodDelete, "/supplier-products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierProductHandler_BestOffer_ResolvesBestOffer(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*products.SupplierProduct, error) {
			return []*products.SupplierProduct{sampleSupplierProduct()}, nil
		},
	}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"item_id":1,"qty":"2"}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/best-offer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierProductHandler_BestOffer_ResolvesBestOfferWithDate(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*products.SupplierProduct, error) {
			return []*products.SupplierProduct{sampleSupplierProduct()}, nil
		},
	}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"item_id":1,"qty":"2","date":"2026-08-01"}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/best-offer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierProductHandler_BestOffer_RejectsValidation(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, products.SupplierProductDAOMock{}, svc)

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/best-offer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_BestOffer_RequiresOrganization(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTestNoOrg(t, products.SupplierProductDAOMock{}, svc)

	body := `{"item_id":1,"qty":"2"}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/best-offer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_BestOffer_RejectsInvalidQuantity(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, products.SupplierProductDAOMock{}, svc)

	body := `{"item_id":1,"qty":"abc"}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/best-offer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_BestOffer_RejectsInvalidDate(t *testing.T) {
	svc := supplierProductOffer(products.ItemVariantDAOMock{}, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, products.SupplierProductDAOMock{}, svc)

	body := `{"item_id":1,"qty":"2","date":"bogus"}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/best-offer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_BestOffer_ReturnsNotFoundForVariant(t *testing.T) {
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return nil, nil
			},
		},
	}
	svc := supplierProductOffer(variants, products.ItemDAOMock{}, products.SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := supplierProductHandlerTest(t, products.SupplierProductDAOMock{}, svc)

	body := `{"item_id":1,"qty":"2"}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/best-offer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierProductHandler_BestOffer_RejectsNoValidOffer(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*products.SupplierProduct, error) {
			return []*products.SupplierProduct{}, nil
		},
	}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"item_id":1,"qty":"2"}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/best-offer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierProductHandler_BestOffer_ReturnsServerError(t *testing.T) {
	supplierProducts := products.SupplierProductDAOMock{
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*products.SupplierProduct, error) {
			return nil, errors.New("db down")
		},
	}
	variants, templates, contactDAO, supplierDAO := validSupplierProductDeps()
	svc := supplierProductOffer(variants, templates, supplierProducts, contactDAO, supplierDAO)
	app := supplierProductHandlerTest(t, supplierProducts, svc)

	body := `{"item_id":1,"qty":"2"}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-products/best-offer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierProductHandler_WriteSupplierProductError_MapsErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "variant not found", err: products.ErrVariantNotFound, want: http.StatusNotFound},
		{name: "supplier not found", err: products.ErrSupplierNotFound, want: http.StatusNotFound},
		{name: "supplier not supplier", err: products.ErrSupplierNotSupplier, want: http.StatusUnprocessableEntity},
		{name: "invalid validity", err: products.ErrInvalidValidity, want: http.StatusUnprocessableEntity},
		{name: "invalid min qty", err: products.ErrInvalidMinQty, want: http.StatusUnprocessableEntity},
		{name: "no valid offer", err: products.ErrNoValidOffer, want: http.StatusUnprocessableEntity},
		{name: "unexpected error", err: errors.New("boom"), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/err", func(c fiber.Ctx) error {
				return writeSupplierProductError(c, tt.err)
			})

			resp, err := doRequest(app, http.MethodPost, "/err", "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}
