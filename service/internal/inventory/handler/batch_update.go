package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type UpdateBatchRequest struct {
	ItemID         uint64     `json:"item_id" validate:"required,gt=0"`
	Name           string     `json:"name" validate:"required"`
	Ref            *string    `json:"ref"`
	ExpiryDate     *time.Time `json:"expiry_date"`
	BestBeforeDate *time.Time `json:"best_before_date"`
}

type UpdateBatchResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateBatchResponse `json:"data"`
}
type UpdateBatchResponse struct {
	Lot BatchResponse `json:"batch"`
}

// @Summary Update stock batch
// @Description Updates an existing stock batch by its id, returning 404 Not Found if the batch does not exist. The item's batch/serial tracking requirement and the uniqueness of the batch name per item are re-validated before saving. Returns the updated batch.
// @Tags Stock Lots
// @Accept json
// @Produce json
// @Param id path integer true "Stock batch ID"
// @Param body body UpdateBatchRequest true "Stock batch details"
// @Success 200 {object} UpdateBatchResponseEnvelope "Stock batch updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock batch not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /stock-lots/{id} [put]
func (h BatchHandler) Update(c fiber.Ctx) error {
	batchID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid stock batch id provided.", nil)
	}

	var request UpdateBatchRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	batch, err := h.svc.FindInOrg(c, batchID, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("stock batch lookup failed", "batch_id", batchID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update stock batch.", err)
	}
	if batch == nil {
		return httpx.CreateNotFoundResponse(c, "Stock batch not found.")
	}

	batch.ItemID = request.ItemID
	batch.Name = request.Name
	batch.Ref = request.Ref
	batch.ExpiryDate = request.ExpiryDate
	batch.BestBeforeDate = request.BestBeforeDate

	updated, err := h.svc.Update(c, *organizationID, batch)
	if err != nil {
		return writeLotError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Stock batch updated successfully.", UpdateBatchResponse{
		Lot: newBatchResponse(updated),
	})
}
