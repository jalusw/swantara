package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crosscutting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type MessageHandler struct {
	svc crosscutting.MessageService
}

func NewMessageHandler(svc crosscutting.MessageService) MessageHandler {
	return MessageHandler{svc: svc}
}

type MessageResponse struct {
	ID             uint64 `json:"id"`
	OrganizationID uint64 `json:"organization_id"`
	OwnerType      string `json:"owner_type"`
	OwnerID        uint64 `json:"owner_id"`
	AuthorID       uint64 `json:"author_id"`
	Body           string `json:"body"`
	MessageType    string `json:"message_type"`
}

func newMessageResponse(message *crosscutting.Message) MessageResponse {
	return MessageResponse{
		ID:             message.ID,
		OrganizationID: message.OrganizationID,
		OwnerType:      message.OwnerType,
		OwnerID:        message.OwnerID,
		AuthorID:       message.AuthorID,
		Body:           message.Body,
		MessageType:    message.MessageType,
	}
}

var messageQueryAllowlist = map[string]struct{}{
	"owner_type":   {},
	"owner_id":     {},
	"author_id":    {},
	"message_type": {},
	"created_at":   {},
}

type ListMessagesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListMessagesResponse `json:"data"`
}
type ListMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}

// @Summary List messages
// @Description Lists messages with pagination, sorting, and filtering. Filters by owner type, owner id, author, and message type are supported so discussions on a resource can be browsed.
// @Tags Messages
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListMessagesResponseEnvelope "Messages retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /messages [get]
func (h MessageHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, messageQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("message list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve messages.", err)
	}
	items := make([]MessageResponse, len(page.Items))
	for i, message := range page.Items {
		items[i] = newMessageResponse(message)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Messages retrieved successfully.", ListMessagesResponse{
		Messages: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetMessageResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetMessageResponse `json:"data"`
}
type GetMessageResponse struct {
	Message MessageResponse `json:"message"`
}

// @Summary Get message
// @Description Gets a single message by id, including its owning resource, author, body, and message type.
// @Tags Messages
// @Accept json
// @Produce json
// @Param id path integer true "Message ID"
// @Success 200 {object} GetMessageResponseEnvelope "Message retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Message not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid message id"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /messages/{id} [get]
func (h MessageHandler) Get(c fiber.Ctx) error {
	messageID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid message id provided.", nil)
	}
	message, err := h.svc.Find(c, messageID)
	if err != nil {
		httpx.RequestLog(c).Error("message lookup failed", "message_id", messageID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get message.", err)
	}
	if message == nil || !httpx.OwnsTenant(c, &message.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Message not found.")
	}
	return httpx.CreateSuccessResponse(c, "Message retrieved successfully.", GetMessageResponse{
		Message: newMessageResponse(message),
	})
}

type CreateMessageRequest struct {
	OwnerType   string `json:"owner_type" validate:"required"`
	OwnerID     uint64 `json:"owner_id" validate:"required,gt=0"`
	Body        string `json:"body" validate:"required"`
	MessageType string `json:"message_type"`
}

type CreateMessageResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateMessageResponse `json:"data"`
}
type CreateMessageResponse struct {
	Message MessageResponse `json:"message"`
}

// @Summary Create message
// @Description Creates a message on an owner resource with the authenticated user recorded as the author. A non-empty body and a valid owner are required.
// @Tags Messages
// @Accept json
// @Produce json
// @Param body body CreateMessageRequest true "Message details"
// @Success 201 {object} CreateMessageResponseEnvelope "Message created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /messages [post]
func (h MessageHandler) Create(c fiber.Ctx) error {
	var request CreateMessageRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	authorID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Authenticated user is required to post a message.", nil)
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}
	created, err := h.svc.Post(c, *organizationID, request.OwnerType, request.OwnerID, authorID, request.Body, request.MessageType)
	if err != nil {
		return writeMessageError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Message created successfully.", CreateMessageResponse{
		Message: newMessageResponse(created),
	})
}

func writeMessageError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, crosscutting.ErrMessageOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	case errors.Is(err, crosscutting.ErrMessageNotFound):
		return httpx.CreateNotFoundResponse(c, "Message not found.")
	case errors.Is(err, crosscutting.ErrMessageOwner):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Message owner is required.", nil)
	case errors.Is(err, crosscutting.ErrMessageAuthor):
		return httpx.CreateUnauthorizedErrorResponse(c, "Authenticated user is required to post a message.", nil)
	case errors.Is(err, crosscutting.ErrMessageBody):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Message body is required.", nil)
	default:
		httpx.RequestLog(c).Error("message write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process message.", err)
	}
}
