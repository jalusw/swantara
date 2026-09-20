package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
)

func writeInterorganizationError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, interorganization.ErrDropShipSourceNotFound),
		errors.Is(err, interorganization.ErrDropShipOrderNotFound),
		errors.Is(err, interorganization.ErrMirrorSourceNotFound),
		errors.Is(err, interorganization.ErrMirrorRuleMissing),
		errors.Is(err, interorganization.ErrRuleNotFound),
		errors.Is(err, interorganization.ErrConsolidationRunNotFound):
		return httpx.CreateNotFoundResponse(c, "Resource not found.")
	case errors.Is(err, interorganization.ErrDropShipSourceNotActive),
		errors.Is(err, interorganization.ErrDropShipNoLines),
		errors.Is(err, interorganization.ErrDropShipAlreadyReceived),
		errors.Is(err, interorganization.ErrDropShipDestinationMiss),
		errors.Is(err, interorganization.ErrDropShipSupplierMissing),
		errors.Is(err, interorganization.ErrMirrorSourceNoLines),
		errors.Is(err, interorganization.ErrMirrorSourceNotPosted),
		errors.Is(err, interorganization.ErrMirrorRuleDisabled),
		errors.Is(err, interorganization.ErrRuleSameOrganization),
		errors.Is(err, interorganization.ErrRuleDuplicate),
		errors.Is(err, interorganization.ErrRuleRequired),
		errors.Is(err, interorganization.ErrRuleContactsRequired),
		errors.Is(err, interorganization.ErrConsolidationNotDraft),
		errors.Is(err, interorganization.ErrConsolidationNoOrg),
		errors.Is(err, interorganization.ErrConsolidationNoPeriod),
		errors.Is(err, interorganization.ErrConsolidationPeriodMiss),
		errors.Is(err, interorganization.ErrConsolidationNoMember),
		errors.Is(err, interorganization.ErrConsolidationNoFx),
		errors.Is(err, interorganization.ErrConsolidationNoAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Request cannot be processed.", nil)
	default:
		httpx.RequestLog(c).Error("interorganization write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process request.", err)
	}
}
