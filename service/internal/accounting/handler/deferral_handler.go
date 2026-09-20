package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/accounting"
)

type DeferralHandler struct {
	svc accounting.DeferralService
}

func NewDeferralHandler(svc accounting.DeferralService) DeferralHandler {
	return DeferralHandler{svc: svc}
}
