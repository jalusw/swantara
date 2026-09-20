package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/service"
)

type ServiceContractResponse struct {
	ID               uint64     `json:"id"`
	OrganizationID   *uint64    `json:"organization_id"`
	Name             string     `json:"name"`
	ContactID        *uint64    `json:"contact_id"`
	EquipmentID      *uint64    `json:"equipment_id"`
	SubscriptionID   *uint64    `json:"subscription_id"`
	Coverage         string     `json:"coverage"`
	SLAResponseHours *int       `json:"sla_response_hours"`
	DateStart        *time.Time `json:"date_start"`
	DateEnd          *time.Time `json:"date_end"`
	State            string     `json:"state"`
}

func newServiceContractResponse(contract *service.ServiceContract) ServiceContractResponse {
	return ServiceContractResponse{
		ID: contract.ID, OrganizationID: contract.OrganizationID, Name: contract.Name,
		ContactID: contract.ContactID, EquipmentID: contract.EquipmentID, SubscriptionID: contract.SubscriptionID,
		Coverage: contract.Coverage, SLAResponseHours: contract.SLAResponseHours,
		DateStart: contract.DateStart, DateEnd: contract.DateEnd, State: contract.State,
	}
}

var serviceContractQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"contact_id":      {},
	"equipment_id":    {},
	"state":           {},
	"created_at":      {},
	"updated_at":      {},
}

type ListServiceContractsResponse struct {
	ServiceContracts []ServiceContractResponse `json:"service_contracts"`
}

type ListServiceContractsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListServiceContractsResponse `json:"data"`
}

// @Summary List service contracts
// @Description Lists service contracts with pagination, sorting, and filtering. Results are scoped to the caller's organization.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Success 200 {object} ListServiceContractsResponseEnvelope "Service contracts retrieved successfully."
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-contracts [get]
func (h ServiceHandler) ListContracts(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, serviceContractQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.ListContracts(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("service contract list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service contracts.", err)
	}
	items := make([]ServiceContractResponse, len(page.Items))
	for i, contract := range page.Items {
		items[i] = newServiceContractResponse(contract)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Service contracts retrieved successfully.", ListServiceContractsResponse{
		ServiceContracts: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetServiceContractResponse struct {
	ServiceContract ServiceContractResponse `json:"service_contract"`
}

type GetServiceContractResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetServiceContractResponse `json:"data"`
}

// @Summary Get service contract
// @Description Gets a single service contract by id. The contract must belong to the caller's organization; a request for a contract owned by another tenant returns 404.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Service contract ID"
// @Success 200 {object} GetServiceContractResponseEnvelope "Service contract retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Service contract not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-contracts/{id} [get]
func (h ServiceHandler) GetContract(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	contract, err := h.svc.FindContract(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("service contract get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service contract.", err)
	}
	if contract == nil || !httpx.OwnsTenant(c, contract.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Service contract not found.")
	}
	return httpx.CreateSuccessResponse(c, "Service contract retrieved successfully.", GetServiceContractResponse{
		ServiceContract: newServiceContractResponse(contract),
	})
}

type CreateServiceContractRequest struct {
	OrganizationID   *uint64    `json:"organization_id"`
	Name             string     `json:"name" validate:"required"`
	ContactID        *uint64    `json:"contact_id"`
	EquipmentID      *uint64    `json:"equipment_id"`
	SubscriptionID   *uint64    `json:"subscription_id"`
	Coverage         string     `json:"coverage"`
	SLAResponseHours *int       `json:"sla_response_hours"`
	DateStart        *time.Time `json:"date_start"`
	DateEnd          *time.Time `json:"date_end"`
}

type CreateServiceContractResponse struct {
	ServiceContract ServiceContractResponse `json:"service_contract"`
}

type CreateServiceContractResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateServiceContractResponse `json:"data"`
}

// @Summary Create service contract
// @Description Creates a service contract. A contract with both start and end dates becomes active; otherwise it stays draft until activated.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param body body CreateServiceContractRequest true "Contract details"
// @Success 201 {object} CreateServiceContractResponseEnvelope "Service contract created successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-contracts [post]
func (h ServiceHandler) CreateContract(c fiber.Ctx) error {
	var request CreateServiceContractRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	created, err := h.svc.CreateContract(c, service.CreateContractRequest{
		OrganizationID:   *organizationID,
		Name:             request.Name,
		ContactID:        request.ContactID,
		EquipmentID:      request.EquipmentID,
		SubscriptionID:   request.SubscriptionID,
		Coverage:         request.Coverage,
		SLAResponseHours: request.SLAResponseHours,
		DateStart:        request.DateStart,
		DateEnd:          request.DateEnd,
	})
	if err != nil {
		return writeServiceError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Service contract created successfully.", CreateServiceContractResponse{
		ServiceContract: newServiceContractResponse(created),
	})
}

// @Summary Activate service contract
// @Description Activates a draft service contract once it has start and end dates.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Service contract ID"
// @Success 200 {object} GetServiceContractResponseEnvelope "Service contract activated successfully."
// @Failure 422 {object} httpx.ErrorResponse "Contract must be draft with dates"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-contracts/{id}/activate [post]
