package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crosscutting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func approvalTestApp(requests crosscutting.ApprovalRequestDAOMock, steps crosscutting.ApprovalStepDAOMock, svc crosscutting.ApprovalService) *fiber.App {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(model.ActorKey, uint64(5))
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewApprovalRequestHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func approvalTestAppNoCaller(requests crosscutting.ApprovalRequestDAOMock, steps crosscutting.ApprovalStepDAOMock, svc crosscutting.ApprovalService) *fiber.App {
	app := fiber.New()
	h := NewApprovalRequestHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func attachmentTestApp(attachments crosscutting.AttachmentDAOMock, svc crosscutting.AttachmentService) *fiber.App {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(model.ActorKey, uint64(5))
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewAttachmentHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func attachmentTestAppNoCaller(attachments crosscutting.AttachmentDAOMock, svc crosscutting.AttachmentService) *fiber.App {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewAttachmentHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func messageTestApp(messages crosscutting.MessageDAOMock, svc crosscutting.MessageService) *fiber.App {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(model.ActorKey, uint64(5))
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewMessageHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func messageTestAppNoCaller(messages crosscutting.MessageDAOMock, svc crosscutting.MessageService) *fiber.App {
	app := fiber.New()
	h := NewMessageHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func messageTestAppZeroCaller(messages crosscutting.MessageDAOMock, svc crosscutting.MessageService) *fiber.App {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(model.ActorKey, uint64(0))
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewMessageHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func sampleApprovalRequest() *crosscutting.ApprovalRequest {
	return &crosscutting.ApprovalRequest{
		Base:           model.Base{ID: 1},
		OrganizationID: 10,
		OwnerType:      "purchase_order",
		OwnerID:        10,
		RequestedBy:    2,
		State:          crosscutting.ApprovalStatePending,
	}
}

func TestApprovalRequestHandler_List_ReturnsRequests(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{
		CRUDMock: daoCRUDMockList(sampleApprovalRequest()),
	}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/approval-requests/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_List_RejectsInvalidQuery(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/approval-requests/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_List_ReturnsServerError(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{
		CRUDMock: daoCRUDMockListError(crosscutting.ApprovalRequest{}),
	}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/approval-requests/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Get_ReturnsRequest(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{
		CRUDMock: daoCRUDMockFind(sampleApprovalRequest()),
	}
	steps := crosscutting.ApprovalStepDAOMock{
		ListByRequestFunc: func(_ context.Context, _ uint64) ([]*crosscutting.ApprovalStep, error) {
			return []*crosscutting.ApprovalStep{{Base: model.Base{ID: 1}, ApproverID: 5, Sequence: 10}}, nil
		},
	}
	svc := crosscutting.NewApprovalService(requests, steps, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, steps, svc)

	resp, err := doRequest(app, http.MethodGet, "/approval-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Get_ReturnsNotFound(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{
		CRUDMock: daoCRUDMockFindNil(crosscutting.ApprovalRequest{}),
	}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/approval-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Get_RejectsInvalidID(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/approval-requests/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Get_ReturnsServerError(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{
		CRUDMock: daoCRUDMockFindError(crosscutting.ApprovalRequest{}),
	}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/approval-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Get_ReturnsServerErrorOnSteps(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{
		CRUDMock: daoCRUDMockFind(sampleApprovalRequest()),
	}
	steps := crosscutting.ApprovalStepDAOMock{
		ListByRequestFunc: func(_ context.Context, _ uint64) ([]*crosscutting.ApprovalStep, error) {
			return nil, errors.New("db down")
		},
	}
	svc := crosscutting.NewApprovalService(requests, steps, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, steps, svc)

	resp, err := doRequest(app, http.MethodGet, "/approval-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Get_ReturnsNotFoundForForeignOrganization(t *testing.T) {
	foreign := sampleApprovalRequest()
	foreign.OrganizationID = 8
	requests := crosscutting.ApprovalRequestDAOMock{
		CRUDMock: daoCRUDMockFind(foreign),
	}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/approval-requests/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Create_RequiresOrganization(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestAppNoCaller(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	body := `{"owner_type":"purchase_order","owner_id":10,"requested_by":2,"approver_ids":[3]}`
	resp, err := doRequest(app, http.MethodPost, "/approval-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Create_CreatesRequest(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{
		CreateWithStepsFunc: func(_ context.Context, request *crosscutting.ApprovalRequest, _ []*crosscutting.ApprovalStep) (*crosscutting.ApprovalRequest, error) {
			request.ID = 1
			return request, nil
		},
	}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	body := `{"owner_type":"purchase_order","owner_id":10,"requested_by":2,"approver_ids":[3,4]}`
	resp, err := doRequest(app, http.MethodPost, "/approval-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Create_RejectsValidation(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	body := `{"owner_type":"purchase_order","owner_id":0,"requested_by":2,"approver_ids":[3]}`
	resp, err := doRequest(app, http.MethodPost, "/approval-requests/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Create_MapsServiceErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"not found", crosscutting.ErrApprovalNotFound, http.StatusNotFound},
		{"step not found", crosscutting.ErrApprovalStepNotFound, http.StatusNotFound},
		{"no approvers", crosscutting.ErrApprovalNoApprovers, http.StatusUnprocessableEntity},
		{"owner", crosscutting.ErrApprovalOwner, http.StatusUnprocessableEntity},
		{"state", crosscutting.ErrApprovalState, http.StatusConflict},
		{"not approver", crosscutting.ErrApprovalNotApprover, http.StatusConflict},
		{"order", crosscutting.ErrApprovalOrder, http.StatusConflict},
		{"unexpected", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requests := crosscutting.ApprovalRequestDAOMock{
				CreateWithStepsFunc: func(_ context.Context, _ *crosscutting.ApprovalRequest, _ []*crosscutting.ApprovalStep) (*crosscutting.ApprovalRequest, error) {
					return nil, tc.err
				},
			}
			svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
			app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

			body := `{"owner_type":"purchase_order","owner_id":10,"requested_by":2,"approver_ids":[3]}`
			resp, err := doRequest(app, http.MethodPost, "/approval-requests/", body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.want)
			}
		})
	}
}

func TestApprovalRequestHandler_Decide_DecidesStep(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{
		CRUDMock: daoCRUDMockFind(sampleApprovalRequest()),
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, request *crosscutting.ApprovalRequest) (*crosscutting.ApprovalRequest, error) {
			return request, nil
		},
	}
	steps := crosscutting.ApprovalStepDAOMock{
		ListByRequestFunc: func(_ context.Context, _ uint64) ([]*crosscutting.ApprovalStep, error) {
			return []*crosscutting.ApprovalStep{{Base: model.Base{ID: 1}, ApproverID: 5, Sequence: 10, Decision: crosscutting.ApprovalStepPending}}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, step *crosscutting.ApprovalStep) (*crosscutting.ApprovalStep, error) {
			return step, nil
		},
	}
	svc := crosscutting.NewApprovalService(requests, steps, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, steps, svc)

	body := `{"step_id":1,"approve":true,"comment":"ok"}`
	resp, err := doRequest(app, http.MethodPost, "/approval-requests/1/decide", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Decide_RejectsInvalidID(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	body := `{"step_id":1,"approve":true}`
	resp, err := doRequest(app, http.MethodPost, "/approval-requests/abc/decide", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Decide_RejectsValidation(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	body := `{"step_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/approval-requests/1/decide", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Decide_ReturnsUnauthorized(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestAppNoCaller(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	body := `{"step_id":1,"approve":true}`
	resp, err := doRequest(app, http.MethodPost, "/approval-requests/1/decide", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Decide_ReturnsNotFound(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{
		CRUDMock: daoCRUDMockFindNil(crosscutting.ApprovalRequest{}),
	}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	body := `{"step_id":1,"approve":true}`
	resp, err := doRequest(app, http.MethodPost, "/approval-requests/1/decide", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Decide_ReturnsConflictOnState(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{
		CRUDMock: daoCRUDMockFind(&crosscutting.ApprovalRequest{Base: model.Base{ID: 1}, OrganizationID: 10, State: crosscutting.ApprovalStateApproved}),
	}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	body := `{"step_id":1,"approve":true}`
	resp, err := doRequest(app, http.MethodPost, "/approval-requests/1/decide", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestApprovalRequestHandler_Decide_ReturnsServerError(t *testing.T) {
	requests := crosscutting.ApprovalRequestDAOMock{
		CRUDMock: daoCRUDMockFindError(crosscutting.ApprovalRequest{}),
	}
	svc := crosscutting.NewApprovalService(requests, crosscutting.ApprovalStepDAOMock{}, crosscutting.TransactionerMock{})
	app := approvalTestApp(requests, crosscutting.ApprovalStepDAOMock{}, svc)

	body := `{"step_id":1,"approve":true}`
	resp, err := doRequest(app, http.MethodPost, "/approval-requests/1/decide", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAttachmentHandler_List_ReturnsAttachments(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockList(&crosscutting.Attachment{Base: model.Base{ID: 1}, OrganizationID: 10, OwnerType: "purchase_order", OwnerID: 10, Filename: "scan.pdf"}),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAttachmentHandler_List_RejectsInvalidQuery(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttachmentHandler_List_ReturnsServerError(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockListError(crosscutting.Attachment{}),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAttachmentHandler_Get_ReturnsAttachment(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockFind(&crosscutting.Attachment{Base: model.Base{ID: 1}, OrganizationID: 10, OwnerType: "purchase_order", OwnerID: 10, Filename: "scan.pdf"}),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAttachmentHandler_Get_ReturnsNotFound(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockFindNil(crosscutting.Attachment{}),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAttachmentHandler_Get_RejectsInvalidID(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttachmentHandler_Get_ReturnsServerError(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockFindError(crosscutting.Attachment{}),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAttachmentHandler_Get_ReturnsNotFoundForForeignOrganization(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockFind(&crosscutting.Attachment{Base: model.Base{ID: 1}, OrganizationID: 8, OwnerType: "purchase_order", OwnerID: 10, Filename: "scan.pdf"}),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAttachmentHandler_Download_DownloadsFile(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockFind(&crosscutting.Attachment{Base: model.Base{ID: 1}, OrganizationID: 10, Filename: "scan.pdf", MimeType: "application/pdf", StorageURL: "attachments/10/abc"}),
	}
	store := crosscutting.StoreMock{
		OpenFunc: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("content")), nil
		},
	}
	svc := crosscutting.NewAttachmentService(attachments, store)
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/1/download", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAttachmentHandler_Download_ReturnsNotFound(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockFindNil(crosscutting.Attachment{}),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/1/download", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAttachmentHandler_Download_RejectsInvalidID(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/abc/download", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttachmentHandler_Download_ReturnsServerError(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockFind(&crosscutting.Attachment{Base: model.Base{ID: 1}, OrganizationID: 10, StorageURL: "attachments/10/abc"}),
	}
	store := crosscutting.StoreMock{
		OpenFunc: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return nil, errors.New("storage down")
		},
	}
	svc := crosscutting.NewAttachmentService(attachments, store)
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodGet, "/attachments/1/download", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAttachmentHandler_Upload_UploadsFile(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockCreate(&crosscutting.Attachment{Base: model.Base{ID: 1}}),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doMultipartUpload(app, "purchase_order", "10", "scan.pdf", "content")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestAttachmentHandler_Upload_UploadsFileWithoutCaller(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockCreate(&crosscutting.Attachment{Base: model.Base{ID: 1}}),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestAppNoCaller(attachments, svc)

	resp, err := doMultipartUpload(app, "purchase_order", "10", "scan.pdf", "content")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestAttachmentHandler_Upload_MissingFileField(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("owner_type", "purchase_order")
	_ = writer.WriteField("owner_id", "10")
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/attachments/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttachmentHandler_Upload_RejectsInvalidOwnerID(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doMultipartUpload(app, "purchase_order", "abc", "scan.pdf", "content")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttachmentHandler_Upload_RejectsEmptyContent(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doMultipartUpload(app, "purchase_order", "10", "empty.pdf", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttachmentHandler_Upload_RejectsEmptyFilename(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doMultipartUpload(app, "purchase_order", "10", "", "content")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttachmentHandler_Upload_MapsServiceErrors(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockCreateError(crosscutting.Attachment{}, crosscutting.ErrAttachmentNotFound),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doMultipartUpload(app, "purchase_order", "10", "scan.pdf", "content")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAttachmentHandler_Upload_ReturnsServerError(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockCreateError(crosscutting.Attachment{}, errors.New("boom")),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doMultipartUpload(app, "purchase_order", "10", "scan.pdf", "content")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAttachmentHandler_Delete_DeletesAttachment(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockFind(&crosscutting.Attachment{Base: model.Base{ID: 1}, OrganizationID: 10, StorageURL: "attachments/10/abc"}),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodDelete, "/attachments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestAttachmentHandler_Delete_ReturnsNotFound(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockFindNil(crosscutting.Attachment{}),
	}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodDelete, "/attachments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAttachmentHandler_Delete_RejectsInvalidID(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{}
	svc := crosscutting.NewAttachmentService(attachments, crosscutting.StoreMock{})
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodDelete, "/attachments/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttachmentHandler_Delete_ReturnsServerError(t *testing.T) {
	attachments := crosscutting.AttachmentDAOMock{
		CRUDMock: daoCRUDMockFind(&crosscutting.Attachment{Base: model.Base{ID: 1}, OrganizationID: 10, StorageURL: "attachments/10/abc"}),
	}
	store := crosscutting.StoreMock{
		DeleteFunc: func(_ context.Context, _ string) error {
			return errors.New("storage down")
		},
	}
	svc := crosscutting.NewAttachmentService(attachments, store)
	app := attachmentTestApp(attachments, svc)

	resp, err := doRequest(app, http.MethodDelete, "/attachments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestMessageHandler_List_ReturnsMessages(t *testing.T) {
	messages := crosscutting.MessageDAOMock{
		CRUDMock: daoCRUDMockList(&crosscutting.Message{Base: model.Base{ID: 1}, OrganizationID: 10, OwnerType: "purchase_order", OwnerID: 10, Body: "hello"}),
	}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestApp(messages, svc)

	resp, err := doRequest(app, http.MethodGet, "/messages/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestMessageHandler_List_RejectsInvalidQuery(t *testing.T) {
	messages := crosscutting.MessageDAOMock{}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestApp(messages, svc)

	resp, err := doRequest(app, http.MethodGet, "/messages/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestMessageHandler_List_ReturnsServerError(t *testing.T) {
	messages := crosscutting.MessageDAOMock{
		CRUDMock: daoCRUDMockListError(crosscutting.Message{}),
	}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestApp(messages, svc)

	resp, err := doRequest(app, http.MethodGet, "/messages/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestMessageHandler_Get_ReturnsMessage(t *testing.T) {
	messages := crosscutting.MessageDAOMock{
		CRUDMock: daoCRUDMockFind(&crosscutting.Message{Base: model.Base{ID: 1}, OrganizationID: 10, OwnerType: "purchase_order", OwnerID: 10, Body: "hello"}),
	}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestApp(messages, svc)

	resp, err := doRequest(app, http.MethodGet, "/messages/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestMessageHandler_Get_ReturnsNotFound(t *testing.T) {
	messages := crosscutting.MessageDAOMock{
		CRUDMock: daoCRUDMockFindNil(crosscutting.Message{}),
	}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestApp(messages, svc)

	resp, err := doRequest(app, http.MethodGet, "/messages/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestMessageHandler_Get_RejectsInvalidID(t *testing.T) {
	messages := crosscutting.MessageDAOMock{}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestApp(messages, svc)

	resp, err := doRequest(app, http.MethodGet, "/messages/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestMessageHandler_Get_ReturnsServerError(t *testing.T) {
	messages := crosscutting.MessageDAOMock{
		CRUDMock: daoCRUDMockFindError(crosscutting.Message{}),
	}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestApp(messages, svc)

	resp, err := doRequest(app, http.MethodGet, "/messages/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestMessageHandler_Get_ReturnsNotFoundForForeignOrganization(t *testing.T) {
	messages := crosscutting.MessageDAOMock{
		CRUDMock: daoCRUDMockFind(&crosscutting.Message{Base: model.Base{ID: 1}, OrganizationID: 8, OwnerType: "purchase_order", OwnerID: 10, Body: "hello"}),
	}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestApp(messages, svc)

	resp, err := doRequest(app, http.MethodGet, "/messages/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestMessageHandler_Create_CreatesMessage(t *testing.T) {
	messages := crosscutting.MessageDAOMock{
		CRUDMock: daoCRUDMockCreate(&crosscutting.Message{Base: model.Base{ID: 1}}),
	}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestApp(messages, svc)

	body := `{"owner_type":"purchase_order","owner_id":10,"body":"please expedite","message_type":"note"}`
	resp, err := doRequest(app, http.MethodPost, "/messages/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestMessageHandler_Create_RejectsValidation(t *testing.T) {
	messages := crosscutting.MessageDAOMock{}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestApp(messages, svc)

	body := `{"owner_type":"purchase_order","owner_id":10,"body":""}`
	resp, err := doRequest(app, http.MethodPost, "/messages/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestMessageHandler_Create_ReturnsUnauthorized(t *testing.T) {
	messages := crosscutting.MessageDAOMock{}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestAppNoCaller(messages, svc)

	body := `{"owner_type":"purchase_order","owner_id":10,"body":"hello"}`
	resp, err := doRequest(app, http.MethodPost, "/messages/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestMessageHandler_Create_MapsAuthorMissing(t *testing.T) {
	messages := crosscutting.MessageDAOMock{}
	svc := crosscutting.NewMessageService(messages)
	app := messageTestAppZeroCaller(messages, svc)

	body := `{"owner_type":"purchase_order","owner_id":10,"body":"hello"}`
	resp, err := doRequest(app, http.MethodPost, "/messages/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestMessageHandler_Create_MapsServiceErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"not found", crosscutting.ErrMessageNotFound, http.StatusNotFound},
		{"owner", crosscutting.ErrMessageOwner, http.StatusUnprocessableEntity},
		{"body", crosscutting.ErrMessageBody, http.StatusUnprocessableEntity},
		{"unexpected", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			messages := crosscutting.MessageDAOMock{
				CRUDMock: daoCRUDMockCreateError(crosscutting.Message{}, tc.err),
			}
			svc := crosscutting.NewMessageService(messages)
			app := messageTestApp(messages, svc)

			body := `{"owner_type":"purchase_order","owner_id":10,"body":"hello"}`
			resp, err := doRequest(app, http.MethodPost, "/messages/", body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.want)
			}
		})
	}
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

func doMultipartUpload(app *fiber.App, ownerType, ownerID, filename, content string) (*http.Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("owner_type", ownerType)
	_ = writer.WriteField("owner_id", ownerID)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write([]byte(content)); err != nil {
		return nil, err
	}
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/attachments/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return app.Test(req)
}

func daoCRUDMockList[E any](item *E) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[E], error) {
			return &query.Page[E]{Items: []*E{item}, Count: 1}, nil
		},
	}
}

func daoCRUDMockListError[E any](_ E) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[E], error) {
			return nil, errors.New("db down")
		},
	}
}

func daoCRUDMockFind[E any](item *E) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		FindFunc: func(_ context.Context, _ uint64) (*E, error) {
			return item, nil
		},
	}
}

func daoCRUDMockFindNil[E any](_ E) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		FindFunc: func(_ context.Context, _ uint64) (*E, error) {
			return nil, nil
		},
	}
}

func daoCRUDMockFindError[E any](_ E) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		FindFunc: func(_ context.Context, _ uint64) (*E, error) {
			return nil, errors.New("db down")
		},
	}
}

func daoCRUDMockCreate[E any](item *E) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		CreateFunc: func(_ context.Context, entity *E) (*E, error) {
			return item, nil
		},
	}
}

func daoCRUDMockCreateError[E any](_ E, err error) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		CreateFunc: func(_ context.Context, _ *E) (*E, error) {
			return nil, err
		},
	}
}
