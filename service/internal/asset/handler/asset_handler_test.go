package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/asset"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			c.Locals(httpx.LocalOrganizationID, uint64(1))
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

type handlerAssetCategoryLookup struct {
	find func(ctx context.Context, id uint64) (*reference.AssetCategory, error)
}

func (m handlerAssetCategoryLookup) Find(ctx context.Context, id uint64) (*reference.AssetCategory, error) {
	if m.find != nil {
		return m.find(ctx, id)
	}
	return nil, nil
}

type handlerInvoiceLineLookup struct {
	find func(ctx context.Context, id uint64) (*accounting.InvoiceLine, error)
}

func (m handlerInvoiceLineLookup) Find(ctx context.Context, id uint64) (*accounting.InvoiceLine, error) {
	if m.find != nil {
		return m.find(ctx, id)
	}
	return nil, nil
}

type handlerInvoiceLookup struct {
	find func(ctx context.Context, id uint64) (*accounting.Invoice, error)
}

func (m handlerInvoiceLookup) Find(ctx context.Context, id uint64) (*accounting.Invoice, error) {
	if m.find != nil {
		return m.find(ctx, id)
	}
	return nil, nil
}

func handlerAssetSvc(
	assets asset.FixedAssetDAOMock,
	lines asset.AssetDepreciationLineDAOMock,
	categories handlerAssetCategoryLookup,
	invoiceLines handlerInvoiceLineLookup,
	invoices handlerInvoiceLookup,
) asset.AssetService {
	return asset.NewAssetService(
		assets,
		lines,
		categories,
		invoiceLines,
		invoices,
		inventory.PosterMock{},
		inventory.TransactionerMock{},
	)
}

func assetCategoryCRUD(categories *reference.AssetCategory) dao.CRUDMock[reference.AssetCategory] {
	return dao.CRUDMock[reference.AssetCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return categories, nil
		},
		CreateFunc: func(_ context.Context, c *reference.AssetCategory) (*reference.AssetCategory, error) {
			return c, nil
		},
	}
}

func TestAssetHandler_ListCategories_ReturnsCategories(t *testing.T) {
	categories := &reference.AssetCategory{Base: model.Base{ID: 1}, Name: "Computer"}
	handler := NewAssetCategoryHandler(assetCategoryCRUD(categories), handlerAssetSvc(asset.FixedAssetDAOMock{}, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := fiber.New()
	handler.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/asset-categories", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAssetHandler_GetCategory_ReturnsNotFound(t *testing.T) {
	categories := dao.CRUDMock[reference.AssetCategory]{}
	handler := NewAssetCategoryHandler(categories, handlerAssetSvc(asset.FixedAssetDAOMock{}, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := fiber.New()
	handler.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/asset-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAssetHandler_RegisterAsset_RegistersAsset(t *testing.T) {
	assetAccount := uint64(1500)
	depreciationAccount := uint64(1510)
	expenseAccount := uint64(6600)
	gainAccount := uint64(4500)
	lossAccount := uint64(6000)
	method := "linear"
	period := "month"
	periods := 36
	categories := &reference.AssetCategory{
		Base:                  model.Base{ID: 1},
		AssetAccountID:        &assetAccount,
		DepreciationAccountID: &depreciationAccount,
		ExpenseAccountID:      &expenseAccount,
		GainAccountID:         &gainAccount,
		LossAccountID:         &lossAccount,
		Method:                &method,
		MethodNumber:          &periods,
		MethodPeriod:          &period,
	}
	moveID := uint64(30)
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			CreateFunc: func(_ context.Context, a *asset.FixedAsset) (*asset.FixedAsset, error) {
				return a, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Laptop", CategoryID: 1, PurchaseValue: 10000000, SalvageValue: 1000000, State: asset.AssetStateRunning}, nil
			},
		},
	}
	svc := handlerAssetSvc(
		assets,
		asset.AssetDepreciationLineDAOMock{},
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return categories, nil
		}},
		handlerInvoiceLineLookup{find: func(_ context.Context, _ uint64) (*accounting.InvoiceLine, error) {
			return &accounting.InvoiceLine{Base: model.Base{ID: 1}, InvoiceID: 9}, nil
		}},
		handlerInvoiceLookup{find: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 9}, Type: accounting.InvoiceTypeSupplierBill, State: accounting.InvoiceStatePosted, EntryID: &moveID}, nil
		}},
	)
	handler := NewAssetCategoryHandler(assetCategoryCRUD(categories), svc)
	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"name":"Laptop","category_id":1,"purchase_value":10000000,"salvage_value":1000000,"acquisition_date":"2026-01-01","in_service_date":"2026-01-01","invoice_line_id":1,"organization_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestAssetHandler_GetAsset_ReturnsNotFound(t *testing.T) {
	assets := asset.FixedAssetDAOMock{}
	handler := NewAssetCategoryHandler(assetCategoryCRUD(nil), handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{}))
	app := fiber.New()
	handler.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/fixed-assets/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAssetHandler_GenerateSchedule_GeneratesSchedule(t *testing.T) {
	assetAccount := uint64(1500)
	depreciationAccount := uint64(1510)
	expenseAccount := uint64(6600)
	gainAccount := uint64(4500)
	lossAccount := uint64(6000)
	method := "linear"
	period := "month"
	periods := 5
	categories := &reference.AssetCategory{
		Base:                  model.Base{ID: 1},
		AssetAccountID:        &assetAccount,
		DepreciationAccountID: &depreciationAccount,
		ExpenseAccountID:      &expenseAccount,
		GainAccountID:         &gainAccount,
		LossAccountID:         &lossAccount,
		Method:                &method,
		MethodNumber:          &periods,
		MethodPeriod:          &period,
	}
	inService := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Server", CategoryID: 1, PurchaseValue: 10000000, SalvageValue: 1000000, State: asset.AssetStateRunning, InServiceDate: &inService}, nil
			},
		},
	}
	svc := handlerAssetSvc(
		assets,
		asset.AssetDepreciationLineDAOMock{
			ListByAssetFunc: func(_ context.Context, _ uint64) ([]*asset.AssetDepreciationLine, error) {
				return []*asset.AssetDepreciationLine{}, nil
			},
		},
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return categories, nil
		}},
		handlerInvoiceLineLookup{},
		handlerInvoiceLookup{},
	)
	handler := NewAssetCategoryHandler(assetCategoryCRUD(categories), svc)
	app := fiber.New()
	handler.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/schedule", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestAssetHandler_DisposeAsset_DisposesAsset(t *testing.T) {
	assetAccount := uint64(1500)
	depreciationAccount := uint64(1510)
	expenseAccount := uint64(6600)
	gainAccount := uint64(4500)
	lossAccount := uint64(6000)
	method := "linear"
	period := "month"
	periods := 5
	categories := &reference.AssetCategory{
		Base:                  model.Base{ID: 1},
		AssetAccountID:        &assetAccount,
		DepreciationAccountID: &depreciationAccount,
		ExpenseAccountID:      &expenseAccount,
		GainAccountID:         &gainAccount,
		LossAccountID:         &lossAccount,
		Method:                &method,
		MethodNumber:          &periods,
		MethodPeriod:          &period,
	}
	assetRecord := &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Server", CategoryID: 1, PurchaseValue: 10000000, SalvageValue: 1000000, State: asset.AssetStateRunning}
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return assetRecord, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, updated *asset.FixedAsset) (*asset.FixedAsset, error) {
			*assetRecord = *updated
			return updated, nil
		},
	}
	svc := handlerAssetSvc(
		assets,
		asset.AssetDepreciationLineDAOMock{
			ListByAssetFunc: func(_ context.Context, _ uint64) ([]*asset.AssetDepreciationLine, error) {
				return []*asset.AssetDepreciationLine{
					{Base: model.Base{ID: 1}, Amount: 1800000, Posted: true},
				}, nil
			},
		},
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return categories, nil
		}},
		handlerInvoiceLineLookup{},
		handlerInvoiceLookup{},
	)
	handler := NewAssetCategoryHandler(assetCategoryCRUD(categories), svc)
	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"journal_id":9,"date":"2026-08-31","state":"disposed"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/dispose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAssetHandler_DisposeAsset_RejectsUnknownState(t *testing.T) {
	assets := asset.FixedAssetDAOMock{}
	svc := handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{})
	handler := NewAssetCategoryHandler(assetCategoryCRUD(nil), svc)
	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"journal_id":9,"date":"2026-08-31","state":"nonsense"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/dispose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}
