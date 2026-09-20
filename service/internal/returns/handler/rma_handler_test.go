package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/returns"
	"gorm.io/gorm"
)

func rmaHandlerTest(
	t *testing.T,
	rmas returns.RMADAOMock,
	lines returns.RMALineDAOMock,
	svc returns.RMAService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewRMAHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func rmaHandlerTestWithTenant(
	t *testing.T,
	rmas returns.RMADAOMock,
	lines returns.RMALineDAOMock,
	svc returns.RMAService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(1))
		return c.Next()
	})
	h := NewRMAHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func rmaTestSvc(
	rmas returns.RMADAOMock,
	lines returns.RMALineDAOMock,
	orders returns.OriginOrderLookupMock,
	movements inventory.StockMovementDAOMock,
	valuer returns.ReturnValuerMock,
	invoices accounting.InvoiceDAOMock,
	credits returns.CreditNoteEngineMock,
	sequences sequence.DAOMock,
) returns.RMAService {
	return returns.NewTestRMAService(rmas, lines, orders, movements, returns.StockLocationLookupMock{}, returns.StockLayerLookupMock{}, valuer, invoices, credits, sequences)
}

func rmaCRUD(crud dao.CRUDMock[returns.RMA]) returns.RMADAOMock {
	return returns.RMADAOMock{CRUDMock: crud}
}

func rmaLineCRUD(lines []*returns.RMALine, listErr error) returns.RMALineDAOMock {
	return returns.RMALineDAOMock{
		CRUDMock: dao.CRUDMock[returns.RMALine]{},
		ListByRMAFunc: func(_ context.Context, _ uint64) ([]*returns.RMALine, error) {
			return lines, listErr
		},
	}
}

func sampleRMA(state string) *returns.RMA {
	orgID := uint64(1)
	name := "RMA-0001"
	originType := "sale_order"
	originID := uint64(100)
	reason := "damaged"
	return &returns.RMA{
		Base:            model.Base{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		OrganizationID:  &orgID,
		Name:            &name,
		Type:            "customer_return",
		ContactID:       10,
		OriginOrderType: &originType,
		OriginOrderID:   &originID,
		Reason:          &reason,
		State:           state,
	}
}

func sampleRMALine() *returns.RMALine {
	return &returns.RMALine{
		Base:        model.Base{ID: 1},
		RMAID:       1,
		ItemID:      5,
		Qty:         2,
		Disposition: "restock",
	}
}

func emptyRMA() returns.RMADAOMock {
	return rmaCRUD(dao.CRUDMock[returns.RMA]{})
}

func ptrUint64(v uint64) *uint64 {
	return &v
}

func TestRMAHandler_List_ReturnsRMAs(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[returns.RMA], error) {
			return &query.Page[returns.RMA]{Items: []*returns.RMA{sampleRMA(returns.StateDraft)}, Count: 1}, nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/rmas/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRMAHandler_List_RejectsInvalidQuery(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/rmas/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_List_ReturnsServerError(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[returns.RMA], error) {
			return nil, errors.New("db down")
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/rmas/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRMAHandler_Get_ReturnsRMA(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return sampleRMA(returns.StateDraft), nil
		},
	})
	lines := rmaLineCRUD([]*returns.RMALine{sampleRMALine()}, nil)
	svc := rmaTestSvc(rmas, lines, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, lines, svc)

	resp, err := doRequest(app, http.MethodGet, "/rmas/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRMAHandler_Get_ReturnsNotFound(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return nil, nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/rmas/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRMAHandler_Get_ReturnsNotFoundForForeignTenant(t *testing.T) {
	foreign := sampleRMA(returns.StateDraft)
	otherOrg := uint64(2)
	foreign.OrganizationID = &otherOrg
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return foreign, nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTestWithTenant(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/rmas/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRMAHandler_Get_RejectsInvalidID(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/rmas/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Get_ReturnsServerError(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return nil, errors.New("db down")
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/rmas/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRMAHandler_Get_ReturnsServerErrorForLines(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return sampleRMA(returns.StateDraft), nil
		},
	})
	lines := rmaLineCRUD(nil, errors.New("db down"))
	svc := rmaTestSvc(rmas, lines, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, lines, svc)

	resp, err := doRequest(app, http.MethodGet, "/rmas/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRMAHandler_Create_CreatesRMA(t *testing.T) {
	orgID := uint64(7)
	created := sampleRMA(returns.StateDraft)
	created.OrganizationID = &orgID
	rmas := returns.RMADAOMock{
		CRUDMock: dao.CRUDMock[returns.RMA]{},
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, _ *returns.RMA, _ []*returns.RMALine) (*returns.RMA, error) {
			return created, nil
		},
	}
	orders := returns.OriginOrderLookupMock{
		CustomerOrderFunc: func(_ context.Context, _ uint64) (uint64, error) {
			return 10, nil
		},
	}
	sequences := sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "RMA-0001"}, nil
		},
	}
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, orders, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequences)
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"organization_id":7,"type":"customer_return","contact_id":10,"origin_order_type":"sale_order","origin_order_id":100,"reason":"damaged","lines":[{"item_id":5,"qty":2,"disposition":"restock"}]}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestRMAHandler_Create_RejectsValidation(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Create_RejectsUnresolvedOrganization(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"type":"customer_return","contact_id":10,"origin_order_type":"sale_order","origin_order_id":100,"lines":[{"item_id":5,"qty":2,"disposition":"restock"}]}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Create_RejectsInvalidType(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"organization_id":7,"type":"bogus","contact_id":10,"origin_order_type":"sale_order","origin_order_id":100,"lines":[{"item_id":5,"qty":2,"disposition":"restock"}]}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Create_ReturnsServerError(t *testing.T) {
	orders := returns.OriginOrderLookupMock{
		CustomerOrderFunc: func(_ context.Context, _ uint64) (uint64, error) {
			return 0, errors.New("db down")
		},
	}
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, orders, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"organization_id":7,"type":"customer_return","contact_id":10,"origin_order_type":"sale_order","origin_order_id":100,"lines":[{"item_id":5,"qty":2,"disposition":"restock"}]}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRMAHandler_Confirm_ConfirmsRMA(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return sampleRMA(returns.StateDraft), nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRMAHandler_Confirm_RejectsInvalidID(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/abc/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Confirm_ReturnsNotFound(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return nil, nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRMAHandler_Confirm_ReturnsStateConflict(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return sampleRMA(returns.StateConfirmed), nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Receive_ReceivesRMA(t *testing.T) {
	supplier := sampleRMA(returns.StateConfirmed)
	supplier.Type = "vendor_return"
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return supplier, nil
		},
	})
	lines := rmaLineCRUD([]*returns.RMALine{sampleRMALine()}, nil)
	movements := inventory.StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[inventory.StockMovement]{},
		ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{{
				Base:          model.Base{ID: 1},
				ItemID:        5,
				UnitID:        ptrUint64(1),
				SrcLocationID: 2,
				DstLocationID: 3,
				State:         inventory.MovementStateDone,
			}}, nil
		},
	}
	valuer := returns.ReturnValuerMock{
		ReturnToSupplierFunc: func(_ context.Context, moveID uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
			return &inventory.CostLayer{MovementID: &moveID}, nil
		},
	}
	svc := rmaTestSvc(rmas, lines, returns.OriginOrderLookupMock{}, movements, valuer, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, lines, svc)

	body := `{"journal_id":1,"date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRMAHandler_Receive_RejectsInvalidID(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"journal_id":1,"date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/abc/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Receive_RejectsValidation(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/1/receive", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Receive_RejectsBadDate(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"journal_id":1,"date":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Receive_ReturnsNotFound(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return nil, nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"journal_id":1,"date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRMAHandler_Receive_ReturnsStateConflict(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return sampleRMA(returns.StateDraft), nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"journal_id":1,"date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Refund_RefundsRMA(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return sampleRMA(returns.StateReceived), nil
		},
	})
	lines := rmaLineCRUD([]*returns.RMALine{sampleRMALine()}, nil)
	invoices := accounting.InvoiceDAOMock{
		FindBySaleOrderFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 1}}, nil
		},
	}
	credits := returns.CreditNoteEngineMock{
		CreateCreditNoteFunc: func(_ context.Context, _ accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := rmaTestSvc(rmas, lines, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, invoices, credits, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, lines, svc)

	body := `{"journal_id":1,"date":"2026-01-01","reference":"CN-001"}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/1/refund", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRMAHandler_Refund_RejectsInvalidID(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"journal_id":1,"date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/abc/refund", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Refund_RejectsValidation(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/1/refund", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Refund_RejectsBadDate(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"journal_id":1,"date":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/1/refund", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Refund_ReturnsNotFound(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return nil, nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"journal_id":1,"date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/1/refund", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRMAHandler_Refund_ReturnsInvoiceMissing(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return sampleRMA(returns.StateReceived), nil
		},
	})
	invoices := accounting.InvoiceDAOMock{
		FindBySaleOrderFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return nil, nil
		},
	}
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, invoices, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	body := `{"journal_id":1,"date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/rmas/1/refund", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Done_CompletesRMA(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return sampleRMA(returns.StateRefunded), nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRMAHandler_Done_RejectsInvalidID(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/abc/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Done_ReturnsNotFound(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return nil, nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRMAHandler_Done_ReturnsStateConflict(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return sampleRMA(returns.StateDraft), nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Cancel_CancelsRMA(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return sampleRMA(returns.StateDraft), nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRMAHandler_Cancel_RejectsInvalidID(t *testing.T) {
	rmas := emptyRMA()
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/abc/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestRMAHandler_Cancel_ReturnsStateConflict(t *testing.T) {
	rmas := rmaCRUD(dao.CRUDMock[returns.RMA]{
		FindFunc: func(_ context.Context, _ uint64) (*returns.RMA, error) {
			return sampleRMA(returns.StateDone), nil
		},
	})
	svc := rmaTestSvc(rmas, returns.RMALineDAOMock{}, returns.OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, returns.ReturnValuerMock{}, accounting.InvoiceDAOMock{}, returns.CreditNoteEngineMock{}, sequence.DAOMock{})
	app := rmaHandlerTest(t, rmas, returns.RMALineDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/rmas/1/cancel", "")
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
