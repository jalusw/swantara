package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crosscutting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type AttachmentHandler struct {
	svc crosscutting.AttachmentService
}

func NewAttachmentHandler(svc crosscutting.AttachmentService) AttachmentHandler {
	return AttachmentHandler{svc: svc}
}

type AttachmentResponse struct {
	ID             uint64 `json:"id"`
	OrganizationID uint64 `json:"organization_id"`
	OwnerType      string `json:"owner_type"`
	OwnerID        uint64 `json:"owner_id"`
	Filename       string `json:"filename"`
	MimeType       string `json:"mime_type"`
	ByteSize       int64  `json:"byte_size"`
	Checksum       string `json:"checksum"`
	UploadedBy     uint64 `json:"uploaded_by"`
}

func newAttachmentResponse(attachment *crosscutting.Attachment) AttachmentResponse {
	return AttachmentResponse{
		ID:             attachment.ID,
		OrganizationID: attachment.OrganizationID,
		OwnerType:      attachment.OwnerType,
		OwnerID:        attachment.OwnerID,
		Filename:       attachment.Filename,
		MimeType:       attachment.MimeType,
		ByteSize:       attachment.ByteSize,
		Checksum:       attachment.Checksum,
		UploadedBy:     attachment.UploadedBy,
	}
}

var attachmentQueryAllowlist = map[string]struct{}{
	"owner_type": {},
	"owner_id":   {},
	"filename":   {},
	"mime_type":  {},
	"created_at": {},
}

type ListAttachmentsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListAttachmentsResponse `json:"data"`
}
type ListAttachmentsResponse struct {
	Attachments []AttachmentResponse `json:"attachments"`
}

// @Summary List attachments
// @Description Lists attachment metadata with pagination, sorting, and filtering. Filters by owner type, owner id, filename, and mime type are supported so files attached to a resource can be browsed.
// @Tags Attachments
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListAttachmentsResponseEnvelope "Attachments retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /attachments [get]
func (h AttachmentHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, attachmentQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("attachment list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve attachments.", err)
	}
	items := make([]AttachmentResponse, len(page.Items))
	for i, attachment := range page.Items {
		items[i] = newAttachmentResponse(attachment)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Attachments retrieved successfully.", ListAttachmentsResponse{
		Attachments: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetAttachmentResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetAttachmentResponse `json:"data"`
}
type GetAttachmentResponse struct {
	Attachment AttachmentResponse `json:"attachment"`
}

// @Summary Get attachment
// @Description Gets a single attachment's metadata by id, including filename, mime type, byte size, checksum, and uploader, without streaming the file content.
// @Tags Attachments
// @Accept json
// @Produce json
// @Param id path integer true "Attachment ID"
// @Success 200 {object} GetAttachmentResponseEnvelope "Attachment retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Attachment not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid attachment id"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /attachments/{id} [get]
func (h AttachmentHandler) Get(c fiber.Ctx) error {
	attachmentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid attachment id provided.", nil)
	}
	attachment, err := h.svc.Find(c, attachmentID)
	if err != nil {
		httpx.RequestLog(c).Error("attachment lookup failed", "attachment_id", attachmentID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get attachment.", err)
	}
	if attachment == nil || !httpx.OwnsTenant(c, &attachment.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Attachment not found.")
	}
	return httpx.CreateSuccessResponse(c, "Attachment retrieved successfully.", GetAttachmentResponse{
		Attachment: newAttachmentResponse(attachment),
	})
}

// @Summary Download attachment
// @Description Downloads an attachment by streaming its stored file content, setting the response filename and content type from the attachment's saved metadata.
// @Tags Attachments
// @Accept json
// @Produce application/octet-stream
// @Param id path integer true "Attachment ID"
// @Success 200 {file} binary "Attachment file"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Attachment not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid attachment id"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /attachments/{id}/download [get]
func (h AttachmentHandler) Download(c fiber.Ctx) error {
	attachmentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid attachment id provided.", nil)
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}
	attachment, reader, err := h.svc.Download(c, *organizationID, attachmentID)
	if err != nil {
		return writeAttachmentError(c, err)
	}
	defer func() { _ = reader.Close() }()
	c.Attachment(attachment.Filename)
	c.Set(fiber.HeaderContentType, attachment.MimeType)
	return c.SendStream(reader)
}

type UploadAttachmentRequest struct {
	OwnerType string `form:"owner_type" validate:"required"`
	OwnerID   uint64 `form:"owner_id" validate:"required,gt=0"`
}

type UploadAttachmentResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UploadAttachmentResponse `json:"data"`
}
type UploadAttachmentResponse struct {
	Attachment AttachmentResponse `json:"attachment"`
}

// @Summary Upload attachment
// @Description Uploads a file and associates it with the given owner type and owner id, recording the filename, mime type, byte size, and checksum. The authenticated user is recorded as the uploader, and empty files or missing owners are rejected.
// @Tags Attachments
// @Accept multipart/form-data
// @Produce json
// @Param owner_type formData string true "Owner type"
// @Param owner_id formData integer true "Owner ID"
// @Param file formData file true "File to upload"
// @Success 201 {object} UploadAttachmentResponseEnvelope "Attachment uploaded successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /attachments [post]
func (h AttachmentHandler) Upload(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A file field is required.", nil)
	}
	ownerType := c.FormValue("owner_type")
	ownerID, err := strconv.ParseUint(c.FormValue("owner_id"), 10, 64)
	if err != nil || ownerID == 0 {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A valid owner_id is required.", nil)
	}
	uploadedBy, ok := httpx.CallerID(c)
	if !ok {
		uploadedBy = 0
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}
	src, err := file.Open()
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Failed to read uploaded file.", nil)
	}
	defer func() { _ = src.Close() }()

	attachment, err := h.svc.Upload(c, *organizationID, ownerType, ownerID, uploadedBy, file.Filename, file.Header.Get("Content-Type"), src)
	if err != nil {
		return writeAttachmentError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Attachment uploaded successfully.", UploadAttachmentResponse{
		Attachment: newAttachmentResponse(attachment),
	})
}

// @Summary Delete attachment
// @Description Permanently deletes an attachment record together with its underlying stored file, returning no content on success.
// @Tags Attachments
// @Accept json
// @Produce json
// @Param id path integer true "Attachment ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Attachment not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid attachment id"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /attachments/{id} [delete]
func (h AttachmentHandler) Delete(c fiber.Ctx) error {
	attachmentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid attachment id provided.", nil)
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}
	if err := h.svc.Delete(c, *organizationID, attachmentID); err != nil {
		return writeAttachmentError(c, err)
	}
	return httpx.CreateNoContentResponse(c)
}

func writeAttachmentError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, crosscutting.ErrAttachmentOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	case errors.Is(err, crosscutting.ErrAttachmentNotFound):
		return httpx.CreateNotFoundResponse(c, "Attachment not found.")
	case errors.Is(err, crosscutting.ErrAttachmentOwner):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Attachment owner is required.", nil)
	case errors.Is(err, crosscutting.ErrAttachmentFilename):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Attachment filename is required.", nil)
	case errors.Is(err, crosscutting.ErrAttachmentEmpty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Attachment content cannot be empty.", nil)
	default:
		httpx.RequestLog(c).Error("attachment write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process attachment.", err)
	}
}
