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
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/pos"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestPOSConfigHandler_List_ReturnsConfigs(t *testing.T) {
	configs := dao.CRUDMock[reference.POSConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSConfig], error) {
			return &query.Page[reference.POSConfig]{Items: []*reference.POSConfig{{Base: model.Base{ID: 1}, Name: "Counter A"}}}, nil
		},
	}
	h := NewPOSConfigHandler(pos.NewPOSConfigService(configs))
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/pos/configs", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPOSConfigHandler_Create_CreatesConfig(t *testing.T) {
	var created *reference.POSConfig
	configs := dao.CRUDMock[reference.POSConfig]{
		CreateFunc: func(_ context.Context, c *reference.POSConfig) (*reference.POSConfig, error) {
			c.ID = 1
			created = c
			return c, nil
		},
	}
	h := NewPOSConfigHandler(pos.NewPOSConfigService(configs))
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/pos/configs", `{"name":"Counter A","warehouse_id":4,"journal_id":2,"price_book_id":3}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if created == nil || created.Name != "Counter A" {
		t.Errorf("created = %+v, want name Counter A", created)
	}
}

func TestPOSConfigHandler_Get_ReturnsConfig(t *testing.T) {
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
			return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Name: "Counter A"}, nil
		},
	}
	h := NewPOSConfigHandler(pos.NewPOSConfigService(configs))
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/pos/configs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPOSSessionHandler_Open_OpensSession(t *testing.T) {
	sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
		CreateFunc: func(_ context.Context, s *pos.POSSession) (*pos.POSSession, error) {
			s.ID = 1
			return s, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
			return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}, nil
		},
	}
	svc := testPOSHandlerService(configs, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
	h := NewPOSSessionHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/pos/sessions", `{"config_id":1,"cashier_id":7,"opening_balance":100}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPOSSessionHandler_Open_RejectsDuplicate(t *testing.T) {
	sessions := pos.POSSessionDAOMock{
		CRUDMock: dao.CRUDMock[pos.POSSession]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[pos.POSSession], error) {
				return &query.Page[pos.POSSession]{Items: []*pos.POSSession{{Base: model.Base{ID: 2}, State: pos.SessionStateOpened}}}, nil
			},
		},
	}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
			return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}, nil
		},
	}
	svc := testPOSHandlerService(configs, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
	h := NewPOSSessionHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/pos/sessions", `{"config_id":1,"cashier_id":7,"opening_balance":100}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPOSSessionHandler_Close_ReconcilesSession(t *testing.T) {
	sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
			return &pos.POSSession{Base: model.Base{ID: 1}, OpeningBalance: 100, State: pos.SessionStateClosing}, nil
		},
	}}
	orders := pos.POSOrderDAOMock{
		CRUDMock: dao.CRUDMock[pos.POSOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[pos.POSOrder], error) {
				return &query.Page[pos.POSOrder]{Items: []*pos.POSOrder{{Base: model.Base{ID: 5}}}}, nil
			},
		},
	}
	payments := pos.POSPaymentDAOMock{
		SumBySessionFunc: func(_ context.Context, _ uint64) ([]pos.PaymentMethodTotal, error) {
			return []pos.PaymentMethodTotal{{Method: "cash", Total: 150}}, nil
		},
	}
	svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, orders, pos.POSOrderLineDAOMock{}, payments)
	h := NewPOSSessionHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/pos/sessions/1/close", `{"closing_balance":250}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPOSSessionHandler_Close_RejectsMismatch(t *testing.T) {
	sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
			return &pos.POSSession{Base: model.Base{ID: 1}, OpeningBalance: 100, State: pos.SessionStateClosing}, nil
		},
	}}
	orders := pos.POSOrderDAOMock{
		CRUDMock: dao.CRUDMock[pos.POSOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[pos.POSOrder], error) {
				return &query.Page[pos.POSOrder]{Items: []*pos.POSOrder{{Base: model.Base{ID: 5}}}}, nil
			},
		},
	}
	payments := pos.POSPaymentDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*pos.POSPayment, error) {
			return []*pos.POSPayment{{Base: model.Base{ID: 1}, Amount: 150}}, nil
		},
	}
	svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, orders, pos.POSOrderLineDAOMock{}, payments)
	h := NewPOSSessionHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/pos/sessions/1/close", `{"closing_balance":240}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPOSOrderHandler_Sell_CreatesOrder(t *testing.T) {
	organizationID := uint64(1)
	config := reference.POSConfig{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		WarehouseID: helper.Ptr(uint64(4)), JournalID: helper.Ptr(uint64(2)), PriceBookID: helper.Ptr(uint64(3)),
	}
	sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
			return &pos.POSSession{Base: model.Base{ID: 1}, ConfigID: 1, State: pos.SessionStateOpened}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	orders := pos.POSOrderDAOMock{}
	svc := testPOSHandlerService(configs, sessions, orders, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
	h := NewPOSOrderHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/pos/orders", `{"session_id":1,"lines":[{"item_id":100,"qty":2,"tax_ids":[9]}],"payments":[{"method":"cash","amount":220}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPOSOrderHandler_Sell_RejectsSessionNotFound(t *testing.T) {
	sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) { return nil, nil },
	}}
	svc := testPOSHandlerService(dao.CRUDMock[reference.POSConfig]{}, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
	h := NewPOSOrderHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/pos/orders", `{"session_id":99,"lines":[{"item_id":100,"qty":2}],"payments":[{"method":"cash","amount":220}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPOSOrderHandler_Invoice_GeneratesInvoice(t *testing.T) {
	organizationID := uint64(1)
	config := reference.POSConfig{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		WarehouseID: helper.Ptr(uint64(4)), JournalID: helper.Ptr(uint64(2)), PriceBookID: helper.Ptr(uint64(3)),
	}
	contactID := uint64(5)
	sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
			return &pos.POSSession{Base: model.Base{ID: 1}, ConfigID: 1, State: pos.SessionStateOpened}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	orders := pos.POSOrderDAOMock{CRUDMock: dao.CRUDMock[pos.POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*pos.POSOrder, error) {
			return &pos.POSOrder{Base: model.Base{ID: 10}, SessionID: 1, ContactID: &contactID, AmountTotal: amount.FromInt64(220)}, nil
		},
	}}
	lines := pos.POSOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*pos.POSOrderLine, error) {
			return []*pos.POSOrderLine{{
				Base: model.Base{ID: 11}, ItemID: helper.Ptr(uint64(100)), Qty: 2,
				UnitPrice: amount.FromFloat64(100), TaxIDs: helper.Int64Array{9}, PriceSubtotal: amount.FromFloat64(200), PriceTax: amount.FromFloat64(20), PriceTotal: amount.FromFloat64(220),
			}}, nil
		},
	}
	svc := testPOSHandlerService(configs, sessions, orders, lines, pos.POSPaymentDAOMock{})
	h := NewPOSOrderHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/pos/orders/10/invoice", `{"journal_id":2,"date":"2026-08-14"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPOSSessionHandler_List_ReturnsSessions(t *testing.T) {
	sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[pos.POSSession], error) {
			return &query.Page[pos.POSSession]{Items: []*pos.POSSession{{Base: model.Base{ID: 1}, State: pos.SessionStateOpened}}}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSConfig], error) {
			return &query.Page[reference.POSConfig]{Items: []*reference.POSConfig{{Base: model.Base{ID: 1}}}}, nil
		},
	}
	svc := testPOSHandlerService(configs, sessions, pos.POSOrderDAOMock{}, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
	h := NewPOSSessionHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/pos/sessions", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPOSSessionHandler_Get_ReturnsSessionWithBreakdown(t *testing.T) {
	sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
			return &pos.POSSession{Base: model.Base{ID: 1}, ConfigID: 1, State: pos.SessionStateOpened}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
			return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1))}, nil
		},
	}
	orders := pos.POSOrderDAOMock{
		CRUDMock: dao.CRUDMock[pos.POSOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[pos.POSOrder], error) {
				return &query.Page[pos.POSOrder]{Items: []*pos.POSOrder{{Base: model.Base{ID: 5}}}}, nil
			},
		},
	}
	payments := pos.POSPaymentDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*pos.POSPayment, error) {
			return []*pos.POSPayment{{Base: model.Base{ID: 1}, Method: "cash", Amount: 150}}, nil
		},
	}
	svc := testPOSHandlerService(configs, sessions, orders, pos.POSOrderLineDAOMock{}, payments)
	h := NewPOSSessionHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/pos/sessions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPOSOrderHandler_List_ReturnsOrders(t *testing.T) {
	sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[pos.POSSession], error) {
			return &query.Page[pos.POSSession]{Items: []*pos.POSSession{{Base: model.Base{ID: 1}}}}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSConfig], error) {
			return &query.Page[reference.POSConfig]{Items: []*reference.POSConfig{{Base: model.Base{ID: 1}}}}, nil
		},
	}
	orders := pos.POSOrderDAOMock{CRUDMock: dao.CRUDMock[pos.POSOrder]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[pos.POSOrder], error) {
			return &query.Page[pos.POSOrder]{Items: []*pos.POSOrder{{Base: model.Base{ID: 1}, State: pos.OrderStateDone}}}, nil
		},
	}}
	svc := testPOSHandlerService(configs, sessions, orders, pos.POSOrderLineDAOMock{}, pos.POSPaymentDAOMock{})
	h := NewPOSOrderHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/pos/orders", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPOSOrderHandler_Get_ReturnsOrderWithLinesAndPayments(t *testing.T) {
	organizationID := uint64(1)
	config := reference.POSConfig{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
	}
	sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
			return &pos.POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	orders := pos.POSOrderDAOMock{CRUDMock: dao.CRUDMock[pos.POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*pos.POSOrder, error) {
			return &pos.POSOrder{Base: model.Base{ID: 10}, SessionID: 1, State: pos.OrderStateDone}, nil
		},
	}}
	lines := pos.POSOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*pos.POSOrderLine, error) {
			return []*pos.POSOrderLine{{
				Base: model.Base{ID: 11}, ItemID: helper.Ptr(uint64(100)), Qty: 2,
				PriceSubtotal: amount.FromFloat64(200), PriceTax: amount.FromFloat64(20), PriceTotal: amount.FromFloat64(220),
			}}, nil
		},
	}
	payments := pos.POSPaymentDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*pos.POSPayment, error) {
			return []*pos.POSPayment{{Base: model.Base{ID: 1}, Method: "cash", Amount: 220}}, nil
		},
	}
	svc := testPOSHandlerService(configs, sessions, orders, lines, payments)
	h := NewPOSOrderHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/pos/orders/10", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPOSOrderHandler_Refund_RefundsOrder(t *testing.T) {
	organizationID := uint64(1)
	config := reference.POSConfig{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		WarehouseID: helper.Ptr(uint64(4)), JournalID: helper.Ptr(uint64(2)), PriceBookID: helper.Ptr(uint64(3)),
	}
	sessions := pos.POSSessionDAOMock{CRUDMock: dao.CRUDMock[pos.POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*pos.POSSession, error) {
			return &pos.POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	orders := pos.POSOrderDAOMock{CRUDMock: dao.CRUDMock[pos.POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*pos.POSOrder, error) {
			return &pos.POSOrder{Base: model.Base{ID: 10}, SessionID: 1, State: pos.OrderStateDone, AmountTotal: amount.FromFloat64(220), AmountTax: amount.FromFloat64(20)}, nil
		},
		UpdateFunc: func(_ context.Context, o *pos.POSOrder) (*pos.POSOrder, error) { return o, nil },
	}}
	lines := pos.POSOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*pos.POSOrderLine, error) {
			return []*pos.POSOrderLine{{
				ItemID: helper.Ptr(uint64(100)), Qty: 2, PriceSubtotal: amount.FromFloat64(200), PriceTax: amount.FromFloat64(20), TaxIDs: helper.Int64Array{9},
			}}, nil
		},
	}
	svc := testPOSHandlerService(configs, sessions, orders, lines, pos.POSPaymentDAOMock{})
	h := NewPOSOrderHandler(svc)
	app := fiber.New()
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/pos/orders/10/refund", `{"journal_id":2,"date":"2026-08-14"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func testPOSHandlerService(configs dao.CRUDMock[reference.POSConfig], sessions pos.POSSessionDAOMock, orders pos.POSOrderDAOMock, lines pos.POSOrderLineDAOMock, payments pos.POSPaymentDAOMock) pos.POSService {
	sequences := sequence.NewSequenceService(sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "POS/00001"}, nil
		},
	})
	return pos.NewPOSService(
		configs,
		sessions,
		orders,
		lines,
		payments,
		testHandlerProductService(),
		dao.CRUDMock[reference.Tax]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
				return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(10.0), TaxAccountID: helper.Ptr(uint64(60))}, nil
			},
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Tax], error) {
				return &query.Page[reference.Tax]{Items: []*reference.Tax{{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(10.0), TaxAccountID: helper.Ptr(uint64(60))}}}, nil
			},
		},
		dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 2}, DefaultAccountID: helper.Ptr(uint64(40))}, nil
			},
		},
		handlerAccountLookup{},
		pos.PaymentAccountLookupMock{},
		inventory.StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
					if q.Filters[0].Field == "usage" {
						return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
					}
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
				},
			},
		},
		inventory.StockBalanceDAOMock{
			FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
				return &inventory.StockBalance{Base: model.Base{ID: 3}, Quantity: 10}, nil
			},
		},
		inventory.StockMovementDAOMock{
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{
				CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
					movement.ID = 40
					return movement, nil
				},
			},
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{
					Base: model.Base{ID: 30}, ItemID: 100, Qty: 2,
					SrcLocationID: 10, State: inventory.MovementStateDone,
				}}, nil
			},
		},
		inventory.CostLayerDAOMock{
			ListByMovementFunc: func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
				return []*inventory.CostLayer{{MovementID: helper.Ptr(uint64(30)), UnitCost: helper.Ptr(100.0)}}, nil
			},
		},
		inventory.ShipmentDAOMock{},
		handlerShipEngine{},
		handlerInvoiceEngine{},
		inventory.PosterMock{},
		iam.MemberDAOMock{
			FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
				return &iam.Member{Base: model.Base{ID: 1}}, nil
			},
		},
		sequences,
		pos.TransactionerMock{},
	)
}

func testHandlerProductService() products.ProductService {
	return products.NewProductService(
		products.ItemDAOMock{
			CRUDMock: dao.CRUDMock[products.Item]{
				FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
					return &products.Item{Base: model.Base{ID: 1}, ListPrice: 100, CategoryID: helper.Ptr(uint64(8))}, nil
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
				return &reference.ItemCategory{Base: model.Base{ID: 8}, IncomeAccountID: helper.Ptr(uint64(70))}, nil
			},
		},
		products.PriceBookDAOMock{
			CRUDMock: dao.CRUDMock[products.PriceBook]{
				FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
					return &products.PriceBook{Base: model.Base{ID: 3}, Name: "Retail"}, nil
				},
			},
		},
		products.PriceRuleDAOMock{},
	)
}

type handlerAccountLookup struct{}

func (m handlerAccountLookup) List(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
	return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 50}, OrganizationID: 1, Type: "receivable", Active: true}}}, nil
}

type handlerShipEngine struct{}

func (m handlerShipEngine) Ship(_ context.Context, moveID uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
	return &inventory.CostLayer{MovementID: &moveID}, nil
}

func (m handlerShipEngine) ShipTx(_ context.Context, _ *gorm.DB, moveID uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
	return m.Ship(context.Background(), moveID, 0, time.Now())
}

func (m handlerShipEngine) Restock(_ context.Context, moveID uint64, _ amount.Amount, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
	return &inventory.CostLayer{MovementID: &moveID}, nil
}

func (m handlerShipEngine) RestockTx(_ context.Context, _ *gorm.DB, moveID uint64, unitCost amount.Amount, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
	return m.Restock(context.Background(), moveID, unitCost, 0, time.Now())
}

type handlerInvoiceEngine struct{}

func (m handlerInvoiceEngine) Create(_ context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	return &accounting.Invoice{Base: model.Base{ID: 77}, ContactID: request.ContactID}, nil
}

func (m handlerInvoiceEngine) CreateTx(_ context.Context, _ *gorm.DB, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	return m.Create(context.Background(), request)
}

func (m handlerInvoiceEngine) CreateCreditNote(_ context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
	return &accounting.Invoice{Base: model.Base{ID: 78}, ContactID: request.OriginalInvoiceID}, nil
}

func (m handlerInvoiceEngine) CreateCreditNoteTx(_ context.Context, _ *gorm.DB, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
	return m.CreateCreditNote(context.Background(), request)
}

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
