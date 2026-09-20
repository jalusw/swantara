package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type CreateOutsideProcessingOrderRequest struct {
	ProductionOrderID uint64 `json:"production_order_id" validate:"required,gt=0"`
	SupplierID        uint64 `json:"supplier_id" validate:"required,gt=0"`
}

// @Summary Create outside processing order
// @Description Creates a draft outside processing order for a confirmed or later production order that uses a subcontract recipe, linking the production order to the contracted supplier. A production order can only have one outside processing order, and the recipe must be of subcontract type.
// @Tags Subcontract Orders
// @Accept json
// @Produce json
// @Param body body CreateOutsideProcessingOrderRequest true "Outside processing order details"
// @Success 201 {object} OutsideProcessingOrderEnvelope "Outside processing order created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 409 {object} httpx.ErrorResponse "Outside processing order already exists for this production order"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or production order is not subcontractable"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subcontract-orders [post]
func (h OutsideProcessingHandler) Create(c fiber.Ctx) error {
	var request CreateOutsideProcessingOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	order, err := h.svc.Create(c, request.ProductionOrderID, request.SupplierID)
	if err != nil {
		return writeSubcontractError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Outside processing order created successfully.", newOutsideProcessingOrderResponse(order))
}
