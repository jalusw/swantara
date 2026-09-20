package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestProductHandler_List_ReturnsProducts(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[products.Item], error) {
				return &query.Page[products.Item]{Items: []*products.Item{sampleProduct()}, Count: 1}, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductHandler_List_ExportsCSV(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[products.Item], error) {
				return &query.Page[products.Item]{Items: []*products.Item{sampleProduct()}, Count: 1}, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, products.ItemDAOMock{}, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_List_RequiresTenant(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTestNoTenant(t, products.ItemDAOMock{}, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestProductHandler_List_ReturnsServerError(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[products.Item], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_Get_ReturnsProduct(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductHandler_Get_ReturnsNotFound(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductHandler_Get_ReturnsNotFoundForForeignTenant(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return foreignProduct(), nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, products.ItemDAOMock{}, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_Get_ReturnsServerError(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_Create_CreatesProduct(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			CreateFunc: func(_ context.Context, template *products.Item) (*products.Item, error) {
				template.ID = 1
				return template, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget","type":"service"}`
	resp, err := doRequest(app, http.MethodPost, "/products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestProductHandler_Create_CreatesProductWithVariants(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			CreateFunc: func(_ context.Context, template *products.Item) (*products.Item, error) {
				template.ID = 1
				return template, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			CreateFunc: func(_ context.Context, variant *products.ItemVariant) (*products.ItemVariant, error) {
				variant.ID = 2
				return variant, nil
			},
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, variants, svc)

	body := `{"name":"Widget","variants":[{"sku":"W-1","active":false}]}`
	resp, err := doRequest(app, http.MethodPost, "/products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestProductHandler_Create_CreatesProductWithFlags(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			CreateFunc: func(_ context.Context, template *products.Item) (*products.Item, error) {
				template.ID = 1
				return template, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget","is_purchasable":false,"is_sellable":false,"active":false}`
	resp, err := doRequest(app, http.MethodPost, "/products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestProductHandler_Create_CreatesProductFromAttributeMatrix(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			CreateFunc: func(_ context.Context, template *products.Item) (*products.Item, error) {
				template.ID = 1
				return template, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget","attribute_matrix":[{"name":"color","values":["red","blue"]}]}`
	resp, err := doRequest(app, http.MethodPost, "/products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestProductHandler_Create_ReturnsServerErrorGeneratingVariants(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			CreateFunc: func(_ context.Context, template *products.Item) (*products.Item, error) {
				template.ID = 1
				return template, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget","attribute_matrix":[{"name":"color","values":["red","blue"]}]}`
	resp, err := doRequest(app, http.MethodPost, "/products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_Create_RejectsValidation(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, products.ItemDAOMock{}, products.ItemVariantDAOMock{}, svc)

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_Create_RejectsUnknownCategory(t *testing.T) {
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return nil, nil
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, categories, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, products.ItemDAOMock{}, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget","category_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_Create_RejectsForeignCategory(t *testing.T) {
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return foreignItemCategory(), nil
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, categories, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, products.ItemDAOMock{}, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget","category_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_Create_RejectsVariantTemplateMissing(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			CreateFunc: func(_ context.Context, template *products.Item) (*products.Item, error) {
				template.ID = 1
				return template, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget","variants":[{"sku":"W-1"}]}`
	resp, err := doRequest(app, http.MethodPost, "/products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_Create_RejectsDuplicateSku(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			CreateFunc: func(_ context.Context, template *products.Item) (*products.Item, error) {
				template.ID = 1
				return template, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 99}}, nil
			},
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, variants, svc)

	body := `{"name":"Widget","variants":[{"sku":"W-1"}]}`
	resp, err := doRequest(app, http.MethodPost, "/products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_Create_ReturnsServerError(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			CreateFunc: func(_ context.Context, _ *products.Item) (*products.Item, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget"}`
	resp, err := doRequest(app, http.MethodPost, "/products/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_Update_UpdatesProduct(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
			UpdateFunc: func(_ context.Context, template *products.Item) (*products.Item, error) {
				return template, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget Pro"}`
	resp, err := doRequest(app, http.MethodPut, "/products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductHandler_Update_UpdatesProductWithFlags(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
			UpdateFunc: func(_ context.Context, template *products.Item) (*products.Item, error) {
				return template, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget Pro","is_purchasable":false,"is_sellable":false,"active":false}`
	resp, err := doRequest(app, http.MethodPut, "/products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductHandler_Update_RejectsInvalidID(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, products.ItemDAOMock{}, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget Pro"}`
	resp, err := doRequest(app, http.MethodPut, "/products/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_Update_RejectsValidation(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_Update_ReturnsNotFound(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget Pro"}`
	resp, err := doRequest(app, http.MethodPut, "/products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductHandler_Update_ReturnsNotFoundForForeignTenant(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return foreignProduct(), nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget Pro"}`
	resp, err := doRequest(app, http.MethodPut, "/products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget Pro"}`
	resp, err := doRequest(app, http.MethodPut, "/products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
			UpdateFunc: func(_ context.Context, _ *products.Item) (*products.Item, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget Pro"}`
	resp, err := doRequest(app, http.MethodPut, "/products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_Update_RejectsForeignCategory(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return foreignItemCategory(), nil
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, categories, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"name":"Widget Pro","category_id":5}`
	resp, err := doRequest(app, http.MethodPut, "/products/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_Delete_DeletesProduct(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestProductHandler_Delete_RejectsInvalidID(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, products.ItemDAOMock{}, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/products/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_Delete_ReturnsNotFound(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return errors.New("referenced")
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/products/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_ListVariants_ReturnsVariants(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		ListByTemplateFunc: func(_ context.Context, _ uint64) ([]*products.ItemVariant, error) {
			return []*products.ItemVariant{sampleVariant()}, nil
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, variants, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/1/variants", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductHandler_ListVariants_ExportsCSV(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		ListByTemplateFunc: func(_ context.Context, _ uint64) ([]*products.ItemVariant, error) {
			return []*products.ItemVariant{sampleVariant()}, nil
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, variants, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/1/variants?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductHandler_ListVariants_RejectsInvalidID(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, products.ItemDAOMock{}, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/abc/variants", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_ListVariants_ReturnsNotFound(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/1/variants", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductHandler_ListVariants_ReturnsServerErrorOnFind(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/1/variants", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_ListVariants_ReturnsServerError(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		ListByTemplateFunc: func(_ context.Context, _ uint64) ([]*products.ItemVariant, error) {
			return nil, errors.New("db down")
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, variants, svc)

	resp, err := doRequest(app, http.MethodGet, "/products/1/variants", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_CreateVariant_CreatesVariant(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			CreateFunc: func(_ context.Context, variant *products.ItemVariant) (*products.ItemVariant, error) {
				variant.ID = 2
				return variant, nil
			},
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, variants, svc)

	body := `{"sku":"W-1","extra_cost":5}`
	resp, err := doRequest(app, http.MethodPost, "/products/1/variants", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestProductHandler_CreateVariant_CreatesInactiveVariant(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			CreateFunc: func(_ context.Context, variant *products.ItemVariant) (*products.ItemVariant, error) {
				variant.ID = 2
				return variant, nil
			},
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, variants, svc)

	body := `{"sku":"W-1","active":false}`
	resp, err := doRequest(app, http.MethodPost, "/products/1/variants", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestProductHandler_CreateVariant_RejectsInvalidID(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, products.ItemDAOMock{}, products.ItemVariantDAOMock{}, svc)

	body := `{"sku":"W-1"}`
	resp, err := doRequest(app, http.MethodPost, "/products/abc/variants", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_CreateVariant_RejectsMalformedBody(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, products.ItemDAOMock{}, products.ItemVariantDAOMock{}, svc)

	body := `{"sku":`
	resp, err := doRequest(app, http.MethodPost, "/products/1/variants", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestProductHandler_CreateVariant_ReturnsNotFound(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"sku":"W-1"}`
	resp, err := doRequest(app, http.MethodPost, "/products/1/variants", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductHandler_CreateVariant_ReturnsServerErrorOnFind(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(templates, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, products.ItemVariantDAOMock{}, svc)

	body := `{"sku":"W-1"}`
	resp, err := doRequest(app, http.MethodPost, "/products/1/variants", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_CreateVariant_RejectsDuplicateSku(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 99}}, nil
			},
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, variants, svc)

	body := `{"sku":"W-1"}`
	resp, err := doRequest(app, http.MethodPost, "/products/1/variants", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductHandler_CreateVariant_ReturnsServerErrorOnCreate(t *testing.T) {
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			CreateFunc: func(_ context.Context, _ *products.ItemVariant) (*products.ItemVariant, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := productHandlerTest(t, templates, variants, svc)

	body := `{"sku":"W-1"}`
	resp, err := doRequest(app, http.MethodPost, "/products/1/variants", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductHandler_WriteProductError_MapsErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "template category", err: products.ErrItemCategory, want: http.StatusUnprocessableEntity},
		{name: "template category organization", err: products.ErrItemCategoryOrganization, want: http.StatusUnprocessableEntity},
		{name: "template not found", err: products.ErrItemNotFound, want: http.StatusNotFound},
		{name: "variant not found", err: products.ErrVariantNotFound, want: http.StatusNotFound},
		{name: "variant template", err: products.ErrVariantTemplate, want: http.StatusUnprocessableEntity},
		{name: "sku taken", err: products.ErrSkuTaken, want: http.StatusUnprocessableEntity},
		{name: "empty attribute matrix", err: products.ErrEmptyAttributeMatrix, want: http.StatusUnprocessableEntity},
		{name: "unexpected error", err: errors.New("boom"), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/err", func(c fiber.Ctx) error {
				return writeProductError(c, tt.err)
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
