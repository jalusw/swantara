package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/service"
)

type EquipmentResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID *uint64    `json:"organization_id"`
	Name           string     `json:"name"`
	ItemID         *uint64    `json:"item_id"`
	SerialBatchID  *uint64    `json:"serial_batch_id"`
	OwnerContactID *uint64    `json:"owner_contact_id"`
	FixedAssetID   *uint64    `json:"fixed_asset_id"`
	Location       string     `json:"location"`
	InstallDate    *time.Time `json:"install_date"`
	WarrantyEnd    *time.Time `json:"warranty_end"`
	Category       string     `json:"category"`
}

func newEquipmentResponse(e *service.Equipment) EquipmentResponse {
	return EquipmentResponse{
		ID: e.ID, OrganizationID: e.OrganizationID, Name: e.Name, ItemID: e.ItemID,
		SerialBatchID: e.SerialBatchID, OwnerContactID: e.OwnerContactID, FixedAssetID: e.FixedAssetID,
		Location: e.Location, InstallDate: e.InstallDate, WarrantyEnd: e.WarrantyEnd, Category: e.Category,
	}
}

var equipmentQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"item_id":         {},
	"category":        {},
	"created_at":      {},
	"updated_at":      {},
}

type ListEquipmentsResponse struct {
	Equipments []EquipmentResponse `json:"equipments"`
}

type ListEquipmentsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListEquipmentsResponse `json:"data"`
}

// @Summary List equipments
// @Description Lists registered equipments with pagination, sorting, and filtering. Results are scoped to the caller's organization.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Success 200 {object} ListEquipmentsResponseEnvelope "Equipments retrieved successfully."
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/equipments [get]
func (h ServiceHandler) ListEquipments(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, equipmentQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.ListEquipments(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("equipment list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve equipments.", err)
	}
	items := make([]EquipmentResponse, len(page.Items))
	for i, equipment := range page.Items {
		items[i] = newEquipmentResponse(equipment)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Equipments retrieved successfully.", ListEquipmentsResponse{
		Equipments: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetEquipmentResponse struct {
	Equipment EquipmentResponse `json:"equipment"`
}

type GetEquipmentResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetEquipmentResponse `json:"data"`
}

// @Summary Get equipment
// @Description Gets a single equipment by id. The equipment must belong to the caller's organization; a request for equipment owned by another tenant returns 404.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Equipment ID"
// @Success 200 {object} GetEquipmentResponseEnvelope "Equipment retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Equipment not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/equipments/{id} [get]
func (h ServiceHandler) GetEquipment(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	equipment, err := h.svc.FindEquipment(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("equipment get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve equipment.", err)
	}
	if equipment == nil || !httpx.OwnsTenant(c, equipment.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Equipment not found.")
	}
	return httpx.CreateSuccessResponse(c, "Equipment retrieved successfully.", GetEquipmentResponse{
		Equipment: newEquipmentResponse(equipment),
	})
}

type CreateEquipmentRequest struct {
	OrganizationID *uint64    `json:"organization_id"`
	Name           string     `json:"name" validate:"required"`
	ItemID         *uint64    `json:"item_id"`
	SerialBatchID  *uint64    `json:"serial_batch_id"`
	OwnerContactID *uint64    `json:"owner_contact_id"`
	FixedAssetID   *uint64    `json:"fixed_asset_id"`
	Location       string     `json:"location"`
	InstallDate    *time.Time `json:"install_date"`
	WarrantyEnd    *time.Time `json:"warranty_end"`
	Category       string     `json:"category"`
}

type CreateEquipmentResponse struct {
	Equipment EquipmentResponse `json:"equipment"`
}

type CreateEquipmentResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateEquipmentResponse `json:"data"`
}

// @Summary Create equipment
// @Description Registers a piece of equipment with its item, owner, and warranty window.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param body body CreateEquipmentRequest true "Equipment details"
// @Success 201 {object} CreateEquipmentResponseEnvelope "Equipment created successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/equipments [post]
func (h ServiceHandler) CreateEquipment(c fiber.Ctx) error {
	var request CreateEquipmentRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	created, err := h.svc.CreateEquipment(c, &service.Equipment{
		OrganizationID: organizationID,
		Name:           request.Name,
		ItemID:         request.ItemID,
		SerialBatchID:  request.SerialBatchID,
		OwnerContactID: request.OwnerContactID,
		FixedAssetID:   request.FixedAssetID,
		Location:       request.Location,
		InstallDate:    request.InstallDate,
		WarrantyEnd:    request.WarrantyEnd,
		Category:       request.Category,
	})
	if err != nil {
		httpx.RequestLog(c).Error("equipment create failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create equipment.", err)
	}
	return httpx.CreateCreatedResponse(c, "Equipment created successfully.", CreateEquipmentResponse{
		Equipment: newEquipmentResponse(created),
	})
}
