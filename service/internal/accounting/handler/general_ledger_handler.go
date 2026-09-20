package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GeneralLedgerHandler struct {
	svc accounting.GeneralLedgerService
}

func NewGeneralLedgerHandler(svc accounting.GeneralLedgerService) GeneralLedgerHandler {
	return GeneralLedgerHandler{svc: svc}
}

func (h GeneralLedgerHandler) List(c fiber.Ctx) error {
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	accountID, err := strconv.ParseUint(c.Query("account_id"), 10, 64)
	if err != nil || accountID == 0 {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Valid account_id is required.", nil)
	}
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "From and to dates are required.", nil)
	}
	from, err := helper.ParseDate(&fromStr)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "From must be in YYYY-MM-DD format.", nil)
	}
	to, err := helper.ParseDate(&toStr)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "To must be in YYYY-MM-DD format.", nil)
	}
	if to.Before(*from) {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date range is invalid.", nil)
	}
	lines, err := h.svc.ListByAccount(c, organizationID, accountID, *from, to.UTC().Add(23*time.Hour+59*time.Minute+59*time.Second))
	if err != nil {
		httpx.RequestLog(c).Error("general ledger lookup failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get general ledger.", err)
	}
	return httpx.CreateSuccessResponse(c, "General ledger retrieved successfully.", lines)
}
