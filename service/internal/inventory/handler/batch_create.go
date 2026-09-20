package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type CreateBatchRequest struct {
	ItemID         uint64     `json:"item_id" validate:"required,gt=0"`
	Name           string     `json:"name" validate:"required"`
	Ref            *string    `json:"ref"`
	ExpiryDate     *time.Time `json:"expiry_date"`
	BestBeforeDate *time.Time `json:"best_before_date"`
}

type CreateBatchResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateBatchResponse `json:"data"`
}
type CreateBatchResponse struct {
	Lot BatchResponse `json:"batch"`
}

// @Summary Create stock batch
// @Description Creates a stock batch for a item, optionally carrying a reference, expiry date, and best-before date. The item must exist and use batch/serial tracking, and the batch name must be unique per item or a 409 Conflict is returned. Returns the created batch with a 201 status.
// @Tags Stock Lots
// @Accept json
// @Produce json
// @Param body body CreateBatchRequest true "Stock batch details"
// @Success 201 {object} CreateBatchResponseEnvelope "Stock batch created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /stock-lots [post]
func (h BatchHandler) Create(c fiber.Ctx) error {
	var request CreateBatchRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	batch, err := h.svc.Create(c, *organizationID, &inventory.Batch{
		ItemID:         request.ItemID,
		Name:           request.Name,
		Ref:            request.Ref,
		ExpiryDate:     request.ExpiryDate,
		BestBeforeDate: request.BestBeforeDate,
	})
	if err != nil {
		return writeLotError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Stock batch created successfully.", CreateBatchResponse{
		Lot: newBatchResponse(batch),
	})
}
