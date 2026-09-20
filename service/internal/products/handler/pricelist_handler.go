package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

type PriceBookHandler struct {
	svc products.ProductService
}

func NewPriceBookHandler(svc products.ProductService) PriceBookHandler {
	return PriceBookHandler{svc: svc}
}

type PriceBookResponse struct {
	ID             uint64    `json:"id"`
	Name           string    `json:"name"`
	CurrencyCode   *string   `json:"currency_code"`
	OrganizationID *uint64   `json:"organization_id"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newPriceBookResponse(price_book *products.PriceBook) PriceBookResponse {
	return PriceBookResponse{
		ID:             price_book.ID,
		Name:           price_book.Name,
		CurrencyCode:   price_book.CurrencyCode,
		OrganizationID: price_book.OrganizationID,
		Active:         price_book.Active,
		CreatedAt:      price_book.CreatedAt,
		UpdatedAt:      price_book.UpdatedAt,
	}
}

var price_bookQueryAllowlist = map[string]struct{}{
	"name":            {},
	"currency_code":   {},
	"organization_id": {},
	"active":          {},
	"created_at":      {},
	"updated_at":      {},
}

type ListPriceBooksResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPriceBooksResponse `json:"data"`
}
type ListPriceBooksResponse struct {
	PriceBooks []PriceBookResponse `json:"price_books"`
}

// @Summary List price_books
// @Description Lists price lists with pagination, sorting, and filtering, automatically restricting results to the caller's organization; the response can be exported as JSON, XML, or CSV.
// @Tags PriceBooks
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. active:eq:true)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListPriceBooksResponseEnvelope "PriceBooks retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/price_books [get]
func (h PriceBookHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, price_bookQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.ListPriceBooks(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("price_book list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve price_books.", err)
	}

	items := make([]PriceBookResponse, len(page.Items))
	for i, price_book := range page.Items {
		items[i] = newPriceBookResponse(price_book)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "price_books.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "PriceBooks retrieved successfully.", ListPriceBooksResponse{
		PriceBooks: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetPriceBookResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetPriceBookResponse `json:"data"`
}
type GetPriceBookResponse struct {
	PriceBook PriceBookResponse `json:"price_book"`
}

// @Summary Get price_book
// @Description Returns a single price list by id with its name, currency, and active status. The price list must belong to the caller's organization, otherwise a 404 is returned.
// @Tags PriceBooks
// @Accept json
// @Produce json
// @Param id path integer true "PriceBook ID"
// @Success 200 {object} GetPriceBookResponseEnvelope "PriceBook retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "PriceBook not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/price_books/{id} [get]
func (h PriceBookHandler) Get(c fiber.Ctx) error {
	price_bookID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid price_book id provided.", nil)
	}

	price_book, err := h.svc.FindPriceBook(c, price_bookID)
	if err != nil {
		httpx.RequestLog(c).Error("price_book lookup failed", "price_book_id", price_bookID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get price_book.", err)
	}
	if price_book == nil || !httpx.OwnsTenant(c, price_book.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "PriceBook not found.")
	}

	return httpx.CreateSuccessResponse(c, "PriceBook retrieved successfully.", GetPriceBookResponse{
		PriceBook: newPriceBookResponse(price_book),
	})
}

type CreatePriceBookRequest struct {
	Name           string  `json:"name" validate:"required"`
	CurrencyCode   *string `json:"currency_code"`
	OrganizationID *uint64 `json:"organization_id"`
	Active         *bool   `json:"active"`
}

type CreatePriceBookResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreatePriceBookResponse `json:"data"`
}
type CreatePriceBookResponse struct {
	PriceBook PriceBookResponse `json:"price_book"`
}

// @Summary Create price_book
// @Description Creates a price list with a required name, an optional currency, and an active flag that defaults to true. The price list is associated with the caller's organization.
// @Tags PriceBooks
// @Accept json
// @Produce json
// @Param body body CreatePriceBookRequest true "PriceBook details"
// @Success 201 {object} CreatePriceBookResponseEnvelope "PriceBook created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/price_books [post]
func (h PriceBookHandler) Create(c fiber.Ctx) error {
	var request CreatePriceBookRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	active := true
	if request.Active != nil {
		active = *request.Active
	}
	price_book, err := h.svc.CreatePriceBook(c, &products.PriceBook{
		Name:           request.Name,
		CurrencyCode:   request.CurrencyCode,
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		Active:         active,
	})
	if err != nil {
		return writePriceBookError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "PriceBook created successfully.", CreatePriceBookResponse{
		PriceBook: newPriceBookResponse(price_book),
	})
}

type UpdatePriceBookRequest struct {
	Name           string  `json:"name" validate:"required"`
	CurrencyCode   *string `json:"currency_code"`
	OrganizationID *uint64 `json:"organization_id"`
	Active         *bool   `json:"active"`
}

type UpdatePriceBookResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdatePriceBookResponse `json:"data"`
}
type UpdatePriceBookResponse struct {
	PriceBook PriceBookResponse `json:"price_book"`
}

// @Summary Update price_book
// @Description Updates a price list's name, currency, and active status. A 404 is returned if the price list does not belong to the caller's organization.
// @Tags PriceBooks
// @Accept json
// @Produce json
// @Param id path integer true "PriceBook ID"
// @Param body body UpdatePriceBookRequest true "PriceBook details"
// @Success 200 {object} UpdatePriceBookResponseEnvelope "PriceBook updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "PriceBook not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/price_books/{id} [put]
func (h PriceBookHandler) Update(c fiber.Ctx) error {
	price_bookID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid price_book id provided.", nil)
	}

	var request UpdatePriceBookRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	price_book, err := h.svc.FindPriceBook(c, price_bookID)
	if err != nil {
		httpx.RequestLog(c).Error("price_book lookup failed", "price_book_id", price_bookID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update price_book.", err)
	}
	if price_book == nil || !httpx.OwnsTenant(c, price_book.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "PriceBook not found.")
	}

	price_book.Name = request.Name
	price_book.CurrencyCode = request.CurrencyCode
	price_book.OrganizationID = httpx.TenantOrganizationID(c, price_book.OrganizationID)
	if request.Active != nil {
		price_book.Active = *request.Active
	}

	updated, err := h.svc.UpdatePriceBook(c, price_book)
	if err != nil {
		httpx.RequestLog(c).Error("price_book update failed", "price_book_id", price_bookID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update price_book.", err)
	}

	return httpx.CreateSuccessResponse(c, "PriceBook updated successfully.", UpdatePriceBookResponse{
		PriceBook: newPriceBookResponse(updated),
	})
}

// @Summary Delete price_book
// @Description Deletes a price list by id. The price list must belong to the caller's organization; a 404 is returned otherwise.
// @Tags PriceBooks
// @Accept json
// @Produce json
// @Param id path integer true "PriceBook ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "PriceBook not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/price_books/{id} [delete]
func (h PriceBookHandler) Delete(c fiber.Ctx) error {
	price_bookID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid price_book id provided.", nil)
	}

	price_book, err := h.svc.FindPriceBook(c, price_bookID)
	if err != nil {
		httpx.RequestLog(c).Error("price_book lookup failed", "price_book_id", price_bookID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete price_book.", err)
	}
	if price_book == nil || !httpx.OwnsTenant(c, price_book.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "PriceBook not found.")
	}

	if err := h.svc.DeletePriceBook(c, price_bookID); err != nil {
		httpx.RequestLog(c).Error("price_book deletion failed", "price_book_id", price_bookID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete price_book.", err)
	}

	return httpx.CreateNoContentResponse(c)
}

func writePriceBookError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, products.ErrPriceBookNotFound):
		return httpx.CreateNotFoundResponse(c, "PriceBook not found.")
	case errors.Is(err, products.ErrVariantNotFound):
		return httpx.CreateNotFoundResponse(c, "Item variant not found.")
	case errors.Is(err, products.ErrItemNotFound):
		return httpx.CreateNotFoundResponse(c, "Item not found.")
	case errors.Is(err, products.ErrInvalidAppliesTo):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid rule applies_to scope.", nil)
	case errors.Is(err, products.ErrInvalidComputeType):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid rule compute type.", nil)
	case errors.Is(err, products.ErrInvalidRuleScope):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Rule scope requires a item or category.", nil)
	case errors.Is(err, products.ErrPriceUnavailable):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No price can be computed for the matching rule.", nil)
	default:
		httpx.RequestLog(c).Error("price_book write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save price_book.", err)
	}
}
