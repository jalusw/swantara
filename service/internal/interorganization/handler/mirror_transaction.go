package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
)

type InterorganizationTransactionResponse struct {
	ID                   uint64  `json:"id"`
	SourceOrganizationID *uint64 `json:"source_organization_id"`
	SourceType           string  `json:"source_type"`
	SourceID             *uint64 `json:"source_id"`
	MirrorOrganizationID *uint64 `json:"mirror_organization_id"`
	MirrorType           string  `json:"mirror_type"`
	MirrorID             *uint64 `json:"mirror_id"`
	Amount               float64 `json:"amount"`
	State                string  `json:"state"`
}

func newInterorganizationTransactionResponse(t *interorganization.InterorganizationTransaction) InterorganizationTransactionResponse {
	return InterorganizationTransactionResponse{
		ID: t.ID, SourceOrganizationID: t.SourceOrganizationID, SourceType: t.SourceType, SourceID: t.SourceID,
		MirrorOrganizationID: t.MirrorOrganizationID, MirrorType: t.MirrorType, MirrorID: t.MirrorID,
		Amount: t.Amount, State: t.State,
	}
}

type CreateInterorganizationTransactionResponse struct {
	PurchaseOrder                DropshipOrderResponse                `json:"purchase_order"`
	InterorganizationTransaction InterorganizationTransactionResponse `json:"interorganization_transaction"`
}

type CreateInterorganizationTransactionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateInterorganizationTransactionResponse `json:"data"`
}

// @Summary Mirror sale order to purchase order
// @Description Mirrors a posted sale order of the caller's organization into a purchase order in the target organization per the interorganization rule, recording an interorganization transaction.
// @Tags Inter-organization
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Param body body MirrorSaleOrderRequest true "Mirror details"
// @Success 201 {object} CreateInterorganizationTransactionResponseEnvelope "Sale order mirrored successfully."
// @Failure 404 {object} httpx.ErrorResponse "Source order or rule not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/interorganization-transactions/mirror-sale-order/{id} [post]
func (h InterorganizationHandler) MirrorSaleOrder(c fiber.Ctx) error {
	var request MirrorSaleOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	po, transaction, err := h.svc.MirrorSaleOrder(c, interorganization.MirrorSaleOrderRequest{
		SaleOrderID:      id,
		ToOrganizationID: request.ToOrganizationID,
	})
	if err != nil {
		return writeInterorganizationError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Sale order mirrored successfully.", CreateInterorganizationTransactionResponse{
		PurchaseOrder:                newDropshipOrderResponse(po),
		InterorganizationTransaction: newInterorganizationTransactionResponse(transaction),
	})
}

type MirrorSaleOrderRequest struct {
	ToOrganizationID uint64 `json:"to_organization_id" validate:"required,gt=0"`
}

type MirrorInvoiceRequest struct {
	ToOrganizationID uint64 `json:"to_organization_id" validate:"required,gt=0"`
	JournalID        uint64 `json:"journal_id" validate:"required,gt=0"`
	ExpenseAccountID uint64 `json:"expense_account_id" validate:"required,gt=0"`
}

type MirrorInvoiceResponse struct {
	SupplierBillID               uint64                               `json:"supplier_bill_id"`
	SupplierBillTotal            float64                              `json:"supplier_bill_total"`
	InterorganizationTransaction InterorganizationTransactionResponse `json:"interorganization_transaction"`
}

type MirrorInvoiceResponseEnvelope struct {
	httpx.EnvelopeBase
	Data MirrorInvoiceResponse `json:"data"`
}

// @Summary Mirror customer invoice to supplier bill
// @Description Mirrors a posted customer invoice of the caller's organization into a supplier bill in the target organization per the interorganization rule, recording an interorganization transaction. Lines post to the given expense account without carried taxes.
// @Tags Inter-organization
// @Accept json
// @Produce json
// @Param id path integer true "Customer invoice ID"
// @Param body body MirrorInvoiceRequest true "Mirror details"
// @Success 201 {object} MirrorInvoiceResponseEnvelope "Invoice mirrored successfully."
// @Failure 404 {object} httpx.ErrorResponse "Source invoice or rule not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/interorganization-transactions/mirror-invoice/{id} [post]
func (h InterorganizationHandler) MirrorInvoice(c fiber.Ctx) error {
	var request MirrorInvoiceRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	bill, transaction, err := h.svc.MirrorInvoice(c, interorganization.MirrorInvoiceRequest{
		InvoiceID:        id,
		ToOrganizationID: request.ToOrganizationID,
		JournalID:        request.JournalID,
		ExpenseAccountID: request.ExpenseAccountID,
	})
	if err != nil {
		return writeInterorganizationError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Invoice mirrored successfully.", MirrorInvoiceResponse{
		SupplierBillID:               bill.ID,
		SupplierBillTotal:            bill.AmountTotal.Float64(),
		InterorganizationTransaction: newInterorganizationTransactionResponse(transaction),
	})
}

type ListInterorganizationTransactionsResponse struct {
	InterorganizationTransactions []InterorganizationTransactionResponse `json:"interorganization_transactions"`
}

type ListInterorganizationTransactionsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListInterorganizationTransactionsResponse `json:"data"`
}

// @Summary List interorganization transactions
// @Description Lists the outgoing interorganization transactions of the caller's organization.
// @Tags Inter-organization
// @Accept json
// @Produce json
// @Success 200 {object} ListInterorganizationTransactionsResponseEnvelope "Interorganization transactions retrieved successfully."
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/interorganization-transactions [get]
func (h InterorganizationHandler) ListTransactions(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	transactions, err := h.svc.ListTransactions(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("interorganization transactions list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve interorganization transactions.", err)
	}
	items := make([]InterorganizationTransactionResponse, len(transactions))
	for i, transaction := range transactions {
		items[i] = newInterorganizationTransactionResponse(transaction)
	}
	return httpx.CreateSuccessResponse(c, "Interorganization transactions retrieved successfully.", ListInterorganizationTransactionsResponse{
		InterorganizationTransactions: items,
	})
}
