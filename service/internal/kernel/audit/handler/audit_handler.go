package handler

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/audit"
)

type AuditLogHandler struct {
	logs audit.LogDAO
}

func NewAuditLogHandler(logs audit.LogDAO) AuditLogHandler {
	return AuditLogHandler{logs: logs}
}

type AuditLogResponse struct {
	ID        uint64          `json:"id"`
	TableName string          `json:"table_name"`
	RecordID  uint64          `json:"record_id"`
	Action    string          `json:"action"`
	ChangedBy uint64          `json:"changed_by"`
	ChangedAt time.Time       `json:"changed_at"`
	Diff      json.RawMessage `json:"diff" swaggertype:"object"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func newAuditLogResponse(log *audit.Log) AuditLogResponse {
	return AuditLogResponse{
		ID:        log.ID,
		TableName: log.EntityTable,
		RecordID:  log.RecordID,
		Action:    string(log.Action),
		ChangedBy: log.ChangedBy,
		ChangedAt: log.ChangedAt,
		Diff:      log.Diff,
		CreatedAt: log.CreatedAt,
		UpdatedAt: log.UpdatedAt,
	}
}

var auditLogQueryAllowlist = map[string]struct{}{
	"table_name": {},
	"record_id":  {},
	"action":     {},
	"changed_by": {},
	"changed_at": {},
	"created_at": {},
}

type ListAuditLogsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListAuditLogsResponse `json:"data"`
}
type ListAuditLogsResponse struct {
	AuditLogs []AuditLogResponse `json:"audit_logs"`
}

// @Summary List audit logs
// @Description Lists audit logs with pagination, sorting, and filtering.
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param organization_id path int true "Organization ID"
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param sort query string false "Sort fields (e.g. created_at:desc)"
// @Success 200 {object} ListAuditLogsResponseEnvelope "Audit logs retrieved successfully."
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/audit-logs [get]
func (h AuditLogHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, auditLogQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	page, err := h.logs.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("audit log list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve audit logs.", err)
	}

	items := make([]AuditLogResponse, len(page.Items))
	for i, log := range page.Items {
		items[i] = newAuditLogResponse(log)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Audit logs retrieved successfully.", ListAuditLogsResponse{
		AuditLogs: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetAuditLogResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetAuditLogResponse `json:"data"`
}
type GetAuditLogResponse struct {
	AuditLog AuditLogResponse `json:"audit_log"`
}

// @Summary Get audit log
// @Description Gets a single audit log entry by id.
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param organization_id path int true "Organization ID"
// @Param id path int true "Audit Log ID"
// @Success 200 {object} GetAuditLogResponseEnvelope "Audit log retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Audit log not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/audit-logs/{id} [get]
func (h AuditLogHandler) Get(c fiber.Ctx) error {
	logID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid audit log id provided.", nil)
	}

	log, err := h.logs.Find(c, logID)
	if err != nil {
		httpx.RequestLog(c).Error("audit log lookup failed", "log_id", logID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve audit log.", err)
	}
	if log == nil {
		return httpx.CreateNotFoundResponse(c, "Audit log not found.")
	}

	return httpx.CreateSuccessResponse(c, "Audit log retrieved successfully.", GetAuditLogResponse{
		AuditLog: newAuditLogResponse(log),
	})
}
