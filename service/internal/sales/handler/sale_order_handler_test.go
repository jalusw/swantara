package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

func saleOrderHandlerTest(
	t *testing.T,
	svc sales.SaleOrderService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(model.ActorKey, uint64(5))
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewSaleOrderHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func saleOrderHandlerTestNoTenant(
	t *testing.T,
	svc sales.SaleOrderService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewSaleOrderHandler(svc)
	h.Register(app, passthroughGuards())
	return app
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

func writeErrorStatus(app *fiber.App, path string, fn func(c fiber.Ctx) error) int {
	app.Post(path, func(c fiber.Ctx) error {
		return fn(c)
	})
	resp, err := doRequest(app, http.MethodPost, path, "")
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

func handlerTestPriceService(listPrice float64) products.ProductService {
	return products.NewProductService(
		products.ItemDAOMock{
			CRUDMock: dao.CRUDMock[products.Item]{
				FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
					return &products.Item{Base: model.Base{ID: 1}, ListPrice: listPrice}, nil
				},
			},
		},
		products.ItemVariantDAOMock{
			CRUDMock: dao.CRUDMock[products.ItemVariant]{
				FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
					return &products.ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
				},
			},
		},
		dao.CRUDMock[reference.ItemCategory]{},
		products.PriceBookDAOMock{
			CRUDMock: dao.CRUDMock[products.PriceBook]{
				FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
					return &products.PriceBook{Base: model.Base{ID: 2}, Name: "Retail", CurrencyCode: helper.Ptr("IDR")}, nil
				},
			},
		},
		products.PriceRuleDAOMock{},
	)
}

func handlerTestPriceServiceWithIncome(listPrice float64) products.ProductService {
	return products.NewProductService(
		products.ItemDAOMock{
			CRUDMock: dao.CRUDMock[products.Item]{
				FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
					return &products.Item{Base: model.Base{ID: 1}, ListPrice: listPrice, CategoryID: helper.Ptr(uint64(3))}, nil
				},
			},
		},
		products.ItemVariantDAOMock{
			CRUDMock: dao.CRUDMock[products.ItemVariant]{
				FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
					return &products.ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
				},
			},
		},
		dao.CRUDMock[reference.ItemCategory]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
				return &reference.ItemCategory{Base: model.Base{ID: 3}, IncomeAccountID: helper.Ptr(uint64(400))}, nil
			},
		},
		products.PriceBookDAOMock{
			CRUDMock: dao.CRUDMock[products.PriceBook]{
				FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
					return &products.PriceBook{Base: model.Base{ID: 2}, Name: "Retail", CurrencyCode: helper.Ptr("IDR")}, nil
				},
			},
		},
		products.PriceRuleDAOMock{},
	)
}

func defaultContact() contacts.ContactDAOMock {
	return contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 5}}, nil
			},
		},
	}
}

func sampleDraftOrder() *sales.SaleOrder {
	return &sales.SaleOrder{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(10)),
		Name:           helper.Ptr("SO/00001"),
		ContactID:      5,
		PriceBookID:    helper.Ptr(uint64(2)),
		ProspectID:     helper.Ptr(uint64(3)),
		WarehouseID:    helper.Ptr(uint64(4)),
		State:          sales.OrderStateDraft,
	}
}

func sampleConfirmedOrder() *sales.SaleOrder {
	order := sampleDraftOrder()
	order.State = sales.OrderStateConfirmed
	return order
}

func TestSaleOrderHandler_List_ReturnsOrders(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[sales.SaleOrder], error) {
				return &query.Page[sales.SaleOrder]{Items: []*sales.SaleOrder{sampleDraftOrder()}, Count: 1}, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/sale-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSaleOrderHandler_List_ExportsCSV(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[sales.SaleOrder], error) {
				return &query.Page[sales.SaleOrder]{Items: []*sales.SaleOrder{sampleDraftOrder()}, Count: 1}, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/sale-orders/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSaleOrderHandler_List_RejectsInvalidQuery(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/sale-orders/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_List_RejectsMissingTenant(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTestNoTenant(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/sale-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestSaleOrderHandler_List_ReturnsServerError(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[sales.SaleOrder], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/sale-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Get_ReturnsOrderWithLines(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
		},
	}
	lines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}, OrderID: 1, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}}, nil
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, lines: lines, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/sale-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Get_RejectsInvalidID(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/sale-orders/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Get_ReturnsNotFound(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/sale-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Get_ReturnsNotFoundForForeignTenant(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return &sales.SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(99))}, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/sale-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Get_ReturnsServerError(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/sale-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Get_ReturnsServerErrorOnLines(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
		},
	}
	lines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return nil, errors.New("db down")
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, lines: lines, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/sale-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Create_ReturnsCreated(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CreateWithLinesFunc: func(_ context.Context, order *sales.SaleOrder, _ []*sales.SaleOrderLine) (*sales.SaleOrder, error) {
			order.ID = 1
			return order, nil
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{
		orders:     orders,
		productSvc: handlerTestPriceService(100),
		contacts:   defaultContact(),
	})
	app := saleOrderHandlerTest(t, svc)

	body := `{"contact_id":5,"price_book_id":2,"prospect_id":3,"warehouse_id":4,"lines":[{"item_id":100,"qty_ordered":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Create_WalkInWithoutProspect(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CreateWithLinesFunc: func(_ context.Context, order *sales.SaleOrder, _ []*sales.SaleOrderLine) (*sales.SaleOrder, error) {
			order.ID = 1
			return order, nil
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{
		orders:     orders,
		productSvc: handlerTestPriceService(100),
		contacts:   defaultContact(),
	})
	app := saleOrderHandlerTest(t, svc)

	body := `{"contact_id":5,"price_book_id":2,"warehouse_id":4,"lines":[{"item_id":100,"qty_ordered":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Create_RejectsInvalidBody(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"contact_id":5,"price_book_id":2,"prospect_id":3,"lines":[{"item_id":100,"qty_ordered":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Create_RejectsInvalidOrderDate(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"contact_id":5,"price_book_id":2,"prospect_id":3,"warehouse_id":4,"order_date":"bogus","lines":[{"item_id":100,"qty_ordered":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Create_RejectsMissingTenant(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTestNoTenant(t, svc)

	body := `{"contact_id":5,"price_book_id":2,"prospect_id":3,"warehouse_id":4,"lines":[{"item_id":100,"qty_ordered":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Create_RejectsContactFromAnotherOrganization(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	contactsMock := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 5}, OrganizationID: helper.Ptr(uint64(20))}, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{
		orders:     orders,
		productSvc: handlerTestPriceService(100),
		contacts:   contactsMock,
	})
	app := saleOrderHandlerTest(t, svc)

	body := `{"contact_id":5,"price_book_id":2,"prospect_id":3,"warehouse_id":4,"lines":[{"item_id":100,"qty_ordered":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Update_ReturnsUpdatedOrder(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *sales.SaleOrder) (*sales.SaleOrder, error) {
				return order, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{
		orders:     orders,
		productSvc: handlerTestPriceService(100),
		contacts:   defaultContact(),
	})
	app := saleOrderHandlerTest(t, svc)

	body := `{"contact_id":5,"price_book_id":2,"warehouse_id":4,"lines":[{"item_id":100,"qty_ordered":3}]}`
	resp, err := doRequest(app, http.MethodPut, "/sale-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Update_RejectsInvalidID(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"contact_id":5,"price_book_id":2,"warehouse_id":4,"lines":[{"item_id":100,"qty_ordered":3}]}`
	resp, err := doRequest(app, http.MethodPut, "/sale-orders/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Update_RejectsInvalidBody(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"contact_id":0,"price_book_id":2,"warehouse_id":4,"lines":[{"item_id":100,"qty_ordered":3}]}`
	resp, err := doRequest(app, http.MethodPut, "/sale-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Update_ReturnsNotFound(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"contact_id":5,"price_book_id":2,"warehouse_id":4,"lines":[{"item_id":100,"qty_ordered":3}]}`
	resp, err := doRequest(app, http.MethodPut, "/sale-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Update_ReturnsConflictWhenNotDraft(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleConfirmedOrder(), nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"contact_id":5,"price_book_id":2,"warehouse_id":4,"lines":[{"item_id":100,"qty_ordered":3}]}`
	resp, err := doRequest(app, http.MethodPut, "/sale-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Update_ReturnsServerError(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"contact_id":5,"price_book_id":2,"warehouse_id":4,"lines":[{"item_id":100,"qty_ordered":3}]}`
	resp, err := doRequest(app, http.MethodPut, "/sale-orders/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Delete_ReturnsNoContent(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodDelete, "/sale-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Delete_RejectsInvalidID(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodDelete, "/sale-orders/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Delete_ReturnsNotFound(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodDelete, "/sale-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Delete_ReturnsServerError(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return errors.New("db down")
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodDelete, "/sale-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Send_ReturnsSentOrder(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *sales.SaleOrder) (*sales.SaleOrder, error) {
				return order, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/send", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Send_RejectsInvalidID(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/abc/send", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Send_ReturnsNotFound(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/send", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Send_ReturnsConflictWhenNotDraft(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleConfirmedOrder(), nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/send", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Send_ReturnsServerError(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/send", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Confirm_ReturnsConfirmedOrder(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *sales.SaleOrder) (*sales.SaleOrder, error) {
				return order, nil
			},
		},
	}
	lines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}, OrderID: 1}}, nil
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, lines: lines, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Confirm_RejectsInvalidID(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/abc/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Confirm_ReturnsNotFound(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Confirm_ReturnsConflictWhenNotConfirmable(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				order := sampleConfirmedOrder()
				return order, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Confirm_RejectsMissingWarehouse(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				order := sampleDraftOrder()
				order.WarehouseID = nil
				return order, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Confirm_ReturnsServerError(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Cancel_ReturnsCancelledOrder(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *sales.SaleOrder) (*sales.SaleOrder, error) {
				return order, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Cancel_RejectsInvalidID(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/abc/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Cancel_ReturnsNotFound(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Cancel_ReturnsConflictWhenNotCancellable(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				order := sampleDraftOrder()
				order.State = sales.OrderStateDone
				return order, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Done_ReturnsDoneOrder(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleConfirmedOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *sales.SaleOrder) (*sales.SaleOrder, error) {
				return order, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Done_RejectsInvalidID(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/abc/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Done_ReturnsNotFound(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Done_ReturnsConflictWhenNotConfirmable(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSaleOrderHandler_RecomputeStatuses_ReturnsUpdatedOrder(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *sales.SaleOrder) (*sales.SaleOrder, error) {
				return order, nil
			},
		},
	}
	lines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}, OrderID: 1, QtyOrdered: 2, QtyDelivered: 2, QtyInvoiced: 2}}, nil
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, lines: lines, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/recompute-statuses", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSaleOrderHandler_RecomputeStatuses_RejectsInvalidID(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/abc/recompute-statuses", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_RecomputeStatuses_ReturnsNotFound(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/recompute-statuses", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Deliver_ReturnsDeliveredOrder(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleConfirmedOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *sales.SaleOrder) (*sales.SaleOrder, error) {
				return order, nil
			},
		},
	}
	lines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}, OrderID: 1, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, QtyDelivered: 0}}, nil
		},
		CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{
			UpdateFunc: func(_ context.Context, line *sales.SaleOrderLine) (*sales.SaleOrderLine, error) {
				return line, nil
			},
		},
	}
	shipments := inventory.ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[inventory.Shipment]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
				return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 1}, State: inventory.ShipmentStateAssigned}}}, nil
			},
			UpdateFunc: func(_ context.Context, shipment *inventory.Shipment) (*inventory.Shipment, error) {
				return shipment, nil
			},
		},
	}
	movements := inventory.StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{{Base: model.Base{ID: 5}, ItemID: 100, Qty: 2, State: inventory.MovementStateConfirmed}}, nil
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{
		orders:     orders,
		lines:      lines,
		productSvc: handlerTestPriceService(100),
		shipments:  shipments,
		movements:  movements,
	})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3,"date":"2026-02-01"}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/deliver", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Deliver_RejectsInvalidID(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/abc/deliver", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Deliver_RejectsInvalidBody(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":0}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/deliver", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Deliver_RejectsInvalidDate(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3,"date":"bogus"}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/deliver", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Deliver_ReturnsNotFound(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/deliver", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Deliver_ReturnsConflictWhenNotConfirmed(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleDraftOrder(), nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/deliver", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Deliver_ReturnsConflictWhenNoShipment(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleConfirmedOrder(), nil
			},
		},
	}
	lines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}, OrderID: 1, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}}, nil
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, lines: lines, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/deliver", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Invoice_ReturnsInvoice(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleConfirmedOrder(), nil
			},
			UpdateFunc: func(_ context.Context, order *sales.SaleOrder) (*sales.SaleOrder, error) {
				return order, nil
			},
		},
	}
	lines := sales.SaleOrderLineDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{
			UpdateFunc: func(_ context.Context, line *sales.SaleOrderLine) (*sales.SaleOrderLine, error) {
				return line, nil
			},
		},
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}, OrderID: 1, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, QtyDelivered: 2, QtyInvoiced: 0, UnitPrice: 100}}, nil
		},
	}
	invoices := sales.InvoiceEngineMock{
		CreateFunc: func(_ context.Context, _ accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 1}, ContactID: 5, State: "posted"}, nil
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{
		orders:     orders,
		lines:      lines,
		productSvc: handlerTestPriceServiceWithIncome(100),
		invoices:   invoices,
	})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3,"date":"2026-02-01"}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/invoice", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Invoice_RejectsInvalidID(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/abc/invoice", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Invoice_RejectsInvalidBody(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":0}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/invoice", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Invoice_RejectsInvalidDate(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3,"date":"bogus"}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/invoice", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Invoice_ReturnsNotFound(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/invoice", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Invoice_ReturnsServerError(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/invoice", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Pay_ReturnsPayment(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return sampleConfirmedOrder(), nil
			},
		},
	}
	openInvoices := sales.InvoiceLookupMock{
		ListOpenByContactFunc: func(_ context.Context, _ uint64) ([]*accounting.Invoice, error) {
			return []*accounting.Invoice{{Base: model.Base{ID: 1}, ContactID: 5}}, nil
		},
	}
	payments := sales.PaymentEngineMock{
		CreateFunc: func(_ context.Context, _ accounting.CreatePaymentRequest) (*accounting.Payment, error) {
			return &accounting.Payment{Base: model.Base{ID: 1}, ContactID: 5, Amount: 100}, nil
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{
		orders:       orders,
		productSvc:   handlerTestPriceService(100),
		openInvoices: openInvoices,
		payments:     payments,
	})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3,"amount":100,"date":"2026-02-01"}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Pay_RejectsInvalidID(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3,"amount":100}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/abc/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Pay_RejectsInvalidBody(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3,"amount":0}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Pay_RejectsInvalidDate(t *testing.T) {
	orders := sales.SaleOrderDAOMock{}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3,"amount":100,"date":"bogus"}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSaleOrderHandler_Pay_ReturnsNotFound(t *testing.T) {
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := newSaleOrderHandlerTestSvc(saleOrderHandlerDeps{orders: orders, productSvc: handlerTestPriceService(100)})
	app := saleOrderHandlerTest(t, svc)

	body := `{"journal_id":3,"amount":100}`
	resp, err := doRequest(app, http.MethodPost, "/sale-orders/1/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSaleOrderHandler_WriteSaleOrderError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "not found", err: sales.ErrOrderNotFound, want: http.StatusNotFound},
		{name: "state", err: sales.ErrOrderState, want: http.StatusConflict},
		{name: "no lines", err: sales.ErrOrderNoLines, want: http.StatusUnprocessableEntity},
		{name: "quantity", err: sales.ErrOrderQty, want: http.StatusUnprocessableEntity},
		{name: "discount", err: sales.ErrOrderDiscount, want: http.StatusUnprocessableEntity},
		{name: "contact", err: sales.ErrOrderContact, want: http.StatusUnprocessableEntity},
		{name: "price_book", err: sales.ErrOrderPriceBook, want: http.StatusUnprocessableEntity},
		{name: "warehouse", err: sales.ErrOrderWarehouse, want: http.StatusUnprocessableEntity},
		{name: "contact organization", err: sales.ErrOrderContactOrganization, want: http.StatusUnprocessableEntity},
		{name: "price_book organization", err: sales.ErrOrderPriceBookOrganization, want: http.StatusUnprocessableEntity},
		{name: "warehouse organization", err: sales.ErrOrderWarehouseOrganization, want: http.StatusUnprocessableEntity},
		{name: "variant", err: sales.ErrOrderVariant, want: http.StatusUnprocessableEntity},
		{name: "variant organization", err: sales.ErrOrderVariantOrganization, want: http.StatusUnprocessableEntity},
		{name: "lead", err: sales.ErrOrderLead, want: http.StatusUnprocessableEntity},
		{name: "lead organization", err: sales.ErrOrderLeadOrganization, want: http.StatusUnprocessableEntity},
		{name: "lead not won", err: sales.ErrOrderLeadNotWon, want: http.StatusConflict},
		{name: "tax", err: sales.ErrOrderTax, want: http.StatusUnprocessableEntity},
		{name: "tax invalid", err: sales.ErrOrderTaxInvalid, want: http.StatusUnprocessableEntity},
		{name: "tax organization", err: sales.ErrOrderTaxOrganization, want: http.StatusUnprocessableEntity},
		{name: "currency", err: sales.ErrOrderCurrency, want: http.StatusUnprocessableEntity},
		{name: "location", err: sales.ErrOrderLocation, want: http.StatusUnprocessableEntity},
		{name: "customer location", err: sales.ErrOrderCustomerLocation, want: http.StatusUnprocessableEntity},
		{name: "stock unavailable", err: sales.ErrOrderStockUnavailable, want: http.StatusConflict},
		{name: "quant not found", err: inventory.ErrBalanceNotFound, want: http.StatusConflict},
		{name: "reservation overflow", err: inventory.ErrHoldOverflow, want: http.StatusConflict},
		{name: "shipment not found", err: sales.ErrOrderShipmentNotFound, want: http.StatusConflict},
		{name: "nothing to deliver", err: sales.ErrOrderNothingToDeliver, want: http.StatusConflict},
		{name: "no invoices", err: sales.ErrOrderNoInvoices, want: http.StatusConflict},
		{name: "no receivable account", err: accounting.ErrNoReceivableAccount, want: http.StatusUnprocessableEntity},
		{name: "no revenue account", err: accounting.ErrNoRevenueAccount, want: http.StatusUnprocessableEntity},
		{name: "invoice no lines", err: accounting.ErrInvoiceNoLines, want: http.StatusUnprocessableEntity},
		{name: "invoice tax invalid", err: accounting.ErrInvoiceTaxInvalid, want: http.StatusUnprocessableEntity},
		{name: "invoice sequence", err: accounting.ErrInvoiceSequence, want: http.StatusUnprocessableEntity},
		{name: "no income account", err: products.ErrNoIncomeAccount, want: http.StatusUnprocessableEntity},
		{name: "payment amount", err: accounting.ErrPaymentAmount, want: http.StatusUnprocessableEntity},
		{name: "payment no invoices", err: accounting.ErrPaymentNoInvoices, want: http.StatusUnprocessableEntity},
		{name: "over allocation", err: accounting.ErrOverAllocation, want: http.StatusConflict},
		{name: "no bank account", err: accounting.ErrNoBankAccount, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			got := writeErrorStatus(app, "/write-error", func(c fiber.Ctx) error {
				return writeSaleOrderError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
