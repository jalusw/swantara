package handler

import "github.com/jalusw/swantara/apps/service/internal/giftcard"

type CouponHandler struct {
	svc giftcard.CouponService
}

func NewCouponHandler(svc giftcard.CouponService) CouponHandler {
	return CouponHandler{svc: svc}
}
