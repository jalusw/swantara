package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
)

type GiftCardHandler struct {
	svc giftcard.GiftCardService
}

func NewGiftCardHandler(svc giftcard.GiftCardService) GiftCardHandler {
	return GiftCardHandler{svc: svc}
}
