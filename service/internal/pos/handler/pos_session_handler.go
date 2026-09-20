package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/pos"
)

type POSSessionHandler struct {
	svc pos.POSService
}

func NewPOSSessionHandler(svc pos.POSService) POSSessionHandler {
	return POSSessionHandler{svc: svc}
}

var posSessionQueryAllowlist = map[string]struct{}{
	"id":         {},
	"config_id":  {},
	"cashier_id": {},
	"state":      {},
	"opened_at":  {},
	"closed_at":  {},
	"created_at": {},
	"updated_at": {},
}
