package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/asset"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type AssetCategoryHandler struct {
	categories asset.AssetCategoryService
	svc        asset.AssetService
}

func NewAssetCategoryHandler(
	categories asset.AssetCategoryService,
	svc asset.AssetService,
) AssetCategoryHandler {
	return AssetCategoryHandler{categories: categories, svc: svc}
}

type AssetCategoryResponse struct {
	ID                    uint64  `json:"id"`
	OrganizationID        *uint64 `json:"organization_id"`
	Name                  string  `json:"name"`
	AssetAccountID        *uint64 `json:"asset_account_id"`
	DepreciationAccountID *uint64 `json:"depreciation_account_id"`
	ExpenseAccountID      *uint64 `json:"expense_account_id"`
	GainAccountID         *uint64 `json:"gain_account_id"`
	LossAccountID         *uint64 `json:"loss_account_id"`
	Method                *string `json:"method"`
	MethodNumber          *int    `json:"method_number"`
	MethodPeriod          *string `json:"method_period"`
}

func newAssetCategoryResponse(c *reference.AssetCategory) AssetCategoryResponse {
	return AssetCategoryResponse{
		ID:                    c.ID,
		OrganizationID:        c.OrganizationID,
		Name:                  c.Name,
		AssetAccountID:        c.AssetAccountID,
		DepreciationAccountID: c.DepreciationAccountID,
		ExpenseAccountID:      c.ExpenseAccountID,
		GainAccountID:         c.GainAccountID,
		LossAccountID:         c.LossAccountID,
		Method:                c.Method,
		MethodNumber:          c.MethodNumber,
		MethodPeriod:          c.MethodPeriod,
	}
}

var assetCategoryQueryAllowlist = map[string]struct{}{
	"organization_id":         {},
	"name":                    {},
	"asset_account_id":        {},
	"depreciation_account_id": {},
	"expense_account_id":      {},
	"method":                  {},
	"created_at":              {},
	"updated_at":              {},
}

type ListAssetCategoriesResponse struct {
	AssetCategories []AssetCategoryResponse `json:"asset_categories"`
}

type ListAssetCategoriesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListAssetCategoriesResponse `json:"data"`
}

// @Summary List asset categories
// @Description Lists asset categories with pagination, sorting, and filtering. Results are scoped to the caller's organization.
// @Tags Asset Categories
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListAssetCategoriesResponseEnvelope "Asset categories retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/asset-categories [get]
func (h AssetCategoryHandler) ListCategories(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, assetCategoryQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.categories.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("asset category list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve asset categories.", err)
	}
	items := make([]AssetCategoryResponse, len(page.Items))
	for i, category := range page.Items {
		items[i] = newAssetCategoryResponse(category)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Asset categories retrieved successfully.", ListAssetCategoriesResponse{
		AssetCategories: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetAssetCategoryResponse struct {
	AssetCategory AssetCategoryResponse `json:"asset_category"`
}

type GetAssetCategoryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetAssetCategoryResponse `json:"data"`
}

// @Summary Get asset category
// @Description Gets a single asset category by id, including its depreciation method, period, and GL accounts. The category must belong to the caller's organization; a request for a category owned by another tenant returns 404.
// @Tags Asset Categories
// @Accept json
// @Produce json
// @Param id path integer true "Asset category ID"
// @Success 200 {object} GetAssetCategoryResponseEnvelope "Asset category retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Asset category not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/asset-categories/{id} [get]
func (h AssetCategoryHandler) GetCategory(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	category, err := h.categories.Find(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("asset category get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve asset category.", err)
	}
	if category == nil || !httpx.OwnsTenant(c, category.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Asset category not found.")
	}
	return httpx.CreateSuccessResponse(c, "Asset category retrieved successfully.", GetAssetCategoryResponse{
		AssetCategory: newAssetCategoryResponse(category),
	})
}

type CreateAssetCategoryRequest struct {
	OrganizationID        *uint64 `json:"organization_id"`
	Name                  string  `json:"name" validate:"required"`
	AssetAccountID        *uint64 `json:"asset_account_id"`
	DepreciationAccountID *uint64 `json:"depreciation_account_id"`
	ExpenseAccountID      *uint64 `json:"expense_account_id"`
	GainAccountID         *uint64 `json:"gain_account_id"`
	LossAccountID         *uint64 `json:"loss_account_id"`
	Method                string  `json:"method" validate:"required,oneof=linear declining declining_then_linear"`
	MethodNumber          *int    `json:"method_number" validate:"required,gt=0"`
	MethodPeriod          string  `json:"method_period" validate:"required,oneof=month year"`
}

type CreateAssetCategoryResponse struct {
	AssetCategory AssetCategoryResponse `json:"asset_category"`
}

type CreateAssetCategoryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateAssetCategoryResponse `json:"data"`
}

// @Summary Create asset category
// @Description Creates an asset category with its depreciation method and GL accounts.
// @Tags Asset Categories
// @Accept json
// @Produce json
// @Param body body CreateAssetCategoryRequest true "Asset category details"
// @Success 201 {object} CreateAssetCategoryResponseEnvelope "Asset category created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/asset-categories [post]
func (h AssetCategoryHandler) CreateCategory(c fiber.Ctx) error {
	var request CreateAssetCategoryRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	created, err := h.categories.Create(c, &reference.AssetCategory{
		OrganizationID:        organizationID,
		Name:                  request.Name,
		AssetAccountID:        request.AssetAccountID,
		DepreciationAccountID: request.DepreciationAccountID,
		ExpenseAccountID:      request.ExpenseAccountID,
		GainAccountID:         request.GainAccountID,
		LossAccountID:         request.LossAccountID,
		Method:                &request.Method,
		MethodNumber:          request.MethodNumber,
		MethodPeriod:          &request.MethodPeriod,
	})
	if err != nil {
		httpx.RequestLog(c).Error("asset category create failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create asset category.", err)
	}
	return httpx.CreateCreatedResponse(c, "Asset category created successfully.", CreateAssetCategoryResponse{
		AssetCategory: newAssetCategoryResponse(created),
	})
}

type FixedAssetResponse struct {
	ID              uint64     `json:"id"`
	OrganizationID  uint64     `json:"organization_id"`
	Name            string     `json:"name"`
	CategoryID      uint64     `json:"category_id"`
	PurchaseValue   float64    `json:"purchase_value"`
	SalvageValue    float64    `json:"salvage_value"`
	AcquisitionDate *time.Time `json:"acquisition_date"`
	InServiceDate   *time.Time `json:"in_service_date"`
	OriginalEntryID *uint64    `json:"original_entry_id"`
	InvoiceLineID   *uint64    `json:"invoice_line_id"`
	State           string     `json:"state"`
	DisposalDate    *time.Time `json:"disposal_date"`
}

func newFixedAssetResponse(a *asset.FixedAsset) FixedAssetResponse {
	return FixedAssetResponse{
		ID:              a.ID,
		OrganizationID:  a.OrganizationID,
		Name:            a.Name,
		CategoryID:      a.CategoryID,
		PurchaseValue:   a.PurchaseValue,
		SalvageValue:    a.SalvageValue,
		AcquisitionDate: a.AcquisitionDate,
		InServiceDate:   a.InServiceDate,
		OriginalEntryID: a.OriginalEntryID,
		InvoiceLineID:   a.InvoiceLineID,
		State:           a.State,
		DisposalDate:    a.DisposalDate,
	}
}

var fixedAssetQueryAllowlist = map[string]struct{}{
	"organization_id":   {},
	"category_id":       {},
	"state":             {},
	"invoice_line_id":   {},
	"original_entry_id": {},
	"created_at":        {},
	"updated_at":        {},
}

type ListFixedAssetsResponse struct {
	FixedAssets []FixedAssetResponse `json:"fixed_assets"`
}

type ListFixedAssetsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListFixedAssetsResponse `json:"data"`
}

// @Summary List fixed assets
// @Description Lists fixed assets with pagination, sorting, and filtering. Results are scoped to the caller's organization.
// @Tags Fixed Assets
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListFixedAssetsResponseEnvelope "Fixed assets retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/fixed-assets [get]
func (h AssetCategoryHandler) ListAssets(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, fixedAssetQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("fixed asset list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve fixed assets.", err)
	}
	items := make([]FixedAssetResponse, len(page.Items))
	for i, asset := range page.Items {
		items[i] = newFixedAssetResponse(asset)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Fixed assets retrieved successfully.", ListFixedAssetsResponse{
		FixedAssets: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetFixedAssetResponse struct {
	FixedAsset FixedAssetResponse `json:"fixed_asset"`
}

type GetFixedAssetResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetFixedAssetResponse `json:"data"`
}

// @Summary Get fixed asset
// @Description Gets a single fixed asset by id, including its category, purchase and salvage values, acquisition and in-service dates, current state, and disposal date. The asset must belong to the caller's organization; a request for an asset owned by another tenant returns 404.
// @Tags Fixed Assets
// @Accept json
// @Produce json
// @Param id path integer true "Fixed asset ID"
// @Success 200 {object} GetFixedAssetResponseEnvelope "Fixed asset retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Fixed asset not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/fixed-assets/{id} [get]
func (h AssetCategoryHandler) GetAsset(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	assetRecord, err := h.svc.Find(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("fixed asset get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve fixed asset.", err)
	}
	if assetRecord == nil || !httpx.OwnsTenant(c, helper.Ptr(assetRecord.OrganizationID)) {
		return httpx.CreateNotFoundResponse(c, "Fixed asset not found.")
	}
	return httpx.CreateSuccessResponse(c, "Fixed asset retrieved successfully.", GetFixedAssetResponse{
		FixedAsset: newFixedAssetResponse(assetRecord),
	})
}

type RegisterAssetRequest struct {
	OrganizationID  *uint64 `json:"organization_id"`
	Name            string  `json:"name" validate:"required"`
	CategoryID      uint64  `json:"category_id" validate:"required,gt=0"`
	PurchaseValue   float64 `json:"purchase_value" validate:"required,gt=0"`
	SalvageValue    float64 `json:"salvage_value"`
	AcquisitionDate string  `json:"acquisition_date" validate:"required"`
	InServiceDate   string  `json:"in_service_date" validate:"required"`
	InvoiceLineID   uint64  `json:"invoice_line_id" validate:"required,gt=0"`
}

// @Summary Register fixed asset
// @Description Registers a fixed asset from a posted supplier bill line, using the category's depreciation method and GL accounts.
// @Tags Fixed Assets
// @Accept json
// @Produce json
// @Param body body RegisterAssetRequest true "Asset registration details"
// @Success 201 {object} CreateAssetCategoryResponseEnvelope "Fixed asset registered successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Supplier bill line not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/fixed-assets [post]
func (h AssetCategoryHandler) RegisterAsset(c fiber.Ctx) error {
	var request RegisterAssetRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	acquisition, err := helper.ParseDateStr(request.AcquisitionDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid acquisition date.", nil)
	}
	inService, err := helper.ParseDateStr(request.InServiceDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid in-service date.", nil)
	}
	created, err := h.svc.Register(c, asset.RegisterAssetRequest{
		OrganizationID:  *organizationID,
		Name:            request.Name,
		CategoryID:      request.CategoryID,
		PurchaseValue:   request.PurchaseValue,
		SalvageValue:    request.SalvageValue,
		AcquisitionDate: acquisition,
		InServiceDate:   inService,
		InvoiceLineID:   request.InvoiceLineID,
	})
	if err != nil {
		return writeAssetError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Fixed asset registered successfully.", GetFixedAssetResponse{
		FixedAsset: newFixedAssetResponse(created),
	})
}

type AssetDepreciationLineResponse struct {
	ID               uint64    `json:"id"`
	AssetID          uint64    `json:"asset_id"`
	Sequence         int       `json:"sequence"`
	DepreciationDate time.Time `json:"depreciation_date"`
	Amount           float64   `json:"amount"`
	Accumulated      float64   `json:"accumulated"`
	RemainingValue   float64   `json:"remaining_value"`
	EntryID          *uint64   `json:"entry_id"`
	Posted           bool      `json:"posted"`
}

func newDepreciationLineResponse(line *asset.AssetDepreciationLine) AssetDepreciationLineResponse {
	return AssetDepreciationLineResponse{
		ID:               line.ID,
		AssetID:          line.AssetID,
		Sequence:         line.Sequence,
		DepreciationDate: line.DepreciationDate,
		Amount:           line.Amount,
		Accumulated:      line.Accumulated,
		RemainingValue:   line.RemainingValue,
		EntryID:          line.EntryID,
		Posted:           line.Posted,
	}
}

type ScheduleResponse struct {
	Lines []AssetDepreciationLineResponse `json:"lines"`
}

// @Summary Generate depreciation schedule
// @Description Generates the asset depreciation schedule from the category's method, creating one line per period.
// @Tags Fixed Assets
// @Accept json
// @Produce json
// @Param id path integer true "Fixed asset ID"
// @Success 201 {object} ScheduleResponse "Depreciation schedule generated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Fixed asset not found"
// @Failure 422 {object} httpx.ErrorResponse "Asset not running or schedule exists"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/fixed-assets/{id}/schedule [post]
func (h AssetCategoryHandler) GenerateSchedule(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	assetRecord, err := h.svc.Find(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("fixed asset get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve fixed asset.", err)
	}
	if assetRecord == nil || !httpx.OwnsTenant(c, helper.Ptr(assetRecord.OrganizationID)) {
		return httpx.CreateNotFoundResponse(c, "Fixed asset not found.")
	}
	schedule, err := h.svc.GenerateSchedule(c, asset.GenerateScheduleRequest{AssetID: assetRecord.ID})
	if err != nil {
		return writeAssetError(c, err)
	}
	items := make([]AssetDepreciationLineResponse, len(schedule))
	for i, line := range schedule {
		items[i] = newDepreciationLineResponse(&line)
	}
	return httpx.CreateCreatedResponse(c, "Depreciation schedule generated successfully.", ScheduleResponse{
		Lines: items,
	})
}

type PostDepreciationRequest struct {
	JournalID uint64 `json:"journal_id" validate:"required,gt=0"`
	Date      string `json:"date" validate:"required"`
}

// @Summary Post period depreciation
// @Description Posts the next due depreciation line to the general ledger (Dr Depreciation Expense / Cr Accumulated Depreciation) and marks it posted.
// @Tags Fixed Assets
// @Accept json
// @Produce json
// @Param id path integer true "Fixed asset ID"
// @Param body body PostDepreciationRequest true "Posting details"
// @Success 201 {object} AssetDepreciationLineResponse "Depreciation posted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Fixed asset not found"
// @Failure 422 {object} httpx.ErrorResponse "Nothing to post or asset not running"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/fixed-assets/{id}/post-depreciation [post]
func (h AssetCategoryHandler) PostDepreciation(c fiber.Ctx) error {
	var request PostDepreciationRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDateStr(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	assetRecord, err := h.svc.Find(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("fixed asset get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve fixed asset.", err)
	}
	if assetRecord == nil || !httpx.OwnsTenant(c, helper.Ptr(assetRecord.OrganizationID)) {
		return httpx.CreateNotFoundResponse(c, "Fixed asset not found.")
	}
	posted, err := h.svc.PostDepreciation(c, asset.PostDepreciationRequest{AssetID: assetRecord.ID, JournalID: request.JournalID, Date: date})
	if err != nil {
		return writeAssetError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Depreciation posted successfully.", newDepreciationLineResponse(posted))
}

type DisposeAssetRequest struct {
	JournalID         uint64  `json:"journal_id" validate:"required,gt=0"`
	Date              string  `json:"date" validate:"required"`
	State             string  `json:"state" validate:"required,oneof=disposed sold"`
	ProceedsAmount    float64 `json:"proceeds_amount"`
	ProceedsAccountID *uint64 `json:"proceeds_account_id"`
}

// @Summary Dispose or sell fixed asset
// @Description Disposes or sells a running asset, posting its removal from the books including realized gain or loss.
// @Tags Fixed Assets
// @Accept json
// @Produce json
// @Param id path integer true "Fixed asset ID"
// @Param body body DisposeAssetRequest true "Disposal details"
// @Success 200 {object} GetFixedAssetResponseEnvelope "Fixed asset disposed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Fixed asset not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid state or proceeds"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/fixed-assets/{id}/dispose [post]
func (h AssetCategoryHandler) DisposeAsset(c fiber.Ctx) error {
	var request DisposeAssetRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDateStr(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	assetRecord, err := h.svc.Find(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("fixed asset get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve fixed asset.", err)
	}
	if assetRecord == nil || !httpx.OwnsTenant(c, helper.Ptr(assetRecord.OrganizationID)) {
		return httpx.CreateNotFoundResponse(c, "Fixed asset not found.")
	}
	updated, err := h.svc.Dispose(c, asset.DisposalRequest{
		AssetID:           assetRecord.ID,
		JournalID:         request.JournalID,
		Date:              date,
		State:             request.State,
		ProceedsAmount:    request.ProceedsAmount,
		ProceedsAccountID: request.ProceedsAccountID,
	})
	if err != nil {
		return writeAssetError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Fixed asset disposed successfully.", GetFixedAssetResponse{
		FixedAsset: newFixedAssetResponse(updated),
	})
}

func writeAssetError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, asset.ErrAssetNotFound),
		errors.Is(err, asset.ErrAssetCategoryNotFound),
		errors.Is(err, asset.ErrAssetInvoiceLineNotFound),
		errors.Is(err, asset.ErrAssetInvoiceNotSupplierBill):
		return httpx.CreateNotFoundResponse(c, "Resource not found.")
	case errors.Is(err, asset.ErrAssetNameRequired),
		errors.Is(err, asset.ErrAssetInvalidState),
		errors.Is(err, asset.ErrAssetInvalidMethod),
		errors.Is(err, asset.ErrAssetInvalidPeriods),
		errors.Is(err, asset.ErrAssetInvalidDates),
		errors.Is(err, asset.ErrAssetInvalidValues),
		errors.Is(err, asset.ErrAssetCategoryAccounts),
		errors.Is(err, asset.ErrAssetNotRunning),
		errors.Is(err, asset.ErrAssetScheduleExists),
		errors.Is(err, asset.ErrAssetNoSchedule),
		errors.Is(err, asset.ErrAssetLineNotFound),
		errors.Is(err, asset.ErrAssetNothingToPost),
		errors.Is(err, asset.ErrAssetLinePosted),
		errors.Is(err, asset.ErrAssetDisposalProceeds):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Request cannot be processed.", nil)
	default:
		httpx.RequestLog(c).Error("asset write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process request.", err)
	}
}
