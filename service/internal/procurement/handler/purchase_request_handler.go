package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type PurchaseRequestHandler struct {
	svc procurement.PurchaseRequestService
}

func NewPurchaseRequestHandler(
	svc procurement.PurchaseRequestService,
) PurchaseRequestHandler {
	return PurchaseRequestHandler{svc: svc}
}
