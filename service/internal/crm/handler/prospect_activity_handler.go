package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ProspectActivityHandler struct {
	svc crm.ProspectActivityService
}

func NewProspectActivityHandler(svc crm.ProspectActivityService) ProspectActivityHandler {
	return ProspectActivityHandler{svc: svc}
}

var prospectActivityQueryAllowlist = map[string]struct{}{
	"prospect_id": {},
	"contact_id":  {},
	"type":        {},
	"summary":     {},
	"due_date":    {},
	"done":        {},
	"user_id":     {},
	"created_at":  {},
	"updated_at":  {},
}

type ProspectActivityResponse struct {
	ID         uint64     `json:"id"`
	ProspectID *uint64    `json:"prospect_id"`
	ContactID  *uint64    `json:"contact_id"`
	Type       string     `json:"type"`
	Summary    string     `json:"summary"`
	Note       *string    `json:"note"`
	DueDate    *time.Time `json:"due_date"`
	Done       bool       `json:"done"`
	DoneAt     *time.Time `json:"done_at"`
	UserID     *uint64    `json:"user_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func newProspectActivityResponse(activity *crm.ProspectActivity) ProspectActivityResponse {
	return ProspectActivityResponse{
		ID:         activity.ID,
		ProspectID: activity.ProspectID,
		ContactID:  activity.ContactID,
		Type:       activity.Type,
		Summary:    activity.Summary,
		Note:       activity.Note,
		DueDate:    activity.DueDate,
		Done:       activity.Done,
		DoneAt:     activity.DoneAt,
		UserID:     activity.UserID,
		CreatedAt:  activity.CreatedAt,
		UpdatedAt:  activity.UpdatedAt,
	}
}

func writeProspectActivityError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, crm.ErrActivityNotFound):
		return httpx.CreateNotFoundResponse(c, "Activity not found.")
	case errors.Is(err, crm.ErrActivitySummaryRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Activity summary is required.", nil)
	case errors.Is(err, crm.ErrActivityDone):
		return httpx.CreateConflictResponse(c, "Activity is already done.", err)
	case errors.Is(err, crm.ErrLeadNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "CRM prospect does not exist.", nil)
	case errors.Is(err, crm.ErrContactNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "CRM contact does not exist.", nil)
	default:
		httpx.RequestLog(c).Error("crm activity write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save activity.", err)
	}
}
