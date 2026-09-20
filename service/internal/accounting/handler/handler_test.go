package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
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

func accountingTestApp(t *testing.T, withTenant bool, register func(api fiber.Router, guards httpx.RouteGuards)) *fiber.App {
	t.Helper()
	app := fiber.New()
	if withTenant {
		app.Use(func(c fiber.Ctx) error {
			c.Locals(httpx.LocalOrganizationID, uint64(10))
			c.Locals(model.ActorKey, uint64(5))
			return c.Next()
		})
	}
	register(app, passthroughGuards())
	return app
}

func sampleJournalEntry() *accounting.JournalEntry {
	return &accounting.JournalEntry{
		Base:           model.Base{ID: 1},
		OrganizationID: 10,
		JournalID:      1,
		Date:           time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		State:          accounting.EntryStatePosted,
	}
}

func sampleJournalLine() *accounting.JournalLine {
	return &accounting.JournalLine{
		Base:      model.Base{ID: 1},
		EntryID:   1,
		AccountID: 100,
		Debit:     amount.FromInt64(100),
		Date:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleInvoice() *accounting.Invoice {
	return &accounting.Invoice{
		Base:           model.Base{ID: 1},
		OrganizationID: ptrUint64(10),
		Type:           accounting.InvoiceTypeCustomerInvoice,
		ContactID:      5,
		State:          accounting.InvoiceStatePosted,
		PaymentState:   accounting.PaymentStateNotPaid,
	}
}

func samplePayment() *accounting.Payment {
	return &accounting.Payment{
		Base:           model.Base{ID: 1},
		OrganizationID: ptrUint64(10),
		ContactID:      5,
		Type:           accounting.PaymentTypeInbound,
		Amount:         100,
		Date:           time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		State:          accounting.PaymentStatePosted,
	}
}

func sampleTaxPeriod() *accounting.TaxPeriod {
	return &accounting.TaxPeriod{
		Base:           model.Base{ID: 1},
		OrganizationID: 10,
		TaxYearID:      1,
		Name:           "January",
		State:          accounting.TaxPeriodStateOpen,
	}
}

func sampleBankStatement() *accounting.BankStatement {
	return &accounting.BankStatement{
		Base:      model.Base{ID: 1},
		JournalID: ptrUint64(1),
		State:     accounting.BankStatementStateOpen,
	}
}

func sampleBudget() *accounting.Budget {
	return &accounting.Budget{
		Base:           model.Base{ID: 1},
		OrganizationID: ptrUint64(10),
		Name:           ptrString("FY2026"),
		State:          accounting.BudgetStateDraft,
	}
}

func sampleTaxRule() *accounting.TaxRule {
	return &accounting.TaxRule{
		Base:           model.Base{ID: 1},
		OrganizationID: ptrUint64(10),
		Name:           ptrString("Local"),
		Active:         true,
	}
}

func sampleWithholdingTax() *accounting.WithholdingTax {
	return &accounting.WithholdingTax{
		Base:           model.Base{ID: 1},
		OrganizationID: ptrUint64(10),
		Name:           ptrString("WHT"),
		RatePct:        2,
		Scope:          accounting.WithholdingScopeSale,
		Active:         true,
	}
}

func sampleTaxReturn() *accounting.TaxReturn {
	return &accounting.TaxReturn{
		Base:           model.Base{ID: 1},
		OrganizationID: ptrUint64(10),
		PeriodID:       1,
		State:          accounting.TaxReturnStateDraft,
	}
}

func ptrUint64(value uint64) *uint64 {
	return &value
}

func ptrString(value string) *string {
	return &value
}
