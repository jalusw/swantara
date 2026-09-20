package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func bomTestApp(t *testing.T, svc manufacturing.RecipeService) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewRecipeHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func sampleRecipe() *manufacturing.Recipe {
	return &manufacturing.Recipe{
		Base:           model.Base{ID: 1},
		OrganizationID: uint64Ptr(10),
		ItemID:         200,
		Qty:            1,
		Type:           manufacturing.RecipeTypeManufacture,
		Version:        1,
		Active:         true,
	}
}

func sampleRecipeLine() *manufacturing.RecipeLine {
	return &manufacturing.RecipeLine{
		Base:        model.Base{ID: 1},
		RecipeID:    1,
		ComponentID: 300,
		Qty:         2,
	}
}

func uint64Ptr(value uint64) *uint64 {
	return &value
}

func bomVariantMock() products.ItemVariantDAOMock {
	return products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, id uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: id}, ItemID: 1}, nil
			},
		},
	}
}

func bomServiceFor(recipes manufacturing.RecipeDAO, lines manufacturing.RecipeLineDAO) manufacturing.RecipeService {
	return manufacturing.NewRecipeService(bomVariantMock(), recipes, lines)
}

func TestRecipeHandler_List_ReturnsRecipes(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[manufacturing.Recipe], error) {
				return &query.Page[manufacturing.Recipe]{Items: []*manufacturing.Recipe{sampleRecipe()}, Count: 1}, nil
			},
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/recipes/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRecipeHandler_List_RejectsInvalidQuery(t *testing.T) {
	app := bomTestApp(t, bomServiceFor(manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/recipes/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{}
	app := fiber.New()
	h := NewRecipeHandler(bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/recipes/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestRecipeHandler_List_ReturnsServerError(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[manufacturing.Recipe], error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/recipes/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRecipeHandler_List_ExportsCSV(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[manufacturing.Recipe], error) {
				return &query.Page[manufacturing.Recipe]{Items: []*manufacturing.Recipe{sampleRecipe()}, Count: 1}, nil
			},
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/recipes/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRecipeHandler_Get_ReturnsRecipe(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/recipes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRecipeHandler_Get_ReturnsNotFound(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return nil, nil
			},
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/recipes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRecipeHandler_Get_RejectsForeignOrganization(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return &manufacturing.Recipe{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(99), ItemID: 200, Type: manufacturing.RecipeTypeManufacture}, nil
			},
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/recipes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRecipeHandler_Get_RejectsInvalidID(t *testing.T) {
	app := bomTestApp(t, bomServiceFor(manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/recipes/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Get_ReturnsServerError(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/recipes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRecipeHandler_Create_CreatesRecipe(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{},
		CreateWithLinesFunc: func(_ context.Context, recipe *manufacturing.Recipe, lines []*manufacturing.RecipeLine) (*manufacturing.Recipe, error) {
			recipe.ID = 1
			return recipe, nil
		},
	}
	lines := manufacturing.RecipeLineDAOMock{}
	svc := bomServiceFor(recipes, lines)
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":"manufacture","qty":1,"lines":[{"component_id":300,"qty":2,"scrap_pct":10}]}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestRecipeHandler_Create_RejectsValidation(t *testing.T) {
	svc := bomServiceFor(manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":""}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Create_MapsInvalidType(t *testing.T) {
	svc := bomServiceFor(manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":"bogus","lines":[{"component_id":300,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Create_MapsNoLines(t *testing.T) {
	svc := bomServiceFor(manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":"manufacture"}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Create_MapsUnknownProduct(t *testing.T) {
	variants := products.ItemVariantDAOMock{}
	svc := manufacturing.NewRecipeService(variants, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":"manufacture","lines":[{"component_id":300,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Create_ReturnsServerError(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{},
		CreateWithLinesFunc: func(_ context.Context, recipe *manufacturing.Recipe, lines []*manufacturing.RecipeLine) (*manufacturing.Recipe, error) {
			return nil, errors.New("db down")
		},
	}
	lines := manufacturing.RecipeLineDAOMock{}
	svc := bomServiceFor(recipes, lines)
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":"manufacture","qty":1,"lines":[{"component_id":300,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRecipeHandler_Update_UpdatesRecipe(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
			UpdateFunc: func(_ context.Context, recipe *manufacturing.Recipe) (*manufacturing.Recipe, error) {
				return recipe, nil
			},
		},
	}
	lines := manufacturing.RecipeLineDAOMock{}
	svc := bomServiceFor(recipes, lines)
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":"manufacture","qty":2,"version":2}`
	resp, err := doRequest(app, http.MethodPut, "/recipes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRecipeHandler_Update_RejectsInvalidID(t *testing.T) {
	svc := bomServiceFor(manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":"manufacture"}`
	resp, err := doRequest(app, http.MethodPut, "/recipes/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Update_ReturnsNotFound(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return nil, nil
			},
		},
	}
	svc := bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":"manufacture"}`
	resp, err := doRequest(app, http.MethodPut, "/recipes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRecipeHandler_Update_RejectsValidation(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
		},
	}
	svc := bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":""}`
	resp, err := doRequest(app, http.MethodPut, "/recipes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Update_ReturnsServerError(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
			UpdateFunc: func(_ context.Context, recipe *manufacturing.Recipe) (*manufacturing.Recipe, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":"manufacture"}`
	resp, err := doRequest(app, http.MethodPut, "/recipes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRecipeHandler_Delete_DeletesRecipe(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodDelete, "/recipes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestRecipeHandler_Delete_RejectsInvalidID(t *testing.T) {
	app := bomTestApp(t, bomServiceFor(manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodDelete, "/recipes/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Delete_ReturnsNotFound(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return nil, nil
			},
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodDelete, "/recipes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRecipeHandler_Delete_ReturnsServerError(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return errors.New("db down")
			},
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodDelete, "/recipes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRecipeHandler_ListLines_ReturnsLines(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
		},
	}
	lines := manufacturing.RecipeLineDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*manufacturing.RecipeLine, error) {
			return []*manufacturing.RecipeLine{sampleRecipeLine()}, nil
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, lines))

	resp, err := doRequest(app, http.MethodGet, "/recipes/1/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRecipeHandler_ListLines_ReturnsNotFound(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return nil, nil
			},
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/recipes/1/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRecipeHandler_ListLines_RejectsInvalidID(t *testing.T) {
	app := bomTestApp(t, bomServiceFor(manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/recipes/abc/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_ListLines_ReturnsServerError(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
		},
	}
	lines := manufacturing.RecipeLineDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*manufacturing.RecipeLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := bomTestApp(t, bomServiceFor(recipes, lines))

	resp, err := doRequest(app, http.MethodGet, "/recipes/1/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRecipeHandler_Explode_ExplodesRecipe(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
		},
	}
	lines := manufacturing.RecipeLineDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*manufacturing.RecipeLine, error) {
			return []*manufacturing.RecipeLine{sampleRecipeLine()}, nil
		},
	}
	svc := bomServiceFor(recipes, lines)
	app := bomTestApp(t, svc)

	body := `{"qty":"4"}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/1/explode", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRecipeHandler_Explode_RejectsInvalidID(t *testing.T) {
	svc := bomServiceFor(manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"qty":"4"}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/abc/explode", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Explode_RejectsValidation(t *testing.T) {
	svc := bomServiceFor(manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/1/explode", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Explode_ReturnsNotFound(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return nil, nil
			},
		},
	}
	svc := bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"qty":"4"}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/1/explode", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRecipeHandler_Explode_MapsNonPositiveQuantity(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
		},
	}
	lines := manufacturing.RecipeLineDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*manufacturing.RecipeLine, error) {
			return []*manufacturing.RecipeLine{sampleRecipeLine()}, nil
		},
	}
	svc := bomServiceFor(recipes, lines)
	app := bomTestApp(t, svc)

	body := `{"qty":"0"}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/1/explode", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Explode_MapsInvalidDecimal(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
		},
	}
	svc := bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"qty":"abc"}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/1/explode", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Explode_ReturnsServerError(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
		},
	}
	lines := manufacturing.RecipeLineDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*manufacturing.RecipeLine, error) {
			return nil, errors.New("db down")
		},
	}
	svc := bomServiceFor(recipes, lines)
	app := bomTestApp(t, svc)

	body := `{"qty":"4"}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/1/explode", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRecipeHandler_Explode_MapsNoLines(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return sampleRecipe(), nil
			},
		},
	}
	svc := bomServiceFor(recipes, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"qty":"4"}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/1/explode", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Create_MapsInvalidLineQuantity(t *testing.T) {
	svc := bomServiceFor(manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":"manufacture","lines":[{"component_id":300,"qty":0}]}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRecipeHandler_Create_MapsUnknownComponent(t *testing.T) {
	variants := products.ItemVariantDAOMock{}
	svc := manufacturing.NewRecipeService(variants, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{})
	app := bomTestApp(t, svc)

	body := `{"item_id":200,"type":"manufacture","lines":[{"component_id":300,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/recipes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func passthroughGuards() httpx.RouteGuards {
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

func doRequest(app *fiber.App, method, path, body string) (*http.Response, error) {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	return app.Test(req)
}

func referenceLocation(locationID uint64, usage string, organizationID uint64) *reference.StockLocation {
	return &reference.StockLocation{Base: model.Base{ID: locationID}, OrganizationID: uint64Ptr(organizationID), Usage: usage}
}
