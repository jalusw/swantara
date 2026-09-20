package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
)

type InterorganizationHandler struct {
	svc interorganization.InterorganizationService
}

func NewInterorganizationHandler(
	svc interorganization.InterorganizationService,
) InterorganizationHandler {
	return InterorganizationHandler{svc: svc}
}
