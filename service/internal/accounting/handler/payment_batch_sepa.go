package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GenerateSEPAResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GenerateSEPAResponse `json:"data"`
}
type GenerateSEPAResponse struct {
	XML string `json:"xml"`
}

// @Summary Generate SEPA file
// @Description Generates a SEPA XML payment file for the payment batch
// @Tags Accounting Payment Batches
// @Accept json
// @Produce json
// @Param id path integer true "Payment batch ID"
// @Success 200 {object} GenerateSEPAResponseEnvelope "SEPA file generated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Payment batch not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid payment batch state"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /payment-batches/{id}/generate-sepa [post]
func (h PaymentBatchHandler) GenerateSEPA(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid batch id provided.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	xml, err := h.svc.GenerateSEPA(c, id, *organizationID)
	if err != nil {
		return writePaymentBatchError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "SEPA file generated successfully.", GenerateSEPAResponse{
		XML: xml,
	})
}
