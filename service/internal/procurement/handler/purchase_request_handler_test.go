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

func requisitionTestApp(t *testing.T, requisitions procurement.PurchaseRequestDAOMock, lines procurement.PurchaseRequestLineDAOMock, svc procurement.PurchaseRequestService) *fiber.App {
	t.Helper()
	return procurementTestApp(t, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewPurchaseRequestHandler(svc)
		h.Register(api, guards)
	})
}

func requisitionTestSvc(requisitions procurement.PurchaseRequestDAOMock, lines procurement.PurchaseRequestLineDAOMock) procurement.PurchaseRequestService {
	return procurement.NewTestPurchaseRequestService(procurement.PurchaseRequestServiceTestDeps{Requisitions: requisitions, Lines: lines})
}

func TestPurchaseRequestHandler_List_ReturnsRequisitions(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.PurchaseRequest], error) {
				return &query.Page[procurement.PurchaseRequest]{Items: []*procurement.PurchaseRequest{samplePurchaseRequest()}, Count: 1}, nil
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-requests/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_List_RejectsInvalidQuery(t *testing.T) {
	app := requisitionTestApp(t, procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-requests/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_List_ReturnsUnauthorizedWhenTenantMissing(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.PurchaseRequest], error) {
				return &query.Page[procurement.PurchaseRequest]{Items: []*procurement.PurchaseRequest{}, Count: 0}, nil
			},
		},
	}
	svc := requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{})
	app := procurementTestAppNoTenant(t, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewPurchaseRequestHandler(svc)
		h.Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/purchase-requests/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_List_ReturnsServerError(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[procurement.PurchaseRequest], error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-requests/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Get_ReturnsRequisition(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return samplePurchaseRequest(), nil
			},
		},
	}
	lines := procurement.PurchaseRequestLineDAOMock{
		ListByRequestFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseRequestLine, error) {
			return []*procurement.PurchaseRequestLine{{Base: model.Base{ID: 100}, RequestID: 1, ItemID: ptrUint64(200), Qty: 2}}, nil
		},
	}
	app := requisitionTestApp(t, requisitions, lines, requisitionTestSvc(requisitions, lines))

	resp, err := doRequest(app, http.MethodGet, "/purchase-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Get_ReturnsNotFound(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return nil, nil
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Get_RejectsForeignTenant(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				requisition := samplePurchaseRequest()
				requisition.OrganizationID = ptrUint64(99)
				return requisition, nil
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Get_RejectsInvalidID(t *testing.T) {
	app := requisitionTestApp(t, procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-requests/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Get_ReturnsServerError(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/purchase-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Get_ReturnsServerErrorOnLines(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return samplePurchaseRequest(), nil
			},
		},
	}
	lines := procurement.PurchaseRequestLineDAOMock{
		ListByRequestFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseRequestLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := requisitionTestApp(t, requisitions, lines, requisitionTestSvc(requisitions, lines))

	resp, err := doRequest(app, http.MethodGet, "/purchase-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Create_CreatesRequisition(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, requisition *procurement.PurchaseRequest, _ []*procurement.PurchaseRequestLine) (*procurement.PurchaseRequest, error) {
			requisition.ID = 1
			return requisition, nil
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Create_RejectsValidation(t *testing.T) {
	app := requisitionTestApp(t, procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}))

	body := `{"requester_id":0,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Create_RejectsInvalidNeededBy(t *testing.T) {
	app := requisitionTestApp(t, procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}))

	body := `{"requester_id":5,"needed_by":"bad-date","lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Create_RejectsInvalidLineNeededBy(t *testing.T) {
	app := requisitionTestApp(t, procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2,"needed_by":"bad-date"}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Create_MapsNotFound(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseRequest, _ []*procurement.PurchaseRequestLine) (*procurement.PurchaseRequest, error) {
			return nil, procurement.ErrRequisitionNotFound
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Create_MapsNoLines(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseRequest, _ []*procurement.PurchaseRequestLine) (*procurement.PurchaseRequest, error) {
			return nil, procurement.ErrRequisitionNoLines
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	body := `{"requester_id":5,"lines":[]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Create_MapsLineQty(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseRequest, _ []*procurement.PurchaseRequestLine) (*procurement.PurchaseRequest, error) {
			return nil, procurement.ErrRequisitionLineQty
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Create_MapsRequester(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseRequest, _ []*procurement.PurchaseRequestLine) (*procurement.PurchaseRequest, error) {
			return nil, procurement.ErrRequisitionRequester
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Create_ReturnsServerError(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *procurement.PurchaseRequest, _ []*procurement.PurchaseRequestLine) (*procurement.PurchaseRequest, error) {
			return nil, errors.New("boom")
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	body := `{"requester_id":5,"lines":[{"item_id":100,"qty":2}]}`
	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Confirm_ConfirmsRequisition(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return samplePurchaseRequest(), nil
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Confirm_RejectsInvalidID(t *testing.T) {
	app := requisitionTestApp(t, procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/abc/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Confirm_MapsNotFound(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return nil, nil
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Confirm_MapsState(t *testing.T) {
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
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Confirm_ReturnsServerError(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Approve_ApprovesRequisition(t *testing.T) {
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
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/approve", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Approve_RejectsInvalidID(t *testing.T) {
	app := requisitionTestApp(t, procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/abc/approve", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Approve_MapsNotFound(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return nil, nil
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/approve", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Approve_MapsState(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return samplePurchaseRequest(), nil
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/approve", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Approve_ReturnsServerError(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/approve", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Cancel_CancelsRequisition(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return samplePurchaseRequest(), nil
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Cancel_RejectsInvalidID(t *testing.T) {
	app := requisitionTestApp(t, procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(procurement.PurchaseRequestDAOMock{}, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/abc/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Cancel_MapsNotFound(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return nil, nil
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Cancel_MapsState(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return &procurement.PurchaseRequest{
					Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10),
					RequesterID: 5, State: procurement.RequestStateDone,
				}, nil
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPurchaseRequestHandler_Cancel_ReturnsServerError(t *testing.T) {
	requisitions := procurement.PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := requisitionTestApp(t, requisitions, procurement.PurchaseRequestLineDAOMock{}, requisitionTestSvc(requisitions, procurement.PurchaseRequestLineDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/purchase-requests/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestWriteRequisitionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "requisition not found", err: procurement.ErrRequisitionNotFound, want: http.StatusNotFound},
		{name: "requisition state", err: procurement.ErrRequestState, want: http.StatusConflict},
		{name: "requisition no lines", err: procurement.ErrRequisitionNoLines, want: http.StatusUnprocessableEntity},
		{name: "requisition line qty", err: procurement.ErrRequisitionLineQty, want: http.StatusUnprocessableEntity},
		{name: "requisition requester", err: procurement.ErrRequisitionRequester, want: http.StatusUnprocessableEntity},
		{name: "unexpected error", err: errors.New("boom"), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/err", func(c fiber.Ctx) error {
				return writeRequisitionError(c, tt.err)
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
