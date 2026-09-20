package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/service"
)

func writeServiceError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrEquipmentNotFound),
		errors.Is(err, service.ErrContractNotFound),
		errors.Is(err, service.ErrOrderNotFound),
		errors.Is(err, service.ErrPlanNotFound):
		return httpx.CreateNotFoundResponse(c, "Resource not found.")
	case errors.Is(err, service.ErrEquipmentNameRequired),
		errors.Is(err, service.ErrContractNameRequired),
		errors.Is(err, service.ErrContractInvalidState),
		errors.Is(err, service.ErrOrderNameRequired),
		errors.Is(err, service.ErrOrderInvalidType),
		errors.Is(err, service.ErrOrderInvalidState),
		errors.Is(err, service.ErrOrderLineInvalidType),
		errors.Is(err, service.ErrOrderLineInvalidQty),
		errors.Is(err, service.ErrOrderLineInvalidPrice),
		errors.Is(err, service.ErrOrderNoLines),
		errors.Is(err, service.ErrOrderNotDone),
		errors.Is(err, service.ErrOrderNotCompleted),
		errors.Is(err, service.ErrOrderRequiresJournal),
		errors.Is(err, service.ErrOrderRequiresAccounts),
		errors.Is(err, service.ErrOrderRequiresContact),
		errors.Is(err, service.ErrOrderEmptyLines),
		errors.Is(err, service.ErrOrderContactForCOGS),
		errors.Is(err, service.ErrPlanNameRequired),
		errors.Is(err, service.ErrPlanIntervalInvalid),
		errors.Is(err, service.ErrPlanNextDueRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Request cannot be processed.", nil)
	default:
		httpx.RequestLog(c).Error("service write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process request.", err)
	}
}
