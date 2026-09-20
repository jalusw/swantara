package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/pos"
)

type POSOrderHandler struct {
	svc pos.POSService
}

func NewPOSOrderHandler(svc pos.POSService) POSOrderHandler {
	return POSOrderHandler{svc: svc}
}

var posOrderQueryAllowlist = map[string]struct{}{
	"id":           {},
	"session_id":   {},
	"contact_id":   {},
	"state":        {},
	"order_time":   {},
	"amount_total": {},
	"created_at":   {},
	"updated_at":   {},
}
