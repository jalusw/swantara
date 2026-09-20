package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetBatchResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetBatchResponse `json:"data"`
}
type GetBatchResponse struct {
	Lot BatchResponse `json:"batch"`
}

// @Summary Get stock batch
// @Description Gets a single stock batch by its id, including its item, reference, expiry date, and best-before date. A missing batch is answered with 404 Not Found. A non-numeric or malformed id is rejected with 422 Unprocessable Entity.
// @Tags Stock Lots
// @Accept json
// @Produce json
// @Param id path integer true "Stock batch ID"
// @Success 200 {object} GetBatchResponseEnvelope "Stock batch retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock batch not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /stock-lots/{id} [get]
func (h BatchHandler) Get(c fiber.Ctx) error {
	batchID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid stock batch id provided.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	batch, err := h.svc.FindInOrg(c, batchID, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("stock batch lookup failed", "batch_id", batchID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get stock batch.", err)
	}
	if batch == nil {
		return httpx.CreateNotFoundResponse(c, "Stock batch not found.")
	}

	return httpx.CreateSuccessResponse(c, "Stock batch retrieved successfully.", GetBatchResponse{
		Lot: newBatchResponse(batch),
	})
}
