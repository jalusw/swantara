package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func itemCategoryHandlerTest(t *testing.T, categories products.ItemCategoryService) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewItemCategoryHandler(categories)
	h.Register(app, passthroughGuards())
	return app
}

func itemCategoryHandlerTestNoTenant(t *testing.T, categories products.ItemCategoryService) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewItemCategoryHandler(categories)
	h.Register(app, passthroughGuards())
	return app
}

func productHandlerTest(
	t *testing.T,
	templates products.ItemDAO,
	variants products.ItemVariantDAO,
	svc products.ProductService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		c.Locals(model.ActorKey, uint64(5))
		return c.Next()
	})
	h := NewProductHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func productHandlerTestNoTenant(
	t *testing.T,
	templates products.ItemDAO,
	variants products.ItemVariantDAO,
	svc products.ProductService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewProductHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func price_bookHandlerTest(
	t *testing.T,
	price_books products.PriceBookDAO,
	rules products.PriceRuleDAO,
	svc products.ProductService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		c.Locals(model.ActorKey, uint64(5))
		return c.Next()
	})
	h := NewPriceBookHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func price_bookHandlerTestNoTenant(
	t *testing.T,
	price_books products.PriceBookDAO,
	rules products.PriceRuleDAO,
	svc products.ProductService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewPriceBookHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func supplierProductHandlerTest(
	t *testing.T,
	supplierProducts products.SupplierProductDAO,
	svc products.SupplierProductService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		c.Locals(model.ActorKey, uint64(5))
		return c.Next()
	})
	h := NewSupplierProductHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func supplierProductHandlerTestNoOrg(
	t *testing.T,
	supplierProducts products.SupplierProductDAO,
	svc products.SupplierProductService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewSupplierProductHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func productTestSvc(
	templates products.ItemDAO,
	variants products.ItemVariantDAO,
	categories dao.CRUD[reference.ItemCategory],
	price_books products.PriceBookDAO,
	rules products.PriceRuleDAO,
) products.ProductService {
	return products.NewProductService(templates, variants, categories, price_books, rules)
}

func supplierProductTestSvc(
	variants products.ItemVariantDAO,
	templates products.ItemDAO,
	supplierProducts products.SupplierProductDAO,
	contactDAO contacts.ContactDAO,
	supplierDAO contacts.SupplierProfileDAO,
) products.SupplierProductService {
	return products.NewSupplierProductService(variants, templates, supplierProducts, contactDAO, supplierDAO)
}

func sampleItemCategory() *reference.ItemCategory {
	return &reference.ItemCategory{
		Base:           model.Base{ID: 1, CreatedAt: timeNow(), UpdatedAt: timeNow()},
		OrganizationID: helper.Ptr(uint64(10)),
		Name:           "Acme",
		CostMethod:     helper.Ptr("average"),
		Valuation:      helper.Ptr("real_time"),
	}
}

func foreignItemCategory() *reference.ItemCategory {
	category := sampleItemCategory()
	category.OrganizationID = helper.Ptr(uint64(99))
	return category
}

func sampleProduct() *products.Item {
	return &products.Item{
		Base:           model.Base{ID: 1, CreatedAt: timeNow(), UpdatedAt: timeNow()},
		OrganizationID: helper.Ptr(uint64(10)),
		Name:           "Widget",
		Type:           "stockable",
		Tracking:       "none",
		IsPurchasable:  true,
		IsSellable:     true,
		IsManufactured: false,
		Active:         true,
	}
}

func foreignProduct() *products.Item {
	template := sampleProduct()
	template.OrganizationID = helper.Ptr(uint64(99))
	return template
}

func sampleVariant() *products.ItemVariant {
	return &products.ItemVariant{
		Base:   model.Base{ID: 1, CreatedAt: timeNow(), UpdatedAt: timeNow()},
		ItemID: 1,
		Active: true,
	}
}

func samplePriceBook() *products.PriceBook {
	return &products.PriceBook{
		Base:           model.Base{ID: 1, CreatedAt: timeNow(), UpdatedAt: timeNow()},
		Name:           "Standard",
		CurrencyCode:   helper.Ptr("IDR"),
		OrganizationID: helper.Ptr(uint64(10)),
		Active:         true,
	}
}

func foreignPriceBook() *products.PriceBook {
	price_book := samplePriceBook()
	price_book.OrganizationID = helper.Ptr(uint64(99))
	return price_book
}

func sampleRule() *products.PriceRule {
	return &products.PriceRule{
		Base:        model.Base{ID: 1, CreatedAt: timeNow(), UpdatedAt: timeNow()},
		PriceBookID: 1,
		AppliesTo:   products.AppliesToAll,
		ComputeType: products.ComputeFixed,
		FixedPrice:  helper.Ptr(100.0),
	}
}

func sampleSupplierProduct() *products.SupplierProduct {
	return &products.SupplierProduct{
		Base:       model.Base{ID: 1, CreatedAt: timeNow(), UpdatedAt: timeNow()},
		ItemID:     1,
		SupplierID: 7,
		MinQty:     1,
		Price:      helper.Ptr(100.0),
		Priority:   10,
	}
}

func timeNow() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
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
