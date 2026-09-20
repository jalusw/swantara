package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetBankStatementResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetBankStatementResponse `json:"data"`
}
type GetBankStatementResponse struct {
	BankStatement BankStatementResponse       `json:"bank_statement"`
	Lines         []BankStatementLineResponse `json:"lines"`
}

// @Summary Get bank statement
// @Description Gets a single bank statement by its id together with its statement lines, returning 404 when the statement does not exist or belongs to another tenant. Each line shows its amount, contact, and reconciliation status.
// @Tags Bank Statements
// @Accept json
// @Produce json
// @Param id path integer true "Bank statement ID"
// @Success 200 {object} GetBankStatementResponseEnvelope "Bank statement retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Bank statement not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/bank-statements/{id} [get]
func (h BankStatementHandler) Get(c fiber.Ctx) error {
	statementID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid bank statement id provided.", nil)
	}

	statement, err := h.svc.Find(c, statementID)
	if err != nil {
		httpx.RequestLog(c).Error("bank statement lookup failed", "statement_id", statementID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get bank statement.", err)
	}
	if statement == nil || !httpx.OwnsTenant(c, statement.JournalID) {
		return httpx.CreateNotFoundResponse(c, "Bank statement not found.")
	}

	lines, err := h.svc.ListLines(c, statement.ID)
	if err != nil {
		httpx.RequestLog(c).Error("bank statement lines lookup failed", "statement_id", statement.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get bank statement lines.", err)
	}
	lineItems := make([]BankStatementLineResponse, len(lines))
	for i, line := range lines {
		lineItems[i] = newBankStatementLineResponse(line)
	}

	return httpx.CreateSuccessResponse(c, "Bank statement retrieved successfully.", GetBankStatementResponse{
		BankStatement: newBankStatementResponse(statement),
		Lines:         lineItems,
	})
}
