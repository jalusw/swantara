package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func writeGiftCardError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, giftcard.ErrGiftCardNotFound):
		return httpx.CreateNotFoundResponse(c, "Gift card not found.")
	case errors.Is(err, giftcard.ErrGiftCardCodeRequired),
		errors.Is(err, giftcard.ErrGiftCardAmountInvalid),
		errors.Is(err, giftcard.ErrGiftCardInactive),
		errors.Is(err, giftcard.ErrGiftCardExpired),
		errors.Is(err, giftcard.ErrGiftCardInsufficient),
		errors.Is(err, giftcard.ErrGiftCardAccounts),
		errors.Is(err, giftcard.ErrGiftCardLiabilityOnly),
		errors.Is(err, giftcard.ErrGiftCardRevenue),
		errors.Is(err, giftcard.ErrGiftCardRefundAccount),
		errors.Is(err, giftcard.ErrGiftCardForfeitAccount),
		errors.Is(err, giftcard.ErrGiftCardAlreadyExists):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Request cannot be processed.", nil)
	default:
		httpx.RequestLog(c).Error("gift card write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process request.", err)
	}
}

func parseOptionalDate(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	return helper.ParseDateStr(raw)
}
