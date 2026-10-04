package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/asset"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestAssetHandler_RegisterAsset_RejectsValidation(t *testing.T) {
	app := assetApp(emptyHandler(), passthroughGuards())

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_RegisterAsset_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	app := assetApp(emptyHandler(), noTenantGuards())

	body := `{"name":"Laptop","category_id":1,"purchase_value":10000000,"salvage_value":1000000,"acquisition_date":"2026-01-01","in_service_date":"2026-01-01","invoice_line_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_RegisterAsset_RejectsInvalidAcquisitionDate(t *testing.T) {
	app := assetApp(emptyHandler(), passthroughGuards())

	body := `{"name":"Laptop","category_id":1,"purchase_value":10000000,"salvage_value":1000000,"acquisition_date":"not-a-date","in_service_date":"2026-01-01","invoice_line_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_RegisterAsset_RejectsInvalidInServiceDate(t *testing.T) {
	app := assetApp(emptyHandler(), passthroughGuards())

	body := `{"name":"Laptop","category_id":1,"purchase_value":10000000,"salvage_value":1000000,"acquisition_date":"2026-01-01","in_service_date":"not-a-date","invoice_line_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_RegisterAsset_ReturnsNotFoundForMissingLine(t *testing.T) {
	svc := handlerAssetSvc(
		asset.FixedAssetDAOMock{},
		asset.AssetDepreciationLineDAOMock{},
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return assetCategoryWithTenant(), nil
		}},
		handlerInvoiceLineLookup{find: func(_ context.Context, _ uint64) (*accounting.InvoiceLine, error) {
			return nil, nil
		}},
		handlerInvoiceLookup{},
	)
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"name":"Laptop","category_id":1,"purchase_value":10000000,"salvage_value":1000000,"acquisition_date":"2026-01-01","in_service_date":"2026-01-01","invoice_line_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAssetHandler_RegisterAsset_RejectsInvalidValues(t *testing.T) {
	moveID := uint64(30)
	svc := handlerAssetSvc(
		asset.FixedAssetDAOMock{},
		asset.AssetDepreciationLineDAOMock{},
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return assetCategoryWithTenant(), nil
		}},
		handlerInvoiceLineLookup{find: func(_ context.Context, _ uint64) (*accounting.InvoiceLine, error) {
			return &accounting.InvoiceLine{Base: model.Base{ID: 1}, InvoiceID: 9}, nil
		}},
		handlerInvoiceLookup{find: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 9}, Type: accounting.InvoiceTypeSupplierBill, State: accounting.InvoiceStatePosted, EntryID: &moveID}, nil
		}},
	)
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"name":"Laptop","category_id":1,"purchase_value":10000000,"salvage_value":10000000,"acquisition_date":"2026-01-01","in_service_date":"2026-01-01","invoice_line_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_RegisterAsset_ReturnsServerError(t *testing.T) {
	svc := handlerAssetSvc(
		asset.FixedAssetDAOMock{},
		asset.AssetDepreciationLineDAOMock{},
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return nil, errors.New("db down")
		}},
		handlerInvoiceLineLookup{},
		handlerInvoiceLookup{},
	)
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"name":"Laptop","category_id":1,"purchase_value":10000000,"salvage_value":1000000,"acquisition_date":"2026-01-01","in_service_date":"2026-01-01","invoice_line_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAssetHandler_GenerateSchedule_ReturnsNotFound(t *testing.T) {
	assets := asset.FixedAssetDAOMock{}
	svc := handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{})
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/schedule", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAssetHandler_GenerateSchedule_ReturnsNotFoundForForeignOrganization(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 99, State: asset.AssetStateRunning}, nil
			},
		},
	}
	svc := handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{})
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/schedule", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAssetHandler_GenerateSchedule_ReturnsServerError(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{})
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/schedule", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAssetHandler_GenerateSchedule_RejectsNonRunningAsset(t *testing.T) {
	inService := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Server", CategoryID: 1, PurchaseValue: 10000000, SalvageValue: 1000000, State: asset.AssetStateDraft, InServiceDate: &inService}, nil
			},
		},
	}
	svc := handlerAssetSvc(
		assets,
		asset.AssetDepreciationLineDAOMock{},
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return assetCategoryWithTenant(), nil
		}},
		handlerInvoiceLineLookup{},
		handlerInvoiceLookup{},
	)
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/schedule", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_GenerateSchedule_ReturnsServerErrorOnWrite(t *testing.T) {
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
		asset.AssetDepreciationLineDAOMock{},
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return nil, errors.New("db down")
		}},
		handlerInvoiceLineLookup{},
		handlerInvoiceLookup{},
	)
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/schedule", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAssetHandler_PostDepreciation_PostsDepreciation(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Server", CategoryID: 1, PurchaseValue: 10000000, SalvageValue: 1000000, State: asset.AssetStateRunning}, nil
			},
		},
	}
	lines := asset.AssetDepreciationLineDAOMock{
		ListByAssetFunc: func(_ context.Context, _ uint64) ([]*asset.AssetDepreciationLine, error) {
			return []*asset.AssetDepreciationLine{{Base: model.Base{ID: 1}, Sequence: 1, DepreciationDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Amount: 100}}, nil
		},
	}
	svc := handlerAssetSvc(
		assets,
		lines,
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return assetCategoryWithTenant(), nil
		}},
		handlerInvoiceLineLookup{},
		handlerInvoiceLookup{},
	)
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"journal_id":9,"date":"2026-02-28"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/post-depreciation", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestAssetHandler_PostDepreciation_RejectsValidation(t *testing.T) {
	app := assetApp(emptyHandler(), passthroughGuards())

	body := `{"journal_id":0,"date":""}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/post-depreciation", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_PostDepreciation_RejectsInvalidDate(t *testing.T) {
	app := assetApp(emptyHandler(), passthroughGuards())

	body := `{"journal_id":9,"date":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/post-depreciation", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_PostDepreciation_ReturnsNotFound(t *testing.T) {
	assets := asset.FixedAssetDAOMock{}
	svc := handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{})
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"journal_id":9,"date":"2026-02-28"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/post-depreciation", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAssetHandler_PostDepreciation_ReturnsServerError(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{})
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"journal_id":9,"date":"2026-02-28"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/post-depreciation", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAssetHandler_PostDepreciation_ReturnsUnprocessableWhenNothingToPost(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Server", CategoryID: 1, PurchaseValue: 10000000, SalvageValue: 1000000, State: asset.AssetStateRunning}, nil
			},
		},
	}
	lines := asset.AssetDepreciationLineDAOMock{
		ListByAssetFunc: func(_ context.Context, _ uint64) ([]*asset.AssetDepreciationLine, error) {
			return []*asset.AssetDepreciationLine{}, nil
		},
	}
	svc := handlerAssetSvc(
		assets,
		lines,
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return assetCategoryWithTenant(), nil
		}},
		handlerInvoiceLineLookup{},
		handlerInvoiceLookup{},
	)
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"journal_id":9,"date":"2026-02-28"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/post-depreciation", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_PostDepreciation_ReturnsServerErrorOnWrite(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Server", CategoryID: 1, PurchaseValue: 10000000, SalvageValue: 1000000, State: asset.AssetStateRunning}, nil
			},
		},
	}
	svc := handlerAssetSvc(
		assets,
		asset.AssetDepreciationLineDAOMock{},
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return nil, errors.New("db down")
		}},
		handlerInvoiceLineLookup{},
		handlerInvoiceLookup{},
	)
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"journal_id":9,"date":"2026-02-28"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/post-depreciation", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAssetHandler_DisposeAsset_RejectsInvalidDate(t *testing.T) {
	app := assetApp(emptyHandler(), passthroughGuards())

	body := `{"journal_id":9,"date":"not-a-date","state":"disposed"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/dispose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_DisposeAsset_ReturnsNotFound(t *testing.T) {
	assets := asset.FixedAssetDAOMock{}
	svc := handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{})
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"journal_id":9,"date":"2026-08-31","state":"disposed"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/dispose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAssetHandler_DisposeAsset_ReturnsServerError(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{})
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"journal_id":9,"date":"2026-08-31","state":"disposed"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/dispose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAssetHandler_DisposeAsset_RejectsInvalidStateTransition(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Server", CategoryID: 1, PurchaseValue: 10000000, SalvageValue: 1000000, State: asset.AssetStateDraft}, nil
			},
		},
	}
	svc := handlerAssetSvc(assets, asset.AssetDepreciationLineDAOMock{}, handlerAssetCategoryLookup{}, handlerInvoiceLineLookup{}, handlerInvoiceLookup{})
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"journal_id":9,"date":"2026-08-31","state":"disposed"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/dispose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAssetHandler_DisposeAsset_ReturnsServerErrorOnWrite(t *testing.T) {
	assets := asset.FixedAssetDAOMock{
		CRUDMock: dao.CRUDMock[asset.FixedAsset]{
			FindFunc: func(_ context.Context, _ uint64) (*asset.FixedAsset, error) {
				return &asset.FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Server", CategoryID: 1, PurchaseValue: 10000000, SalvageValue: 1000000, State: asset.AssetStateRunning}, nil
			},
		},
	}
	svc := handlerAssetSvc(
		assets,
		asset.AssetDepreciationLineDAOMock{},
		handlerAssetCategoryLookup{find: func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
			return nil, errors.New("db down")
		}},
		handlerInvoiceLineLookup{},
		handlerInvoiceLookup{},
	)
	h := NewAssetCategoryHandler(dao.CRUDMock[reference.AssetCategory]{}, svc)
	app := assetApp(h, passthroughGuards())

	body := `{"journal_id":9,"date":"2026-08-31","state":"disposed"}`
	resp, err := doRequest(app, http.MethodPost, "/fixed-assets/1/dispose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func writeErrorStatus(path string, fn func(c fiber.Ctx) error) int {
	app := fiber.New()
	app.Post(path, func(c fiber.Ctx) error {
		return fn(c)
	})
	resp, err := doRequest(app, http.MethodPost, path, "")
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

func TestWriteAssetError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "asset not found", err: asset.ErrAssetNotFound, status: http.StatusNotFound},
		{name: "category not found", err: asset.ErrAssetCategoryNotFound, status: http.StatusNotFound},
		{name: "invoice line not found", err: asset.ErrAssetInvoiceLineNotFound, status: http.StatusNotFound},
		{name: "invoice not supplier bill", err: asset.ErrAssetInvoiceNotSupplierBill, status: http.StatusNotFound},
		{name: "name required", err: asset.ErrAssetNameRequired, status: http.StatusUnprocessableEntity},
		{name: "invalid state", err: asset.ErrAssetInvalidState, status: http.StatusUnprocessableEntity},
		{name: "invalid method", err: asset.ErrAssetInvalidMethod, status: http.StatusUnprocessableEntity},
		{name: "invalid periods", err: asset.ErrAssetInvalidPeriods, status: http.StatusUnprocessableEntity},
		{name: "invalid dates", err: asset.ErrAssetInvalidDates, status: http.StatusUnprocessableEntity},
		{name: "invalid values", err: asset.ErrAssetInvalidValues, status: http.StatusUnprocessableEntity},
		{name: "category accounts", err: asset.ErrAssetCategoryAccounts, status: http.StatusUnprocessableEntity},
		{name: "not running", err: asset.ErrAssetNotRunning, status: http.StatusUnprocessableEntity},
		{name: "schedule exists", err: asset.ErrAssetScheduleExists, status: http.StatusUnprocessableEntity},
		{name: "no schedule", err: asset.ErrAssetNoSchedule, status: http.StatusUnprocessableEntity},
		{name: "line not found", err: asset.ErrAssetLineNotFound, status: http.StatusUnprocessableEntity},
		{name: "nothing to post", err: asset.ErrAssetNothingToPost, status: http.StatusUnprocessableEntity},
		{name: "line posted", err: asset.ErrAssetLinePosted, status: http.StatusUnprocessableEntity},
		{name: "disposal proceeds", err: asset.ErrAssetDisposalProceeds, status: http.StatusUnprocessableEntity},
		{name: "unknown", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeAssetError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
