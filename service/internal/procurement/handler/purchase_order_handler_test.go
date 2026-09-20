package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func orderTestApp(t *testing.T, orders procurement.PurchaseOrderDAOMock, lines procurement.PurchaseOrderLineDAOMock, svc procurement.PurchaseOrderService) *fiber.App {
	t.Helper()
	return procurementTestApp(t, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewPurchaseOrderHandler(svc)
		h.Register(api, guards)
	})
}

func orderTestSvc(orders procurement.PurchaseOrderDAOMock, lines procurement.PurchaseOrderLineDAOMock) procurement.PurchaseOrderService {
	return procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{Orders: orders, Lines: lines})
}

func TestPurchaseOrderHandler_List_ReturnsOrders(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.PurchaseOrder], error) {
				return &query.Page[procurement.PurchaseOrder]{Items: []*procurement.PurchaseOrder{samplePurchaseOrder()}, Count: 1}, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_List_RejectsInvalidQuery(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-orders/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_List_ReturnsUnauthorizedWhenTenantMissing(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.PurchaseOrder], error) {
				return &query.Page[procurement.PurchaseOrder]{Items: []*procurement.PurchaseOrder{}, Count: 0}, nil
			},
		},
	}
	svc := orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{})
	app := procurementTestAppNoTenant(t, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewPurchaseOrderHandler(svc)
		h.Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/purchase-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_List_ReturnsServerError(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.PurchaseOrder], error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Get_ReturnsOrder(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
		},
	}
	lines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{samplePurchaseOrderLine()}, nil
		},
	}
	app := orderTestApp(t, orders, lines, orderTestSvc(orders, lines))

	resp, err := doRequest(app, http.MethodGet, "/purchase-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Get_ReturnsNotFound(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return nil, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Get_RejectsForeignTenant(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.OrganizationID = ptrUint64(99)
				return order, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Get_RejectsInvalidID(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-orders/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Get_ReturnsServerError(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Get_ReturnsServerErrorOnLines(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
		},
	}
	lines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderTestApp(t, orders, lines, orderTestSvc(orders, lines))

	resp, err := doRequest(app, http.MethodGet, "/purchase-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_CreatesOrder(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CreateWithLinesFunc: func(_ context.Context, order *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			order.ID = 1
			return order, nil
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_RejectsValidation(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":0,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_RejectsInvalidOrderDate(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"order_date":"bad-date","lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_RejectsInvalidExpectedDate(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"expected_date":"bad-date","lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_MapsNoLines(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			return nil, procurement.ErrPurchaseOrderNoLines
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_MapsVendorNotFound(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			return nil, procurement.ErrPurchaseOrderVendor
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_MapsVendorNotSupplier(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			return nil, procurement.ErrPurchaseOrderVendorNotSupplier
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_MapsWarehouseNotFound(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			return nil, procurement.ErrPurchaseOrderWarehouse
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_MapsNoOffer(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			return nil, procurement.ErrPurchaseOrderNoOffer
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":0}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_MapsInvalidDiscount(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			return nil, procurement.ErrPurchaseOrderLineDiscount
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000,"discount_pct":150}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_ReturnsServerError(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_CreatesFromRequisition(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CreateWithLinesFunc: func(_ context.Context, order *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			order.ID = 1
			return order, nil
		},
	}
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return &procurement.PurchaseRequest{
					Base:           model.Base{ID: 1},
					OrganizationID: ptrUint64(10),
					RequesterID:    5,
					State:          procurement.RequestStateApproved,
				}, nil
			},
		},
	}
	requisitionLines := procurement.PurchaseRequestLineDAOMock{
		ListByRequestFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseRequestLine, error) {
			return []*procurement.PurchaseRequestLine{
				{Base: model.Base{ID: 1}, RequestID: 1, ItemID: ptrUint64(200), Qty: 7},
			}, nil
		},
	}
	offers := procurement.OfferEngineMock{
		BestOfferForSupplierFunc: func(_ context.Context, variantID, supplierID uint64, _ amount.Amount, _ time.Time) (*products.SupplierProduct, error) {
			return &products.SupplierProduct{ItemID: variantID, SupplierID: supplierID, Price: helper.Ptr(500.0)}, nil
		},
	}
	svc := procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders:           orders,
		Requisitions:     requisitions,
		RequisitionLines: requisitionLines,
		Offers:           offers,
	})
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, svc)

	body := `{"request_id":1,"supplier_id":10,"warehouse_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_MapsRequisitionNotFound(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{}
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return nil, nil
			},
		},
	}
	svc := procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders:       orders,
		Requisitions: requisitions,
	})
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, svc)

	body := `{"request_id":1,"supplier_id":10,"warehouse_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Create_MapsRequisitionNotConvertible(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{}
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return &procurement.PurchaseRequest{
					Base:  model.Base{ID: 1},
					State: procurement.RequestStateConfirmed,
				}, nil
			},
		},
	}
	svc := procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders:       orders,
		Requisitions: requisitions,
	})
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, svc)

	body := `{"request_id":1,"supplier_id":10,"warehouse_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_UpdateDraft_UpdatesOrder(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
				return order, nil
			},
		},
	}
	lines := procurement.PurchaseOrderLineDAOMock{}
	app := orderTestApp(t, orders, lines, orderTestSvc(orders, lines))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPut, "/purchase-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_UpdateDraft_RejectsInvalidID(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPut, "/purchase-orders/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_UpdateDraft_RejectsValidation(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":0,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPut, "/purchase-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_UpdateDraft_RejectsInvalidDate(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"order_date":"bad-date","lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPut, "/purchase-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_UpdateDraft_RejectsInvalidExpectedDate(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"order_date":"2026-01-01","expected_date":"bad-date","lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPut, "/purchase-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_UpdateDraft_MapsNotFound(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return nil, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPut, "/purchase-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_UpdateDraft_MapsState(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateSent
				return order, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPut, "/purchase-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_UpdateDraft_MapsNoLines(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[]}`
	resp, err := doRequest(app, http.MethodPut, "/purchase-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_UpdateDraft_ReturnsServerError(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
			UpdateFunc: func(_ context.Context, _ *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	body := `{"supplier_id":10,"warehouse_id":20,"lines":[{"item_id":100,"qty_ordered":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPut, "/purchase-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Confirm_ConfirmsOrder(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
				return order, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Confirm_ConfirmsOrderWithExplicitByUserID(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
				return order, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/confirm", `{"by_user_id":7}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Confirm_RejectsValidation(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/confirm", `{invalid-json}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Confirm_RejectsInvalidID(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/abc/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Confirm_MapsNotFound(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return nil, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Confirm_MapsState(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateConfirmed
				return order, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Confirm_MapsApprovalPending(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.AmountTotal = 1000000
				return order, nil
			},
		},
	}
	svc := procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders:  orders,
		Configs: procurement.NewTestApprovalConfig(10, 5000, []uint64{7}),
	})
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Confirm_MapsApprovalRequired(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.AmountTotal = 1000000
				return order, nil
			},
		},
	}
	svc := procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders:  orders,
		Configs: procurement.NewTestApprovalConfig(10, 5000, nil),
	})
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Confirm_ReturnsServerError(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
			UpdateFunc: func(_ context.Context, _ *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Cancel_CancelsOrder(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
				return order, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Cancel_RejectsInvalidID(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/abc/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Cancel_MapsNotFound(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return nil, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Cancel_MapsState(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateDone
				return order, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Cancel_ReturnsServerError(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
			UpdateFunc: func(_ context.Context, _ *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func receiveTestSvc(orders procurement.PurchaseOrderDAOMock, lines procurement.PurchaseOrderLineDAOMock) procurement.PurchaseOrderService {
	order := samplePurchaseOrder()
	order.State = procurement.PurchaseOrderStateSent
	return procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders: orders,
		Lines:  lines,
		Shipments: inventory.ShipmentDAOMock{
			CRUDMock: dao.CRUDMock[inventory.Shipment]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
					return &query.Page[inventory.Shipment]{}, nil
				},
			},
			CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, _ []*inventory.StockMovement) (*inventory.Shipment, error) {
				shipment.ID = 50
				return shipment, nil
			},
		},
		Locations: inventory.StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
					if q.Filters[0].Field == "usage" {
						return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 30}, Usage: "supplier"}}}, nil
					}
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 40}, OrganizationID: helper.Ptr(uint64(10)), Usage: "internal"}}}, nil
				},
			},
		},
		Movements: inventory.StockMovementDAOMock{
			ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{Base: model.Base{ID: 60}, ItemID: 200, Qty: 2, State: inventory.MovementStateConfirmed}}, nil
			},
		},
	})
}

func TestPurchaseOrderHandler_Receive_ReceivesOrder(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateSent
				return order, nil
			},
			UpdateFunc: func(_ context.Context, order *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
				return order, nil
			},
		},
	}
	lines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{samplePurchaseOrderLine()}, nil
		},
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrderLine]{
			UpdateFunc: func(_ context.Context, line *procurement.PurchaseOrderLine) (*procurement.PurchaseOrderLine, error) {
				return line, nil
			},
		},
	}
	app := orderTestApp(t, orders, lines, receiveTestSvc(orders, lines))

	body := `{"journal_id":90}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Receive_ReceivesOrderWithDate(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateSent
				return order, nil
			},
			UpdateFunc: func(_ context.Context, order *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
				return order, nil
			},
		},
	}
	lines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{samplePurchaseOrderLine()}, nil
		},
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrderLine]{
			UpdateFunc: func(_ context.Context, line *procurement.PurchaseOrderLine) (*procurement.PurchaseOrderLine, error) {
				return line, nil
			},
		},
	}
	app := orderTestApp(t, orders, lines, receiveTestSvc(orders, lines))

	body := `{"journal_id":90,"date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Receive_RejectsInvalidID(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/abc/receive", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Receive_RejectsValidation(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/receive", `{"journal_id":0}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Receive_RejectsInvalidDate(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/receive", `{"journal_id":90,"date":"bad-date"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Receive_MapsNotFound(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return nil, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/receive", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Receive_MapsState(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateDraft
				return order, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/receive", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Receive_MapsNothingToReceive(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateSent
				return order, nil
			},
		},
	}
	lines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			line := samplePurchaseOrderLine()
			line.QtyReceived = 2
			return []*procurement.PurchaseOrderLine{line}, nil
		},
	}
	app := orderTestApp(t, orders, lines, receiveTestSvc(orders, lines))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/receive", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Receive_ReturnsServerError(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateSent
				return order, nil
			},
		},
	}
	lines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderTestApp(t, orders, lines, orderTestSvc(orders, lines))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/receive", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_CreateSupplierBill_CreatesBill(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateConfirmed
				return order, nil
			},
			UpdateFunc: func(_ context.Context, order *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
				return order, nil
			},
		},
	}
	lines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			line := samplePurchaseOrderLine()
			line.QtyReceived = 2
			return []*procurement.PurchaseOrderLine{line}, nil
		},
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrderLine]{
			UpdateFunc: func(_ context.Context, line *procurement.PurchaseOrderLine) (*procurement.PurchaseOrderLine, error) {
				return line, nil
			},
		},
	}
	svc := procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders: orders,
		Lines:  lines,
	})
	app := orderTestApp(t, orders, lines, svc)

	body := `{"journal_id":90}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/supplier-bill", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_CreateSupplierBill_RejectsInvalidID(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/abc/supplier-bill", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_CreateSupplierBill_RejectsValidation(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/supplier-bill", `{"journal_id":0}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_CreateSupplierBill_RejectsInvalidDate(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/supplier-bill", `{"journal_id":90,"date":"bad-date"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_CreateSupplierBill_MapsNotFound(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return nil, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/supplier-bill", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_CreateSupplierBill_MapsState(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateDraft
				return order, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/supplier-bill", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_CreateSupplierBill_MapsNothingToBill(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateConfirmed
				return order, nil
			},
		},
	}
	lines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			line := samplePurchaseOrderLine()
			line.QtyReceived = 2
			line.QtyBilled = 2
			return []*procurement.PurchaseOrderLine{line}, nil
		},
	}
	app := orderTestApp(t, orders, lines, orderTestSvc(orders, lines))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/supplier-bill", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_CreateSupplierBill_MapsQualityBlocked(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateConfirmed
				return order, nil
			},
		},
	}
	lines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			line := samplePurchaseOrderLine()
			line.QtyReceived = 2
			return []*procurement.PurchaseOrderLine{line}, nil
		},
	}
	quality := procurement.QualityEngineMock{
		HasFailedChecksFunc: func(_ context.Context, _ uint64) (bool, error) {
			return true, nil
		},
	}
	svc := procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders: orders,
		Lines:  lines,
		Shipments: inventory.ShipmentDAOMock{
			CRUDMock: dao.CRUDMock[inventory.Shipment]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
					return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 50}, Type: inventory.ShipmentTypeIncoming}}}, nil
				},
			},
		},
		Quality: quality,
	})
	app := orderTestApp(t, orders, lines, svc)

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/supplier-bill", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_CreateSupplierBill_ReturnsServerError(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				order := samplePurchaseOrder()
				order.State = procurement.PurchaseOrderStateConfirmed
				return order, nil
			},
		},
	}
	lines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			line := samplePurchaseOrderLine()
			line.QtyReceived = 2
			return []*procurement.PurchaseOrderLine{line}, nil
		},
	}
	bills := procurement.BillEngineMock{
		CreateSupplierBillFunc: func(_ context.Context, _ accounting.CreateSupplierBillRequest) (*accounting.Invoice, error) {
			return nil, errors.New("db down")
		},
	}
	svc := procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders: orders,
		Lines:  lines,
		Bills:  bills,
	})
	app := orderTestApp(t, orders, lines, svc)

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/supplier-bill", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Pay_PaysBill(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
		},
	}
	openInvoices := procurement.OpenInvoiceLookupMock{
		ListOpenByContactFunc: func(_ context.Context, _ uint64) ([]*accounting.Invoice, error) {
			return []*accounting.Invoice{{Base: model.Base{ID: 500}, AmountResidual: amount.FromFloat64(2000)}}, nil
		},
	}
	payments := procurement.OutboundPaymentEngineMock{
		CreateOutboundFunc: func(_ context.Context, request accounting.CreatePaymentRequest) (*accounting.Payment, error) {
			return &accounting.Payment{Base: model.Base{ID: 900}, ContactID: request.ContactID, Amount: request.Amount}, nil
		},
	}
	svc := procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders:       orders,
		OpenInvoices: openInvoices,
		Payments:     payments,
	})
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, svc)

	body := `{"journal_id":90}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Pay_RejectsInvalidID(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/abc/pay", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Pay_RejectsValidation(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/pay", `{"journal_id":0}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Pay_RejectsInvalidDate(t *testing.T) {
	app := orderTestApp(t, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/pay", `{"journal_id":90,"date":"bad-date"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Pay_MapsNotFound(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return nil, nil
			},
		},
	}
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, orderTestSvc(orders, procurement.PurchaseOrderLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/pay", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Pay_MapsNoOpenBills(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
		},
	}
	openInvoices := procurement.OpenInvoiceLookupMock{
		ListOpenByContactFunc: func(_ context.Context, _ uint64) ([]*accounting.Invoice, error) {
			return []*accounting.Invoice{}, nil
		},
	}
	svc := procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders:       orders,
		OpenInvoices: openInvoices,
	})
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/pay", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_Pay_ReturnsServerError(t *testing.T) {
	orders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return samplePurchaseOrder(), nil
			},
		},
	}
	openInvoices := procurement.OpenInvoiceLookupMock{
		ListOpenByContactFunc: func(_ context.Context, _ uint64) ([]*accounting.Invoice, error) {
			return []*accounting.Invoice{{Base: model.Base{ID: 500}}}, nil
		},
	}
	payments := procurement.OutboundPaymentEngineMock{
		CreateOutboundFunc: func(_ context.Context, _ accounting.CreatePaymentRequest) (*accounting.Payment, error) {
			return nil, errors.New("db down")
		},
	}
	svc := procurement.NewTestPurchaseOrderService(procurement.PurchaseOrderServiceTestDeps{
		Orders:       orders,
		OpenInvoices: openInvoices,
		Payments:     payments,
	})
	app := orderTestApp(t, orders, procurement.PurchaseOrderLineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/purchase-orders/1/pay", `{"journal_id":90}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseOrderHandler_WritePurchaseOrderError_MapsErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "order not found", err: procurement.ErrPurchaseOrderNotFound, want: http.StatusNotFound},
		{name: "order state", err: procurement.ErrPurchaseOrderState, want: http.StatusConflict},
		{name: "order no lines", err: procurement.ErrPurchaseOrderNoLines, want: http.StatusUnprocessableEntity},
		{name: "order line qty", err: procurement.ErrPurchaseOrderLineQty, want: http.StatusUnprocessableEntity},
		{name: "order line discount", err: procurement.ErrPurchaseOrderLineDiscount, want: http.StatusUnprocessableEntity},
		{name: "order supplier", err: procurement.ErrPurchaseOrderVendor, want: http.StatusUnprocessableEntity},
		{name: "order supplier not supplier", err: procurement.ErrPurchaseOrderVendorNotSupplier, want: http.StatusUnprocessableEntity},
		{name: "order warehouse", err: procurement.ErrPurchaseOrderWarehouse, want: http.StatusUnprocessableEntity},
		{name: "order no offer", err: procurement.ErrPurchaseOrderNoOffer, want: http.StatusUnprocessableEntity},
		{name: "order approval pending", err: procurement.ErrPurchaseOrderApprovalPending, want: http.StatusConflict},
		{name: "order approval required", err: procurement.ErrPurchaseOrderApprovalRequired, want: http.StatusConflict},
		{name: "order nothing to receive", err: procurement.ErrPurchaseOrderNothingToReceive, want: http.StatusConflict},
		{name: "order nothing to bill", err: procurement.ErrPurchaseOrderNothingToBill, want: http.StatusConflict},
		{name: "order bill quality blocked", err: procurement.ErrPurchaseOrderBillQualityBlocked, want: http.StatusConflict},
		{name: "order no open bills", err: procurement.ErrPurchaseOrderNoOpenBills, want: http.StatusConflict},
		{name: "requisition not found", err: procurement.ErrRequisitionNotFound, want: http.StatusNotFound},
		{name: "requisition not convertible", err: procurement.ErrRequisitionNotConvertible, want: http.StatusConflict},
		{name: "quoteRequest not found", err: procurement.ErrRFQNotFound, want: http.StatusNotFound},
		{name: "quoteRequest state", err: procurement.ErrQuoteRequestState, want: http.StatusConflict},
		{name: "quoteRequest no lines", err: procurement.ErrRFQNoLines, want: http.StatusUnprocessableEntity},
		{name: "quoteRequest line qty", err: procurement.ErrRFQLineQty, want: http.StatusUnprocessableEntity},
		{name: "quoteRequest requester", err: procurement.ErrRFQRequester, want: http.StatusUnprocessableEntity},
		{name: "quoteRequest supplier", err: procurement.ErrRFQVendor, want: http.StatusUnprocessableEntity},
		{name: "quoteRequest not convertible", err: procurement.ErrRFQNotConvertible, want: http.StatusConflict},
		{name: "quoteRequest quote not found", err: procurement.ErrRFQQuoteNotFound, want: http.StatusNotFound},
		{name: "quoteRequest quote state", err: procurement.ErrSupplierQuoteState, want: http.StatusConflict},
		{name: "quoteRequest quote no lines", err: procurement.ErrRFQQuoteNoLines, want: http.StatusUnprocessableEntity},
		{name: "quoteRequest quote line qty", err: procurement.ErrRFQQuoteLineQty, want: http.StatusUnprocessableEntity},
		{name: "quoteRequest quote no accepted", err: procurement.ErrRFQQuoteNoAccepted, want: http.StatusConflict},
		{name: "quoteRequest quote supplier", err: procurement.ErrRFQQuoteVendor, want: http.StatusUnprocessableEntity},
		{name: "quoteRequest quote not submitted", err: procurement.ErrRFQQuoteNotSubmitted, want: http.StatusConflict},
		{name: "quoteRequest quote price", err: procurement.ErrRFQQuotePrice, want: http.StatusUnprocessableEntity},
		{name: "quoteRequest quote duplicate line", err: procurement.ErrRFQQuoteDuplicateLine, want: http.StatusUnprocessableEntity},
		{name: "no payable account", err: accounting.ErrNoPayableAccount, want: http.StatusUnprocessableEntity},
		{name: "invoice no lines", err: accounting.ErrInvoiceNoLines, want: http.StatusUnprocessableEntity},
		{name: "invoice tax invalid", err: accounting.ErrInvoiceTaxInvalid, want: http.StatusUnprocessableEntity},
		{name: "unexpected error", err: errors.New("boom"), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/err", func(c fiber.Ctx) error {
				return writePurchaseOrderError(c, tt.err)
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
