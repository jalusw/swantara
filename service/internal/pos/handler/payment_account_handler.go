package handler

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/pos"
)

type PaymentAccountHandler struct {
	svc *pos.PaymentAccountService
}

func NewPaymentAccountHandler(svc *pos.PaymentAccountService) PaymentAccountHandler {
	return PaymentAccountHandler{svc: svc}
}

type POSPaymentAccountResponse struct {
	ID             uint64 `json:"id"`
	OrganizationID uint64 `json:"organization_id"`
	Method         string `json:"method"`
	AccountID      uint64 `json:"account_id"`
}

type ListPaymentAccountsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPaymentAccountsResponse `json:"data"`
}
type ListPaymentAccountsResponse struct {
	PaymentAccounts []POSPaymentAccountResponse `json:"payment_accounts"`
}

// @Summary List POS payment method accounts
// @Description Lists the payment-method-to-GL-account mappings used when posting POS sales and refunds, scoped to the caller's organization.
// @Tags POS Payment Accounts
// @Produce json
// @Success 200 {object} ListPaymentAccountsResponseEnvelope "Payment accounts retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/payment-accounts [get]
func (h PaymentAccountHandler) List(c fiber.Ctx) error {
	parsedQuery := &query.Query{}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("pos payment account list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve POS payment accounts.", err)
	}
	items := make([]POSPaymentAccountResponse, len(page.Items))
	for i, account := range page.Items {
		items[i] = POSPaymentAccountResponse{
			ID:             account.ID,
			OrganizationID: account.OrganizationID,
			Method:         account.Method,
			AccountID:      account.AccountID,
		}
	}
	return httpx.CreateSuccessResponseWithMeta(c, "POS payment accounts retrieved successfully.", ListPaymentAccountsResponse{
		PaymentAccounts: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type UpsertPaymentAccountRequest struct {
	AccountID uint64 `json:"account_id" validate:"required,gt=0"`
}

type UpsertPaymentAccountResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpsertPaymentAccountResponse `json:"data"`
}
type UpsertPaymentAccountResponse struct {
	PaymentAccount POSPaymentAccountResponse `json:"payment_account"`
}

// @Summary Set the GL account for a POS payment method
// @Description Creates or updates the mapping from a payment method (such as cash or card) to a GL account. Mapped methods receive their own posting line on POS sales and refunds; unmapped methods fall back to the journal's default account.
// @Tags POS Payment Accounts
// @Accept json
// @Produce json
// @Param method path string true "Payment method"
// @Param body body UpsertPaymentAccountRequest true "Payment account details"
// @Success 200 {object} UpsertPaymentAccountResponseEnvelope "Payment account updated successfully."
// @Success 201 {object} UpsertPaymentAccountResponseEnvelope "Payment account created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/payment-accounts/{method} [put]
func (h PaymentAccountHandler) Upsert(c fiber.Ctx) error {
	method := strings.TrimSpace(c.Params("method"))
	if method == "" {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A payment method is required.", nil)
	}
	var request UpsertPaymentAccountRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	account, created, err := h.svc.Upsert(c, *organizationID, method, request.AccountID)
	if err != nil {
		switch {
		case errors.Is(err, pos.ErrPaymentAccountAlreadyExists):
			return httpx.CreateConflictResponse(c, "A mapping for this payment method already exists.", nil)
		default:
			httpx.RequestLog(c).Error("pos payment account upsert failed", "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to save POS payment account.", err)
		}
	}

	resp := UpsertPaymentAccountResponse{
		PaymentAccount: POSPaymentAccountResponse{
			ID:             account.ID,
			OrganizationID: account.OrganizationID,
			Method:         account.Method,
			AccountID:      account.AccountID,
		},
	}
	if created {
		return httpx.CreateCreatedResponse(c, "POS payment account created successfully.", resp)
	}
	return httpx.CreateSuccessResponse(c, "POS payment account updated successfully.", resp)
}

// @Summary Delete a POS payment method account mapping
// @Description Removes the GL account mapping for a payment method; future POS postings for that method fall back to the journal's default account.
// @Tags POS Payment Accounts
// @Produce json
// @Param method path string true "Payment method"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Mapping not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/payment-accounts/{method} [delete]
func (h PaymentAccountHandler) Delete(c fiber.Ctx) error {
	method := strings.TrimSpace(c.Params("method"))
	if method == "" {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A payment method is required.", nil)
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	if err := h.svc.Delete(c, *organizationID, method); err != nil {
		switch {
		case errors.Is(err, pos.ErrPaymentAccountNotFound):
			return httpx.CreateNotFoundResponse(c, "POS payment account not found.")
		default:
			httpx.RequestLog(c).Error("pos payment account delete failed", "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to delete POS payment account.", err)
		}
	}
	return httpx.CreateNoContentResponse(c)
}
