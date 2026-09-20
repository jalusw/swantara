package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type DimensionHandler struct {
	svc reference.DimensionService
}

func NewDimensionHandler(svc reference.DimensionService) DimensionHandler {
	return DimensionHandler{svc: svc}
}

type DimensionResponse struct {
	ID             uint64    `json:"id"`
	OrganizationID *uint64   `json:"organization_id"`
	Name           string    `json:"name"`
	Code           *string   `json:"code"`
	Kind           *string   `json:"kind"`
	ParentID       *uint64   `json:"parent_id"`
	Active         *bool     `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newDimensionResponse(account *reference.Dimension) DimensionResponse {
	return DimensionResponse{
		ID:             account.ID,
		OrganizationID: account.OrganizationID,
		Name:           account.Name,
		Code:           account.Code,
		Kind:           account.Kind,
		ParentID:       account.ParentID,
		Active:         account.Active,
		CreatedAt:      account.CreatedAt,
		UpdatedAt:      account.UpdatedAt,
	}
}

var dimensionQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"code":            {},
	"kind":            {},
	"parent_id":       {},
	"active":          {},
	"created_at":      {},
	"updated_at":      {},
}

type ListDimensionsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListDimensionsResponse `json:"data"`
}
type ListDimensionsResponse struct {
	Accounts []DimensionResponse `json:"accounts"`
}

// @Summary List dimension accounts
// @Description Lists dimension accounts with pagination, sorting, and filtering, automatically scoping results to the caller's organization; the response can be exported as JSON, XML, or CSV.
// @Tags Dimension Accounts
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. organization_id:eq:1)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListDimensionsResponseEnvelope "Dimension accounts retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/dimension-accounts [get]
func (h DimensionHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, dimensionQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("dimension account list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve dimension accounts.", err)
	}

	items := make([]DimensionResponse, len(page.Items))
	for i, account := range page.Items {
		items[i] = newDimensionResponse(account)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "dimension-accounts.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Dimension accounts retrieved successfully.", ListDimensionsResponse{
		Accounts: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetDimensionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetDimensionResponse `json:"data"`
}
type GetDimensionResponse struct {
	Account DimensionResponse `json:"account"`
}

// @Summary Get dimension account
// @Description Returns a single dimension account by id. The account must belong to the caller's organization, otherwise a 404 is returned.
// @Tags Dimension Accounts
// @Accept json
// @Produce json
// @Param id path integer true "Dimension account ID"
// @Success 200 {object} GetDimensionResponseEnvelope "Dimension account retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Dimension account not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/dimension-accounts/{id} [get]
func (h DimensionHandler) Get(c fiber.Ctx) error {
	accountID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid dimension account id provided.", nil)
	}

	account, err := h.svc.Find(c, accountID)
	if err != nil {
		httpx.RequestLog(c).Error("dimension account lookup failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get dimension account.", err)
	}
	if account == nil || !httpx.OwnsTenant(c, account.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Dimension account not found.")
	}

	return httpx.CreateSuccessResponse(c, "Dimension account retrieved successfully.", GetDimensionResponse{
		Account: newDimensionResponse(account),
	})
}

type CreateDimensionRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name" validate:"required"`
	Code           *string `json:"code"`
	Kind           *string `json:"kind"`
	ParentID       *uint64 `json:"parent_id"`
	Active         *bool   `json:"active"`
}

type CreateDimensionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateDimensionResponse `json:"data"`
}
type CreateDimensionResponse struct {
	Account DimensionResponse `json:"account"`
}

// @Summary Create dimension account
// @Description Creates an dimension account with an optional code, kind, parent, and active flag. A non-empty code must be unique within the organization and the parent, if given, must exist; the account is assigned to the caller's organization.
// @Tags Dimension Accounts
// @Accept json
// @Produce json
// @Param body body CreateDimensionRequest true "Dimension account details"
// @Success 201 {object} CreateDimensionResponseEnvelope "Dimension account created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error, duplicate code, or unknown parent"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/dimension-accounts [post]
func (h DimensionHandler) Create(c fiber.Ctx) error {
	var request CreateDimensionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	account, err := h.svc.Create(c, &reference.Dimension{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		Name:           request.Name,
		Code:           request.Code,
		Kind:           request.Kind,
		ParentID:       request.ParentID,
		Active:         request.Active,
	})
	if err != nil {
		return writeDimensionError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Dimension account created successfully.", CreateDimensionResponse{
		Account: newDimensionResponse(account),
	})
}

type UpdateDimensionRequest struct {
	Name     string  `json:"name" validate:"required"`
	Code     *string `json:"code"`
	Kind     *string `json:"kind"`
	ParentID *uint64 `json:"parent_id"`
	Active   *bool   `json:"active"`
}

type UpdateDimensionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateDimensionResponse `json:"data"`
}
type UpdateDimensionResponse struct {
	Account DimensionResponse `json:"account"`
}

// @Summary Update dimension account
// @Description Updates an dimension account's name, code, kind, parent, and active flag. A non-empty code must remain unique within the organization and the parent must exist; a 404 is returned if the account does not belong to the caller.
// @Tags Dimension Accounts
// @Accept json
// @Produce json
// @Param id path integer true "Dimension account ID"
// @Param body body UpdateDimensionRequest true "Dimension account details"
// @Success 200 {object} UpdateDimensionResponseEnvelope "Dimension account updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Dimension account not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error, duplicate code, or unknown parent"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/dimension-accounts/{id} [put]
func (h DimensionHandler) Update(c fiber.Ctx) error {
	accountID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid dimension account id provided.", nil)
	}

	var request UpdateDimensionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	account, err := h.svc.Find(c, accountID)
	if err != nil {
		httpx.RequestLog(c).Error("dimension account lookup failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update dimension account.", err)
	}
	if account == nil || !httpx.OwnsTenant(c, account.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Dimension account not found.")
	}

	account.Name = request.Name
	account.Code = request.Code
	account.Kind = request.Kind
	account.ParentID = request.ParentID
	account.Active = request.Active
	account.OrganizationID = httpx.TenantOrganizationID(c, account.OrganizationID)

	updated, err := h.svc.Update(c, account)
	if err != nil {
		return writeDimensionError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Dimension account updated successfully.", UpdateDimensionResponse{
		Account: newDimensionResponse(updated),
	})
}

// @Summary Delete dimension account
// @Description Deletes an dimension account by id. The account must belong to the caller's organization; a 404 is returned otherwise.
// @Tags Dimension Accounts
// @Accept json
// @Produce json
// @Param id path integer true "Dimension account ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Dimension account not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/dimension-accounts/{id} [delete]
func (h DimensionHandler) Delete(c fiber.Ctx) error {
	accountID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid dimension account id provided.", nil)
	}

	account, err := h.svc.Find(c, accountID)
	if err != nil {
		httpx.RequestLog(c).Error("dimension account lookup failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete dimension account.", err)
	}
	if account == nil || !httpx.OwnsTenant(c, account.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Dimension account not found.")
	}

	if err := h.svc.Delete(c, accountID); err != nil {
		httpx.RequestLog(c).Error("dimension account deletion failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete dimension account.", err)
	}

	return httpx.CreateNoContentResponse(c)
}

func writeDimensionError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reference.ErrDuplicateCode):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Dimension account code already exists for the organization.", nil)
	case errors.Is(err, reference.ErrParentNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Parent dimension account does not exist.", nil)
	default:
		httpx.RequestLog(c).Error("dimension account write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save dimension account.", err)
	}
}
