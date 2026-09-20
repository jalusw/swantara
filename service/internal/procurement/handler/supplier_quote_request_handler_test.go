package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

func rfqTestApp(t *testing.T, quote_requests procurement.SupplierQuoteRequestDAOMock, lines procurement.SupplierQuoteRequestLineDAOMock, quotes procurement.SupplierQuoteDAOMock, quoteLines procurement.SupplierQuoteLineDAOMock, svc procurement.SupplierQuoteRequestService) *fiber.App {
	t.Helper()
	return procurementTestApp(t, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewSupplierQuoteRequestHandler(svc)
		h.Register(api, guards)
	})
}

func rfqTestSvc(quote_requests procurement.SupplierQuoteRequestDAOMock, lines procurement.SupplierQuoteRequestLineDAOMock, quotes procurement.SupplierQuoteDAOMock, quoteLines procurement.SupplierQuoteLineDAOMock) procurement.SupplierQuoteRequestService {
	return procurement.NewTestSupplierQuoteRequestService(procurement.SupplierQuoteRequestServiceTestDeps{
		QuoteRequests: quote_requests,
		RFQLines:      lines,
		Quotes:        quotes,
		QuoteLines:    quoteLines,
	})
}

func TestSupplierQuoteRequestHandler_List_ReturnsRFQs(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierQuoteRequest], error) {
				return &query.Page[procurement.SupplierQuoteRequest]{Items: []*procurement.SupplierQuoteRequest{sampleSupplierQuoteRequest()}, Count: 1}, nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_List_RejectsInvalidQuery(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_List_ReturnsUnauthorizedWhenTenantMissing(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierQuoteRequest], error) {
				return &query.Page[procurement.SupplierQuoteRequest]{Items: []*procurement.SupplierQuoteRequest{}, Count: 0}, nil
			},
		},
	}
	svc := rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{})
	app := procurementTestAppNoTenant(t, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewSupplierQuoteRequestHandler(svc)
		h.Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_List_ReturnsServerError(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierQuoteRequest], error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Get_ReturnsRFQ(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return sampleSupplierQuoteRequest(), nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Get_ReturnsNotFound(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return nil, nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Get_RejectsForeignTenant(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				quoteRequest := sampleSupplierQuoteRequest()
				quoteRequest.OrganizationID = ptrUint64(99)
				return quoteRequest, nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Get_RejectsInvalidID(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Get_ReturnsServerError(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Create_CreatesRFQ(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, quoteRequest *procurement.SupplierQuoteRequest, _ []*procurement.SupplierQuoteRequestLine) (*procurement.SupplierQuoteRequest, error) {
			quoteRequest.ID = 1
			return quoteRequest, nil
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Create_RejectsValidation(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"requester_id":0,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Create_MapsNotFound(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.SupplierQuoteRequest, _ []*procurement.SupplierQuoteRequestLine) (*procurement.SupplierQuoteRequest, error) {
			return nil, procurement.ErrRFQNotFound
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Create_MapsNoLines(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.SupplierQuoteRequest, _ []*procurement.SupplierQuoteRequestLine) (*procurement.SupplierQuoteRequest, error) {
			return nil, procurement.ErrRFQNoLines
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Create_MapsLineQty(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.SupplierQuoteRequest, _ []*procurement.SupplierQuoteRequestLine) (*procurement.SupplierQuoteRequest, error) {
			return nil, procurement.ErrRFQLineQty
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Create_MapsRequester(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.SupplierQuoteRequest, _ []*procurement.SupplierQuoteRequestLine) (*procurement.SupplierQuoteRequest, error) {
			return nil, procurement.ErrRFQRequester
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Create_ReturnsServerError(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.SupplierQuoteRequest, _ []*procurement.SupplierQuoteRequestLine) (*procurement.SupplierQuoteRequest, error) {
			return nil, errors.New("boom")
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreateFromRequest_CreatesRFQ(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return &procurement.PurchaseRequest{
					Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10),
					RequesterID: 5, State: procurement.RequestStateApproved,
				}, nil
			},
		},
	}
	requisitionLines := procurement.PurchaseRequestLineDAOMock{
		ListByRequestFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseRequestLine, error) {
			return []*procurement.PurchaseRequestLine{{Base: model.Base{ID: 100}, RequestID: 1, ItemID: ptrUint64(200), Qty: 2}}, nil
		},
	}
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, quoteRequest *procurement.SupplierQuoteRequest, _ []*procurement.SupplierQuoteRequestLine) (*procurement.SupplierQuoteRequest, error) {
			quoteRequest.ID = 1
			return quoteRequest, nil
		},
	}
	svc := procurement.NewTestSupplierQuoteRequestService(procurement.SupplierQuoteRequestServiceTestDeps{
		QuoteRequests:    quote_requests,
		RFQLines:         procurement.SupplierQuoteRequestLineDAOMock{},
		Quotes:           procurement.SupplierQuoteDAOMock{},
		QuoteLines:       procurement.SupplierQuoteLineDAOMock{},
		Requisitions:     requisitions,
		RequisitionLines: requisitionLines,
	})
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, svc)

	body := `{"request_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/from-request", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreateFromRequest_RejectsValidation(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"request_id":0}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/from-request", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreateFromRequest_MapsNotFound(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return nil, nil
			},
		},
	}
	svc := procurement.NewTestSupplierQuoteRequestService(procurement.SupplierQuoteRequestServiceTestDeps{
		QuoteRequests:    procurement.SupplierQuoteRequestDAOMock{},
		RFQLines:         procurement.SupplierQuoteRequestLineDAOMock{},
		Quotes:           procurement.SupplierQuoteDAOMock{},
		QuoteLines:       procurement.SupplierQuoteLineDAOMock{},
		Requisitions:     requisitions,
		RequisitionLines: procurement.PurchaseRequestLineDAOMock{},
	})
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, svc)

	body := `{"request_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/from-request", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreateFromRequest_MapsNotConvertible(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return &procurement.PurchaseRequest{
					Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10),
					RequesterID: 5, State: procurement.RequestStateConfirmed,
				}, nil
			},
		},
	}
	svc := procurement.NewTestSupplierQuoteRequestService(procurement.SupplierQuoteRequestServiceTestDeps{
		QuoteRequests:    procurement.SupplierQuoteRequestDAOMock{},
		RFQLines:         procurement.SupplierQuoteRequestLineDAOMock{},
		Quotes:           procurement.SupplierQuoteDAOMock{},
		QuoteLines:       procurement.SupplierQuoteLineDAOMock{},
		Requisitions:     requisitions,
		RequisitionLines: procurement.PurchaseRequestLineDAOMock{},
	})
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, svc)

	body := `{"request_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/from-request", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreateFromRequest_ReturnsServerError(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := procurement.NewTestSupplierQuoteRequestService(procurement.SupplierQuoteRequestServiceTestDeps{
		QuoteRequests:    procurement.SupplierQuoteRequestDAOMock{},
		RFQLines:         procurement.SupplierQuoteRequestLineDAOMock{},
		Quotes:           procurement.SupplierQuoteDAOMock{},
		QuoteLines:       procurement.SupplierQuoteLineDAOMock{},
		Requisitions:     requisitions,
		RequisitionLines: procurement.PurchaseRequestLineDAOMock{},
	})
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, svc)

	body := `{"request_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/from-request", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Send_SendsRFQ(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return sampleSupplierQuoteRequest(), nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/send", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Send_RejectsInvalidID(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/abc/send", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Send_MapsNotFound(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return nil, nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/send", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Send_MapsState(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return &procurement.SupplierQuoteRequest{
					Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10),
					RequesterID: 5, State: procurement.QuoteRequestStateSent,
				}, nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/send", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Send_ReturnsServerError(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/send", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Cancel_CancelsRFQ(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return sampleSupplierQuoteRequest(), nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Cancel_RejectsInvalidID(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/abc/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Cancel_MapsNotFound(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return nil, nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Cancel_MapsState(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return &procurement.SupplierQuoteRequest{
					Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10),
					RequesterID: 5, State: procurement.QuoteRequestStateDone,
				}, nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_Cancel_ReturnsServerError(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_ListLines_ReturnsLines(t *testing.T) {
	lines := procurement.SupplierQuoteRequestLineDAOMock{
		ListByRFQFunc: func(_ context.Context, _ uint64) ([]*procurement.SupplierQuoteRequestLine, error) {
			return []*procurement.SupplierQuoteRequestLine{sampleSupplierQuoteRequestLine()}, nil
		},
	}
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, lines, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, lines, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/1/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_ListLines_RejectsInvalidID(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/abc/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_ListLines_ReturnsServerError(t *testing.T) {
	lines := procurement.SupplierQuoteRequestLineDAOMock{
		ListByRFQFunc: func(_ context.Context, _ uint64) ([]*procurement.SupplierQuoteRequestLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, lines, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, lines, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/1/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_ListQuotes_ReturnsQuotes(t *testing.T) {
	quotes := procurement.SupplierQuoteDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuote]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierQuote], error) {
				return &query.Page[procurement.SupplierQuote]{Items: []*procurement.SupplierQuote{sampleSupplierQuote()}, Count: 1}, nil
			},
		},
	}
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/1/quotes", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_ListQuotes_RejectsInvalidID(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/abc/quotes", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_ListQuotes_ReturnsServerError(t *testing.T) {
	quotes := procurement.SupplierQuoteDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuote]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierQuote], error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/supplier-quote-requests/1/quotes", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_SubmitQuote_SubmitsQuote(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return &procurement.SupplierQuoteRequest{
					Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10),
					RequesterID: 5, CurrencyCode: ptrString("IDR"), State: procurement.QuoteRequestStateSent,
				}, nil
			},
		},
	}
	lines := procurement.SupplierQuoteRequestLineDAOMock{
		ListByRFQFunc: func(_ context.Context, _ uint64) ([]*procurement.SupplierQuoteRequestLine, error) {
			return []*procurement.SupplierQuoteRequestLine{sampleSupplierQuoteRequestLine()}, nil
		},
	}
	quotes := procurement.SupplierQuoteDAOMock{
		CreateWithLinesFunc: func(_ context.Context, quote *procurement.SupplierQuote, _ []*procurement.SupplierQuoteLine) (*procurement.SupplierQuote, error) {
			quote.ID = 1
			return quote, nil
		},
	}
	app := rfqTestApp(t, quote_requests, lines, quotes, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, lines, quotes, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"supplier_id":5,"lines":[{"quoteRequest_line_id":100,"qty":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/quotes", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_SubmitQuote_RejectsInvalidID(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"supplier_id":5,"lines":[{"quoteRequest_line_id":100,"qty":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/abc/quotes", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_SubmitQuote_RejectsValidation(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"supplier_id":5,"lines":[]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/quotes", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_SubmitQuote_MapsState(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return sampleSupplierQuoteRequest(), nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"supplier_id":5,"lines":[{"quoteRequest_line_id":100,"qty":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/quotes", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_SubmitQuote_ReturnsServerError(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	body := `{"supplier_id":5,"lines":[{"quoteRequest_line_id":100,"qty":2,"unit_price":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/quotes", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_AcceptQuote_AcceptsQuote(t *testing.T) {
	quotes := procurement.SupplierQuoteDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuote]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuote, error) {
				return sampleSupplierQuote(), nil
			},
		},
	}
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return &procurement.SupplierQuoteRequest{
					Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10),
					RequesterID: 5, State: procurement.QuoteRequestStateSent,
				}, nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/quotes/1/accept", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_AcceptQuote_RejectsInvalidSupplierQuoteID(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/quotes/abc/accept", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_AcceptQuote_MapsNotFound(t *testing.T) {
	quotes := procurement.SupplierQuoteDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuote]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuote, error) {
				return nil, nil
			},
		},
	}
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/quotes/1/accept", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_AcceptQuote_MapsQuoteState(t *testing.T) {
	quotes := procurement.SupplierQuoteDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuote]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuote, error) {
				return &procurement.SupplierQuote{
					Base: model.Base{ID: 1}, QuoteRequestID: 1, SupplierID: 10,
					State: procurement.SupplierQuoteStateDraft,
				}, nil
			},
		},
	}
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/quotes/1/accept", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_AcceptQuote_ReturnsServerError(t *testing.T) {
	quotes := procurement.SupplierQuoteDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuote]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuote, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/quotes/1/accept", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreatePurchaseOrder_CreatesOrder(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return &procurement.SupplierQuoteRequest{
					Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10),
					RequesterID: 5, State: procurement.QuoteRequestStateDone,
				}, nil
			},
		},
	}
	quotes := procurement.SupplierQuoteDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuote]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierQuote], error) {
				quote := sampleSupplierQuote()
				quote.State = procurement.SupplierQuoteStateAccepted
				return &query.Page[procurement.SupplierQuote]{Items: []*procurement.SupplierQuote{quote}, Count: 1}, nil
			},
		},
	}
	quoteLines := procurement.SupplierQuoteLineDAOMock{
		ListByQuoteFunc: func(_ context.Context, _ uint64) ([]*procurement.SupplierQuoteLine, error) {
			return []*procurement.SupplierQuoteLine{sampleSupplierQuoteLine()}, nil
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, quoteLines, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, quoteLines))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/purchase-order", "{}")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreatePurchaseOrder_CreatesOrderWithSupplierID(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return &procurement.SupplierQuoteRequest{
					Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10),
					RequesterID: 5, State: procurement.QuoteRequestStateDone,
				}, nil
			},
		},
	}
	quotes := procurement.SupplierQuoteDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuote]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierQuote], error) {
				quote := sampleSupplierQuote()
				quote.State = procurement.SupplierQuoteStateAccepted
				return &query.Page[procurement.SupplierQuote]{Items: []*procurement.SupplierQuote{quote}, Count: 1}, nil
			},
		},
	}
	quoteLines := procurement.SupplierQuoteLineDAOMock{
		ListByQuoteFunc: func(_ context.Context, _ uint64) ([]*procurement.SupplierQuoteLine, error) {
			return []*procurement.SupplierQuoteLine{sampleSupplierQuoteLine()}, nil
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, quoteLines, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, quoteLines))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/purchase-order", `{"supplier_id":10}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreatePurchaseOrder_RejectsMalformedBody(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/purchase-order", `{invalid-json}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreatePurchaseOrder_RejectsInvalidID(t *testing.T) {
	app := rfqTestApp(t, procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(procurement.SupplierQuoteRequestDAOMock{}, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/abc/purchase-order", "{}")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreatePurchaseOrder_MapsNotFound(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return nil, nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/purchase-order", "{}")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreatePurchaseOrder_MapsNotConvertible(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return sampleSupplierQuoteRequest(), nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/purchase-order", "{}")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreatePurchaseOrder_MapsNoAccepted(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return &procurement.SupplierQuoteRequest{
					Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10),
					RequesterID: 5, State: procurement.QuoteRequestStateDone,
				}, nil
			},
		},
	}
	quotes := procurement.SupplierQuoteDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuote]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierQuote], error) {
				return &query.Page[procurement.SupplierQuote]{Items: []*procurement.SupplierQuote{}, Count: 0}, nil
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, quotes, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/purchase-order", "{}")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSupplierQuoteRequestHandler_CreatePurchaseOrder_ReturnsServerError(t *testing.T) {
	quote_requests := procurement.SupplierQuoteRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.SupplierQuoteRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.SupplierQuoteRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := rfqTestApp(t, quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}, rfqTestSvc(quote_requests, procurement.SupplierQuoteRequestLineDAOMock{}, procurement.SupplierQuoteDAOMock{}, procurement.SupplierQuoteLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/supplier-quote-requests/1/purchase-order", "{}")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
