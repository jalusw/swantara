package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type BestSupplierOfferRequest struct {
	ItemID uint64 `json:"item_id" validate:"required,gt=0"`
	Qty    string `json:"qty" validate:"required"`
	Date   string `json:"date"`
}

type BestSupplierOfferResponseEnvelope struct {
	httpx.EnvelopeBase
	Data BestSupplierOfferResponse `json:"data"`
}
type BestSupplierOfferResponse struct {
	SupplierProduct SupplierProductResponse `json:"supplier_product"`
}

// @Summary Resolve best supplier offer
// @Description Resolves the best supplier offer for a item variant within the caller's organization at a given quantity and date, selecting the highest-priority entry whose minimum quantity and validity window match. The quantity must be a valid decimal and the date must be in YYYY-MM-DD format (defaulting to today when omitted); a 404 is returned if the variant is not found.
// @Tags Supplier Products
// @Accept json
// @Produce json
// @Param body body BestSupplierOfferRequest true "Resolution details"
// @Success 200 {object} BestSupplierOfferResponseEnvelope "Best supplier offer resolved successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Item variant not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or no valid offer"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/supplier-products/best-offer [post]
func (h SupplierProductHandler) BestOffer(c fiber.Ctx) error {
	var request BestSupplierOfferRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	qty, err := amount.FromString(request.Qty)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Quantity must be a valid decimal.", nil)
	}

	date := time.Now()
	if request.Date != "" {
		parsed, err := helper.ParseDate(&request.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
		}
		date = *parsed
	}

	offer, err := h.svc.BestOffer(c, organizationID, request.ItemID, qty, date)
	if err != nil {
		return writeSupplierProductError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Best supplier offer resolved successfully.", BestSupplierOfferResponse{
		SupplierProduct: newSupplierProductResponse(offer),
	})
}
