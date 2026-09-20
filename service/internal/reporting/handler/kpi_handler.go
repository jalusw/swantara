package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

type KpiHandler struct {
	svc reporting.KpiService
}

func NewKpiHandler(svc reporting.KpiService) KpiHandler {
	return KpiHandler{svc: svc}
}

func parseDateRange(c fiber.Ctx) (time.Time, time.Time, bool) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := now
	if raw := c.Query("start"); raw != "" {
		parsed, err := helper.ParseDate(&raw)
		if err != nil {
			return time.Time{}, time.Time{}, false
		}
		start = *parsed
	}
	if raw := c.Query("end"); raw != "" {
		parsed, err := helper.ParseDate(&raw)
		if err != nil {
			return time.Time{}, time.Time{}, false
		}
		end = *parsed
	}
	return start, end, true
}

func organizationOf(c fiber.Ctx) (*uint64, error) {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return nil, httpx.ErrMissingOrganizationInContext
	}
	return organizationID, nil
}

func writeKpiError(c fiber.Ctx, err error) error {
	httpx.RequestLog(c).Error("kpi query failed", "error", err)
	return httpx.CreateInternalServerErrorResponse(c, "Failed to compute KPI.", err)
}
