package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type CreateProspectActivityRequest struct {
	ProspectID *uint64 `json:"lead_id"`
	ContactID  *uint64 `json:"contact_id"`
	Type       string  `json:"type"`
	Summary    string  `json:"summary" validate:"required"`
	Note       *string `json:"note"`
	DueDate    *string `json:"due_date"`
	Done       *bool   `json:"done"`
}

type CreateProspectActivityResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateProspectActivityResponse `json:"data"`
}

type CreateProspectActivityResponse struct {
	Activity ProspectActivityResponse `json:"activity"`
}

// @Summary Create activity
// @Description Logs a CRM activity (call, meeting, email, note, or task) optionally linked to a prospect or contact, requiring a summary and validating that any referenced prospect or contact exists. The due date is parsed as a timestamp, the acting user is recorded as the activity's assignee, and an omitted type defaults to note.
// @Tags CRM Activities
// @Accept json
// @Produce json
// @Param body body CreateProspectActivityRequest true "Activity details"
// @Success 201 {object} CreateProspectActivityResponseEnvelope "Activity created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown reference"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /crm/activities [post]
func (h ProspectActivityHandler) Create(c fiber.Ctx) error {
	var request CreateProspectActivityRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	dueDate, err := helper.ParseTimestamp(request.DueDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid due date provided.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	activity, err := h.svc.CreateActivity(c, *organizationID, &crm.ProspectActivity{
		ProspectID: request.ProspectID,
		ContactID:  request.ContactID,
		Type:       request.Type,
		Summary:    request.Summary,
		Note:       request.Note,
		DueDate:    dueDate,
		Done:       request.Done != nil && *request.Done,
	})
	if err != nil {
		return writeProspectActivityError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Activity created successfully.", CreateProspectActivityResponse{
		Activity: newProspectActivityResponse(activity),
	})
}
