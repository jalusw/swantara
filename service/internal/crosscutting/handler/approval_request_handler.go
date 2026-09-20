package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crosscutting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ApprovalRequestHandler struct {
	svc crosscutting.ApprovalService
}

func NewApprovalRequestHandler(svc crosscutting.ApprovalService) ApprovalRequestHandler {
	return ApprovalRequestHandler{svc: svc}
}

type ApprovalStepResponse struct {
	ID         uint64     `json:"id"`
	ApproverID uint64     `json:"approver_id"`
	Sequence   int        `json:"sequence"`
	Decision   string     `json:"decision"`
	DecidedAt  *time.Time `json:"decided_at"`
	Comment    *string    `json:"comment"`
}

type ApprovalRequestResponse struct {
	ID             uint64                 `json:"id"`
	OrganizationID uint64                 `json:"organization_id"`
	OwnerType      string                 `json:"owner_type"`
	OwnerID        uint64                 `json:"owner_id"`
	RequestedBy    uint64                 `json:"requested_by"`
	State          string                 `json:"state"`
	Steps          []ApprovalStepResponse `json:"steps,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

func newApprovalStepResponse(step *crosscutting.ApprovalStep) ApprovalStepResponse {
	return ApprovalStepResponse{
		ID:         step.ID,
		ApproverID: step.ApproverID,
		Sequence:   step.Sequence,
		Decision:   step.Decision,
		DecidedAt:  step.DecidedAt,
		Comment:    step.Comment,
	}
}

func newApprovalRequestResponse(request *crosscutting.ApprovalRequest) ApprovalRequestResponse {
	return ApprovalRequestResponse{
		ID:             request.ID,
		OrganizationID: request.OrganizationID,
		OwnerType:      request.OwnerType,
		OwnerID:        request.OwnerID,
		RequestedBy:    request.RequestedBy,
		State:          request.State,
		CreatedAt:      request.CreatedAt,
		UpdatedAt:      request.UpdatedAt,
	}
}

var approvalRequestQueryAllowlist = map[string]struct{}{
	"owner_type":   {},
	"owner_id":     {},
	"requested_by": {},
	"state":        {},
	"created_at":   {},
	"updated_at":   {},
}

type ListApprovalRequestsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListApprovalRequestsResponse `json:"data"`
}
type ListApprovalRequestsResponse struct {
	Requests []ApprovalRequestResponse `json:"approval_requests"`
}

// @Summary List approval requests
// @Description Lists approval requests with pagination, sorting, and filtering. Filters by owner type, owner id, requester, state, and timestamps are supported so pending and resolved requests can be tracked.
// @Tags Approval Requests
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListApprovalRequestsResponseEnvelope "Approval requests retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /approval-requests [get]
func (h ApprovalRequestHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, approvalRequestQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("approval request list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve approval requests.", err)
	}

	items := make([]ApprovalRequestResponse, len(page.Items))
	for i, request := range page.Items {
		items[i] = newApprovalRequestResponse(request)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Approval requests retrieved successfully.", ListApprovalRequestsResponse{
		Requests: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetApprovalRequestResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetApprovalRequestResponse `json:"data"`
}
type GetApprovalRequestResponse struct {
	Request ApprovalRequestResponse `json:"approval_request"`
}

// @Summary Get approval request
// @Description Gets a single approval request by id, including its ordered approval steps with each approver's decision, decision time, and comment.
// @Tags Approval Requests
// @Accept json
// @Produce json
// @Param id path integer true "Approval Request ID"
// @Success 200 {object} GetApprovalRequestResponseEnvelope "Approval request retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Approval request not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid approval request id"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /approval-requests/{id} [get]
func (h ApprovalRequestHandler) Get(c fiber.Ctx) error {
	requestID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid approval request id provided.", nil)
	}

	request, err := h.svc.Find(c, requestID)
	if err != nil {
		httpx.RequestLog(c).Error("approval request lookup failed", "request_id", requestID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get approval request.", err)
	}
	if request == nil || !httpx.OwnsTenant(c, &request.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Approval request not found.")
	}

	steps, err := h.svc.ListSteps(c, requestID)
	if err != nil {
		httpx.RequestLog(c).Error("approval step list failed", "request_id", requestID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get approval request.", err)
	}

	response := newApprovalRequestResponse(request)
	response.Steps = make([]ApprovalStepResponse, len(steps))
	for i, step := range steps {
		response.Steps[i] = newApprovalStepResponse(step)
	}

	return httpx.CreateSuccessResponse(c, "Approval request retrieved successfully.", GetApprovalRequestResponse{
		Request: response,
	})
}

type CreateApprovalRequestRequest struct {
	OwnerType   string   `json:"owner_type" validate:"required"`
	OwnerID     uint64   `json:"owner_id" validate:"required,gt=0"`
	RequestedBy uint64   `json:"requested_by" validate:"required,gt=0"`
	ApproverIDs []uint64 `json:"approver_ids" validate:"required,min=1"`
}

type CreateApprovalRequestResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateApprovalRequestResponse `json:"data"`
}
type CreateApprovalRequestResponse struct {
	Request ApprovalRequestResponse `json:"approval_request"`
}

// @Summary Create approval request
// @Description Creates an approval request for a resource with an ordered list of approvers, generating one approval step per approver in the given sequence. Requires a valid owner and at least one approver.
// @Tags Approval Requests
// @Accept json
// @Produce json
// @Param body body CreateApprovalRequestRequest true "Approval request details"
// @Success 201 {object} CreateApprovalRequestResponseEnvelope "Approval request created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /approval-requests [post]
func (h ApprovalRequestHandler) Create(c fiber.Ctx) error {
	var request CreateApprovalRequestRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	created, err := h.svc.Create(c, *organizationID, request.OwnerType, request.OwnerID, request.RequestedBy, request.ApproverIDs)
	if err != nil {
		return writeApprovalError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Approval request created successfully.", CreateApprovalRequestResponse{
		Request: newApprovalRequestResponse(created),
	})
}

type DecideApprovalRequestRequest struct {
	StepID  uint64 `json:"step_id" validate:"required,gt=0"`
	Approve *bool  `json:"approve" validate:"required"`
	Comment string `json:"comment"`
}

type DecideApprovalRequestResponseEnvelope struct {
	httpx.EnvelopeBase
	Data DecideApprovalRequestResponse `json:"data"`
}
type DecideApprovalRequestResponse struct {
	Request ApprovalRequestResponse `json:"approval_request"`
}

// @Summary Decide approval step
// @Description Decides on a specific approval step by approving or refusing it as the step's approver, optionally adding a comment. Steps must be decided in sequence and only by the assigned approver, and invalid or already-decided steps are rejected with a conflict.
// @Tags Approval Requests
// @Accept json
// @Produce json
// @Param id path integer true "Approval Request ID"
// @Param body body DecideApprovalRequestRequest true "Decision details"
// @Success 200 {object} DecideApprovalRequestResponseEnvelope "Approval step decided successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Approval request or step not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /approval-requests/{id}/decide [post]
func (h ApprovalRequestHandler) Decide(c fiber.Ctx) error {
	requestID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid approval request id provided.", nil)
	}

	var request DecideApprovalRequestRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	approverID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Authenticated user is required to decide an approval step.", nil)
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	updated, err := h.svc.Decide(c, *organizationID, requestID, request.StepID, approverID, *request.Approve, request.Comment)
	if err != nil {
		return writeApprovalError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Approval step decided successfully.", DecideApprovalRequestResponse{
		Request: newApprovalRequestResponse(updated),
	})
}

func writeApprovalError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, crosscutting.ErrApprovalOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	case errors.Is(err, crosscutting.ErrApprovalNotFound):
		return httpx.CreateNotFoundResponse(c, "Approval request not found.")
	case errors.Is(err, crosscutting.ErrApprovalStepNotFound):
		return httpx.CreateNotFoundResponse(c, "Approval step not found.")
	case errors.Is(err, crosscutting.ErrApprovalNoApprovers):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Approval request must have at least one approver.", nil)
	case errors.Is(err, crosscutting.ErrApprovalOwner):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Approval request owner is required.", nil)
	case errors.Is(err, crosscutting.ErrApprovalState):
		return httpx.CreateConflictResponse(c, "Approval request cannot be decided in its current state.", err)
	case errors.Is(err, crosscutting.ErrApprovalNotApprover):
		return httpx.CreateConflictResponse(c, "Approval step does not belong to the actor.", err)
	case errors.Is(err, crosscutting.ErrApprovalOrder):
		return httpx.CreateConflictResponse(c, "Approval step must be decided after its predecessors.", err)
	default:
		httpx.RequestLog(c).Error("approval request write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process approval request.", err)
	}
}
