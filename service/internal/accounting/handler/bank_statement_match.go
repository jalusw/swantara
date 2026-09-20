package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type MatchBankStatementResponseEnvelope struct {
	httpx.EnvelopeBase
	Data MatchBankStatementResponse `json:"data"`
}
type MatchBankStatementResponse struct {
	Unreconciled []BankStatementLineResponse `json:"unreconciled"`
}

// @Summary Match bank statement
// @Description Auto-matches unreconciled bank statement lines to posted payments for the same contact and amount, marking each matched line and its payment as reconciled. The response returns the statement lines that remain unreconciled after matching.
// @Tags Bank Statements
// @Accept json
// @Produce json
// @Param id path integer true "Bank statement ID"
// @Success 200 {object} MatchBankStatementResponseEnvelope "Bank statement matched successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Bank statement not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/bank-statements/{id}/match [post]
func (h BankStatementHandler) Match(c fiber.Ctx) error {
	statementID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid bank statement id provided.", nil)
	}

	unreconciled, err := h.svc.Match(c, statementID)
	if err != nil {
		return writeBankStatementError(c, err)
	}
	items := make([]BankStatementLineResponse, len(unreconciled))
	for i, line := range unreconciled {
		items[i] = newBankStatementLineResponse(line)
	}

	return httpx.CreateSuccessResponse(c, "Bank statement matched successfully.", MatchBankStatementResponse{
		Unreconciled: items,
	})
}
