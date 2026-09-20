package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

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

func procurementTestApp(t *testing.T, register func(api fiber.Router, guards httpx.RouteGuards)) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		c.Locals(model.ActorKey, uint64(5))
		return c.Next()
	})
	register(app, passthroughGuards())
	return app
}

func procurementTestAppNoTenant(t *testing.T, register func(api fiber.Router, guards httpx.RouteGuards)) *fiber.App {
	t.Helper()
	app := fiber.New()
	register(app, passthroughGuards())
	return app
}

func samplePurchaseOrder() *procurement.PurchaseOrder {
	return &procurement.PurchaseOrder{
		Base:           model.Base{ID: 1, CreatedAt: timeNow(), UpdatedAt: timeNow()},
		OrganizationID: ptrUint64(10),
		Name:           ptrString("PO/00001"),
		SupplierID:     10,
		CurrencyCode:   ptrString("IDR"),
		WarehouseID:    ptrUint64(20),
		State:          procurement.PurchaseOrderStateDraft,
		AmountUntaxed:  2000,
		AmountTotal:    2000,
		InvoiceStatus:  procurement.PurchaseOrderInvoiceStatusNo,
		ReceiptStatus:  procurement.PurchaseOrderReceiptStatusPending,
	}
}

func samplePurchaseOrderLine() *procurement.PurchaseOrderLine {
	return &procurement.PurchaseOrderLine{
		Base:        model.Base{ID: 100},
		OrderID:     1,
		ItemID:      ptrUint64(200),
		Description: ptrString("Widget"),
		QtyOrdered:  2,
		UnitPrice:   1000,
	}
}

func samplePurchaseRequest() *procurement.PurchaseRequest {
	return &procurement.PurchaseRequest{
		Base:           model.Base{ID: 1, CreatedAt: timeNow(), UpdatedAt: timeNow()},
		OrganizationID: ptrUint64(10),
		Name:           ptrString("REQ/00001"),
		RequesterID:    5,
		State:          procurement.RequestStateDraft,
	}
}

func sampleSupplierQuoteRequest() *procurement.SupplierQuoteRequest {
	return &procurement.SupplierQuoteRequest{
		Base:           model.Base{ID: 1, CreatedAt: timeNow(), UpdatedAt: timeNow()},
		OrganizationID: ptrUint64(10),
		Name:           ptrString("QuoteRequest/00001"),
		RequesterID:    5,
		CurrencyCode:   ptrString("IDR"),
		State:          procurement.QuoteRequestStateDraft,
	}
}

func sampleSupplierQuoteRequestLine() *procurement.SupplierQuoteRequestLine {
	return &procurement.SupplierQuoteRequestLine{
		Base:           model.Base{ID: 100},
		QuoteRequestID: 1,
		ItemID:         ptrUint64(200),
		Description:    ptrString("Widget"),
		Qty:            2,
	}
}

func sampleSupplierQuoteLine() *procurement.SupplierQuoteLine {
	return &procurement.SupplierQuoteLine{
		Base:               model.Base{ID: 100},
		SupplierQuoteID:    1,
		QuoteRequestLineID: 100,
		ItemID:             ptrUint64(200),
		Description:        ptrString("Widget"),
		Qty:                2,
		UnitPrice:          1000,
		PriceSubtotal:      2000,
	}
}

func sampleSupplierQuote() *procurement.SupplierQuote {
	return &procurement.SupplierQuote{
		Base:           model.Base{ID: 1},
		QuoteRequestID: 1,
		SupplierID:     10,
		CurrencyCode:   ptrString("IDR"),
		State:          procurement.SupplierQuoteStateSubmitted,
		AmountUntaxed:  2000,
		AmountTotal:    2000,
	}
}

func ptrUint64(value uint64) *uint64 {
	return &value
}

func ptrString(value string) *string {
	return &value
}

func timeNow() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}
