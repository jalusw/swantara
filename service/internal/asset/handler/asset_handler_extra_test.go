package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/asset"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func noTenantGuards() httpx.RouteGuards {
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

func assetApp(h AssetCategoryHandler, guards httpx.RouteGuards) *fiber.App {
	app := fiber.New()
	h.Register(app, guards)
	return app
}

func assetCategoryWithTenant() *reference.AssetCategory {
	assetAccount := uint64(1500)
	depreciationAccount := uint64(1510)
	expenseAccount := uint64(6600)
	gainAccount := uint64(4500)
	lossAccount := uint64(6000)
	method := "linear"
	period := "month"
	periods := 5
	organizationID := uint64(1)
	return &reference.AssetCategory{
		Base:                  model.Base{ID: 1},
		OrganizationID:        &organizationID,
		Name:                  "Computer",
		AssetAccountID:        &assetAccount,
		DepreciationAccountID: &depreciationAccount,
		ExpenseAccountID:      &expenseAccount,
		GainAccountID:         &gainAccount,
		LossAccountID:         &lossAccount,
		Method:                &method,
		MethodNumber:          &periods,
		MethodPeriod:          &period,
	}
}

func assetCategoryCRUDWithList(category *reference.AssetCategory) dao.CRUDMock[reference.AssetCategory] {
	return dao.CRUDMock[reference.AssetCategory]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.AssetCategory], error) {
			return &query.Page[reference.AssetCategory]{Items: []*reference.AssetCategory{category}, Count: 1}, nil
		},
		FindFunc: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return category, nil
		},
		CreateFunc: func(_ context.Context, c *reference.AssetCategory) (*reference.AssetCategory, error) {
			return c, nil
		},
	}
}

func emptyHandler() AssetCategoryHandler {
	return NewAssetCategoryHandler(asset.NewAssetCategoryService(dao.CRUDMock[reference.AssetCategory]{}), handlerAssetSvc(asset.FixedAssetDAOMock{}, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
}

func TestAssetHandler_ListCategories_ReturnsTenantScopedCategories(t *testing.T) {
	category := assetCategoryWithTenant()
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(assetCategoryCRUDWithList(category)), handlerAssetSvc(asset.FixedAssetDAOMock{}, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/asset-categories/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAssetHandler_ListCategories_RejectsInvalidQuery(t *testing.T) {
	app := assetApp(emptyHandler(), passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/asset-categories/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_ListCategories_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	app := assetApp(emptyHandler(), noTenantGuards())

	resp, err := doRequest(app, http.MethodGet, "/asset-categories/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestAssetHandler_ListCategories_ReturnsServerError(t *testing.T) {
	categories := dao.CRUDMock[reference.AssetCategory]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.AssetCategory], error) {
			return nil, errors.New("db down")
		},
	}
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(categories), handlerAssetSvc(asset.FixedAssetDAOMock{}, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/asset-categories/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAssetHandler_GetCategory_ReturnsCategory(t *testing.T) {
	category := assetCategoryWithTenant()
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(assetCategoryCRUDWithList(category)), handlerAssetSvc(asset.FixedAssetDAOMock{}, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/asset-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAssetHandler_GetCategory_ReturnsNotFoundForForeignOrganization(t *testing.T) {
	category := assetCategoryWithTenant()
	other := uint64(99)
	category.OrganizationID = &other
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(assetCategoryCRUDWithList(category)), handlerAssetSvc(asset.FixedAssetDAOMock{}, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/asset-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAssetHandler_GetCategory_ReturnsServerError(t *testing.T) {
	categories := dao.CRUDMock[reference.AssetCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return nil, errors.New("db down")
		},
	}
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(categories), handlerAssetSvc(asset.FixedAssetDAOMock{}, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/asset-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAssetHandler_CreateCategory_CreatesCategory(t *testing.T) {
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(assetCategoryCRUDWithList(assetCategoryWithTenant())), handlerAssetSvc(asset.FixedAssetDAOMock{}, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	body := `{"name":"Computer","asset_account_id":1500,"depreciation_account_id":1510,"expense_account_id":6600,"gain_account_id":4500,"loss_account_id":6000,"method":"linear","method_number":5,"method_period":"month"}`
	resp, err := doRequest(app, http.MethodPost, "/asset-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestAssetHandler_CreateCategory_RejectsValidation(t *testing.T) {
	app := assetApp(emptyHandler(), passthroughGuards())

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/asset-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_CreateCategory_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	app := assetApp(emptyHandler(), noTenantGuards())

	body := `{"name":"Computer","method":"linear","method_number":5,"method_period":"month"}`
	resp, err := doRequest(app, http.MethodPost, "/asset-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_CreateCategory_ReturnsServerError(t *testing.T) {
	categories := dao.CRUDMock[reference.AssetCategory]{
		CreateFunc: func(_ context.Context, _ *reference.AssetCategory) (*reference.AssetCategory, error) {
			return nil, errors.New("db down")
		},
	}
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(categories), handlerAssetSvc(asset.FixedAssetDAOMock{}, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	body := `{"name":"Computer","method":"linear","method_number":5,"method_period":"month"}`
	resp, err := doRequest(app, http.MethodPost, "/asset-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAssetHandler_ListAssets_ReturnsAssets(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[asset.FixedAsset], error) {
				return &query.Page[asset.FixedAsset]{Items: []*asset.FixedAsset{{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Laptop"}}, Count: 1}, nil
			},
		},
	}
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(dao.CRUDMock[reference.AssetCategory]{}), handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/fixed-assets/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAssetHandler_ListAssets_RejectsInvalidQuery(t *testing.T) {
	app := assetApp(emptyHandler(), passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/fixed-assets/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_ListAssets_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	app := assetApp(emptyHandler(), noTenantGuards())

	resp, err := doRequest(app, http.MethodGet, "/fixed-assets/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestAssetHandler_ListAssets_ReturnsServerError(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[asset.FixedAsset], error) {
				return nil, errors.New("db down")
			},
		},
	}
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(dao.CRUDMock[reference.AssetCategory]{}), handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/fixed-assets/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAssetHandler_GetAsset_ReturnsAsset(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Laptop", State: asset.AssetStateRunning}, nil
			},
		},
	}
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(dao.CRUDMock[reference.AssetCategory]{}), handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/fixed-assets/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAssetHandler_GetAsset_ReturnsNotFoundForForeignOrganization(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 99, Name: "Laptop", State: asset.AssetStateRunning}, nil
			},
		},
	}
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(dao.CRUDMock[reference.AssetCategory]{}), handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/fixed-assets/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAssetHandler_GetAsset_ReturnsServerError(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return nil, errors.New("db down")
			},
		},
	}
	h := NewAssetCategoryHandler(asset.NewAssetCategoryService(dao.CRUDMock[reference.AssetCategory]{}), handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/fixed-assets/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
