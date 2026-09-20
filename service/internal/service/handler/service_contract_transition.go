package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h ServiceHandler) ActivateContract(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	contract, err := h.svc.FindContract(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("service contract get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service contract.", err)
	}
	if contract == nil || !httpx.OwnsTenant(c, contract.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Service contract not found.")
	}
	updated, err := h.svc.ActivateContract(c, contract.ID)
	if err != nil {
		return writeServiceError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Service contract activated successfully.", GetServiceContractResponse{
		ServiceContract: newServiceContractResponse(updated),
	})
}

// @Summary Cancel service contract
// @Description Cancels an active service contract.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Service contract ID"
// @Success 200 {object} GetServiceContractResponseEnvelope "Service contract cancelled successfully."
// @Failure 422 {object} httpx.ErrorResponse "Contract must be active"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-contracts/{id}/cancel [post]
func (h ServiceHandler) CancelContract(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	contract, err := h.svc.FindContract(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("service contract get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service contract.", err)
	}
	if contract == nil || !httpx.OwnsTenant(c, contract.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Service contract not found.")
	}
	updated, err := h.svc.CancelContract(c, contract.ID)
	if err != nil {
		return writeServiceError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Service contract cancelled successfully.", GetServiceContractResponse{
		ServiceContract: newServiceContractResponse(updated),
	})
}
