package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
)

type ContractHandler struct {
	svc payroll.HRService
}

func NewContractHandler(svc payroll.HRService) ContractHandler {
	return ContractHandler{svc: svc}
}

type ContractResponse struct {
	ID           uint64  `json:"id"`
	EmployeeID   uint64  `json:"employee_id"`
	DateStart    string  `json:"date_start"`
	DateEnd      *string `json:"date_end"`
	Wage         float64 `json:"wage"`
	WageType     string  `json:"wage_type"`
	CurrencyCode string  `json:"currency_code"`
	State        string  `json:"state"`
}

func newContractResponse(contract *payroll.EmploymentContract) ContractResponse {
	response := ContractResponse{
		ID:           contract.ID,
		EmployeeID:   contract.EmployeeID,
		DateStart:    contract.DateStart.Format("2006-01-02"),
		DateEnd:      helper.FormatDatePtr(contract.DateEnd),
		Wage:         contract.Wage,
		WageType:     contract.WageType,
		CurrencyCode: contract.CurrencyCode,
		State:        contract.State,
	}
	return response
}

type CreateContractRequest struct {
	EmployeeID   uint64  `json:"employee_id" validate:"required,gt=0"`
	DateStart    *string `json:"date_start"`
	DateEnd      *string `json:"date_end"`
	Wage         float64 `json:"wage" validate:"required,gt=0"`
	WageType     string  `json:"wage_type"`
	CurrencyCode string  `json:"currency_code"`
}

type ListContractsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []ContractResponse `json:"data"`
}

type ContractResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ContractResponse `json:"data"`
}

var contractQueryAllowlist = map[string]struct{}{
	"employee_id": {},
	"state":       {},
}

// @Summary List contracts
// @Description Lists employment contracts with pagination, sorting, and filtering, scoped to the caller's organization. Queries are restricted to allowlisted fields (employee and state), and invalid query parameters return a 422 response.
// @Tags Contracts
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListContractsResponseEnvelope "Contracts retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/contracts [get]
func (h ContractHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, contractQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	page, err := h.svc.ListContractsInOrg(c, parsedQuery, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("contract list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve contracts.", err)
	}
	items := make([]ContractResponse, len(page.Items))
	for i, contract := range page.Items {
		items[i] = newContractResponse(contract)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Contracts retrieved successfully.", items, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get contract
// @Description Gets a single employment contract by its id. The contract must belong to an employee in the caller's organization, and a 404 not found is returned when it does not, even if a contract with that id exists in another tenant.
// @Tags Contracts
// @Accept json
// @Produce json
// @Param id path integer true "Contract ID"
// @Success 200 {object} ContractResponseEnvelope "Contract retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Contract not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/contracts/{id} [get]
func (h ContractHandler) Get(c fiber.Ctx) error {
	contractID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid contract id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	contract, err := h.svc.FindContractInOrg(c, contractID, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("contract lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get contract.", err)
	}
	if contract == nil {
		return httpx.CreateNotFoundResponse(c, "Contract not found.")
	}
	return httpx.CreateSuccessResponse(c, "Contract retrieved successfully.", newContractResponse(contract))
}

// @Summary Create contract
// @Description Creates a new employment contract in active state for an existing employee in the caller's organization. The wage must be positive, dates must be in YYYY-MM-DD format, and default currency (IDR), monthly wage type, and a contract start date of today are applied when not supplied.
// @Tags Contracts
// @Accept json
// @Produce json
// @Param request body CreateContractRequest true "Contract details"
// @Success 201 {object} ContractResponseEnvelope "Contract created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Resource not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/contracts [post]
func (h ContractHandler) Create(c fiber.Ctx) error {
	var request CreateContractRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	dateStart, err := helper.ParseDate(request.DateStart)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date start must be in YYYY-MM-DD format.", nil)
	}
	dateEnd, err := helper.ParseDate(request.DateEnd)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date end must be in YYYY-MM-DD format.", nil)
	}
	contract := &payroll.EmploymentContract{
		EmployeeID:   request.EmployeeID,
		DateEnd:      dateEnd,
		Wage:         request.Wage,
		WageType:     request.WageType,
		CurrencyCode: request.CurrencyCode,
	}
	if dateStart != nil {
		contract.DateStart = *dateStart
	}
	created, err := h.svc.CreateContract(c, organizationID, contract)
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Contract created successfully.", newContractResponse(created))
}

type UpdateContractRequest struct {
	DateStart    *string `json:"date_start"`
	DateEnd      *string `json:"date_end"`
	Wage         float64 `json:"wage" validate:"required,gt=0"`
	WageType     string  `json:"wage_type"`
	CurrencyCode string  `json:"currency_code"`
	State        string  `json:"state"`
}

// @Summary Update contract
// @Description Updates an existing employment contract's dates, wage, wage type, currency, and state. The contract must belong to an employee in the caller's organization (404 otherwise), dates must be in YYYY-MM-DD format, and the wage must remain positive.
// @Tags Contracts
// @Accept json
// @Produce json
// @Param id path integer true "Contract ID"
// @Param request body UpdateContractRequest true "Contract details"
// @Success 200 {object} ContractResponseEnvelope "Contract updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Contract not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/contracts/{id} [put]
func (h ContractHandler) Update(c fiber.Ctx) error {
	contractID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid contract id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	contract, err := h.svc.FindContractInOrg(c, contractID, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("contract lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get contract.", err)
	}
	if contract == nil {
		return httpx.CreateNotFoundResponse(c, "Contract not found.")
	}
	var request UpdateContractRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	dateStart, err := helper.ParseDate(request.DateStart)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date start must be in YYYY-MM-DD format.", nil)
	}
	dateEnd, err := helper.ParseDate(request.DateEnd)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date end must be in YYYY-MM-DD format.", nil)
	}
	if dateStart != nil {
		contract.DateStart = *dateStart
	}
	contract.DateEnd = dateEnd
	contract.Wage = request.Wage
	contract.WageType = request.WageType
	contract.CurrencyCode = request.CurrencyCode
	contract.State = request.State
	updated, err := h.svc.UpdateContract(c, contract)
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Contract updated successfully.", newContractResponse(updated))
}

// @Summary Terminate contract
// @Description Terminates an employment contract, moving it from active to closed state. Only active contracts can be terminated, the contract must belong to an employee in the caller's organization (404 otherwise), and attempting to terminate a contract in any other state returns a 422 response.
// @Tags Contracts
// @Accept json
// @Produce json
// @Param id path integer true "Contract ID"
// @Success 200 {object} ContractResponseEnvelope "Contract terminated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Contract not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/contracts/{id}/terminate [post]
func (h ContractHandler) Terminate(c fiber.Ctx) error {
	contractID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid contract id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	if contract, err := h.svc.FindContractInOrg(c, contractID, organizationID); err != nil {
		httpx.RequestLog(c).Error("contract lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get contract.", err)
	} else if contract == nil {
		return httpx.CreateNotFoundResponse(c, "Contract not found.")
	}
	contract, err := h.svc.TerminateContract(c, contractID)
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Contract terminated successfully.", newContractResponse(contract))
}
