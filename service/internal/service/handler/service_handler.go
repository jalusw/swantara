package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/service"
)

type ServiceHandler struct {
	svc         service.ServiceService
	maintenance service.MaintenanceService
}

func NewServiceHandler(
	svc service.ServiceService,
	maintenance service.MaintenanceService,
) ServiceHandler {
	return ServiceHandler{svc: svc, maintenance: maintenance}
}

func parseOptionalDate(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	return helper.ParseDateStr(raw)
}
