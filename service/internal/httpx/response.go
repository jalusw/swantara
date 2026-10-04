package httpx

import (
	"cmp"
	"encoding/xml"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

var responseVersion = "dev"

type SortMeta struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type FilterMeta map[string]string

type Meta struct {
	Timestamp  string          `json:"timestamp"`
	RequestID  string          `json:"request_id,omitempty"`
	Version    string          `json:"version,omitempty"`
	Elapsed    string          `json:"elapsed,omitempty"`
	Sort       *SortMeta       `json:"sort,omitempty"`
	Filter     FilterMeta      `json:"filter,omitempty"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type Response[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Meta    *Meta  `json:"meta,omitempty"`
	Links   []Link `json:"links,omitempty"`
}

type SuccessResponse = Response[any]

type EnvelopeBase struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Meta    *Meta  `json:"meta,omitempty"`
	Links   []Link `json:"links,omitempty"`
}

type EmptyEnvelope struct {
	EnvelopeBase
	Data any `json:"data"`
}

type ErrorResponse struct {
	Success     bool          `json:"success"`
	Message     string        `json:"message"`
	Detail      string        `json:"error,omitempty"`
	ErrorCode   string        `json:"error_code"`
	FieldErrors *[]FieldError `json:"field_errors,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func CreateSuccessResponse[T any](c fiber.Ctx, message string, data T) error {
	return writeSuccess(c, http.StatusOK, message, data, nil)
}

func CreateCreatedResponse[T any](c fiber.Ctx, message string, data T) error {
	return writeSuccess(c, http.StatusCreated, message, data, nil)
}

func CreateAcceptedResponse[T any](c fiber.Ctx, message string, data T) error {
	return writeSuccess(c, http.StatusAccepted, message, data, nil)
}

func writeSuccess[T any](c fiber.Ctx, status int, message string, data T, links []Link) error {
	return writeSuccessWithEnvelope(c, status, Response[T]{
		Success: true,
		Message: cmp.Or(message, DefaultSuccessMessage),
		Data:    data,
		Meta:    populateMeta(c, &Meta{}),
		Links:   links,
	})
}

func writeSuccessWithEnvelope[T any](c fiber.Ctx, status int, envelope Response[T]) error {
	negotiated := NegotiatedResponseFormat(c)
	explicit, hasExplicit := explicitQueryFormat(c)
	switch negotiated {
	case FormatXML:
		return c.Status(status).XML(xmlResponse[T]{
			Success: envelope.Success,
			Message: envelope.Message,
			Data:    envelope.Data,
			Meta:    xmlSafeMeta(envelope.Meta),
			Links:   envelope.Links,
		})
	case FormatCSV:
		if isCSVList(envelope.Meta, envelope.Data) {
			items, ok := csvItems(envelope.Data)
			if !ok {
				return c.Status(status).JSON(envelope)
			}
			return exportCSVWithMeta(c, status, items, envelope.Meta, envelope.Links)
		}
		if HasExplicitFormatQuery(c, FormatCSV) {
			return writeError(c, http.StatusBadRequest, ErrBadRequestCode, "CSV export is not available for this endpoint.", "", nil)
		}
		return c.Status(status).JSON(envelope)
	default:
		if hasExplicit && explicit != FormatJSON {
			return writeError(c, http.StatusBadRequest, ErrBadRequestCode, "Format "+string(explicit)+" is not available for this endpoint.", "", nil)
		}
		return c.Status(status).JSON(envelope)
	}
}

type xmlResponse[T any] struct {
	XMLName xml.Name `xml:"response"`
	Success bool     `xml:"success"`
	Message string   `xml:"message"`
	Data    T        `xml:"data"`
	Meta    *xmlMeta `xml:"meta"`
	Links   []Link   `xml:"link"`
}

type xmlMeta struct {
	Timestamp  string          `xml:"timestamp"`
	RequestID  string          `xml:"request_id"`
	Version    string          `xml:"version"`
	Elapsed    string          `xml:"elapsed"`
	Sort       *SortMeta       `xml:"sort"`
	Pagination *PaginationMeta `xml:"pagination"`
}

func xmlSafeMeta(meta *Meta) *xmlMeta {
	if meta == nil {
		return nil
	}
	return &xmlMeta{
		Timestamp:  meta.Timestamp,
		RequestID:  meta.RequestID,
		Version:    meta.Version,
		Elapsed:    meta.Elapsed,
		Sort:       meta.Sort,
		Pagination: meta.Pagination,
	}
}

func CreateNoContentResponse(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

func CreateSuccessResponseWithMeta[T any](
	c fiber.Ctx,
	message string,
	data T,
	meta Meta,
) error {
	meta = *populateMeta(c, &meta)
	return writeSuccessWithEnvelope(c, http.StatusOK, Response[T]{
		Success: true,
		Message: cmp.Or(message, DefaultSuccessMessage),
		Data:    data,
		Meta:    &meta,
		Links:   paginationLinks(c, meta.Pagination),
	})
}

func CreateSuccessResponseWithLinks[T any](
	c fiber.Ctx,
	message string,
	data T,
	links []Link,
) error {
	return writeSuccessWithEnvelope(c, http.StatusOK, Response[T]{
		Success: true,
		Message: cmp.Or(message, DefaultSuccessMessage),
		Data:    data,
		Meta:    populateMeta(c, &Meta{}),
		Links:   links,
	})
}

func CreateCreatedResponseWithLinks[T any](
	c fiber.Ctx,
	message string,
	data T,
	links []Link,
) error {
	return writeSuccessWithEnvelope(c, http.StatusCreated, Response[T]{
		Success: true,
		Message: cmp.Or(message, DefaultSuccessMessage),
		Data:    data,
		Meta:    populateMeta(c, &Meta{}),
		Links:   links,
	})
}

func CreateErrorResponse(message, errorCode string, err error) ErrorResponse {
	return ErrorResponse{
		Success:   false,
		Message:   cmp.Or(message, DefaultErrorMessage),
		ErrorCode: errorCode,
		Detail:    errDetail(err),
	}
}

func CreateBadRequestResponse(c fiber.Ctx, message string, err error) error {
	setResponseProblem(c, ErrBadRequestCode, "", nil)
	return writeError(c, http.StatusBadRequest, ErrBadRequestCode, cmp.Or(message, DefaultBadRequestMessage), "", nil)
}

func CreateUnauthorizedErrorResponse(c fiber.Ctx, message string, err error) error {
	return writeError(c, http.StatusUnauthorized, ErrUnauthorizedCode, cmp.Or(message, DefaultUnauthorizedMessage), "", nil)
}

func CreateTokenReusedErrorResponse(c fiber.Ctx, message string, err error) error {
	return writeError(c, http.StatusUnauthorized, ErrTokenReusedCode, cmp.Or(message, DefaultUnauthorizedMessage), "", nil)
}

func CreateForbiddenErrorResponse(c fiber.Ctx, message string, err error) error {
	return writeError(c, http.StatusForbidden, ErrForbiddenCode, cmp.Or(message, "You don't have permission to access this resource."), "", nil)
}

func CreateNotFoundResponse(c fiber.Ctx, message string) error {
	return writeError(c, http.StatusNotFound, ErrNotFoundCode, cmp.Or(message, DefaultNotFoundMessage), "", nil)
}

func CreateConflictResponse(c fiber.Ctx, message string, err error) error {
	return writeError(c, http.StatusConflict, ErrConflictCode, cmp.Or(message, "The resource is in a state that cannot fulfill the request."), "", nil)
}

func CreateUnprocessableEntityErrorResponse(c fiber.Ctx, message string, fieldErrors *[]FieldError) error {
	setResponseProblem(c, ErrUnprocessableCode, "Can't process the entity", fieldErrors)
	return writeError(c, http.StatusUnprocessableEntity, ErrUnprocessableCode, cmp.Or(message, DefaultUnprocessableMessage), "Can't process the entity", fieldErrors)
}

func CreateInternalServerErrorResponse(c fiber.Ctx, message string, err error) error {
	setResponseProblem(c, ErrInternalServerCode, errDetail(err), nil)
	return writeError(c, http.StatusInternalServerError, ErrInternalServerCode, cmp.Or(message, DefaultInternalServerMessage), DefaultInternalServerMessage, nil)
}

func CreateServiceUnavailableErrorResponse(c fiber.Ctx, message string, err error) error {
	setResponseProblem(c, ErrServiceUnavailableCode, errDetail(err), nil)
	return writeError(c, http.StatusServiceUnavailable, ErrServiceUnavailableCode, cmp.Or(message, "Service is unavailable."), "Service is unavailable.", nil)
}

func CreateUnsupportedMediaTypeResponse(c fiber.Ctx, message string, err error) error {
	setResponseProblem(c, ErrUnsupportedMediaTypeCode, errDetail(err), nil)
	return writeError(c, http.StatusUnsupportedMediaType, ErrUnsupportedMediaTypeCode, cmp.Or(message, "Unsupported media type."), errDetail(err), nil)
}

// WriteListResponse centralizes list negotiation: JSON default, XML/CSV only
// when negotiated. data is the JSON envelope payload (e.g.
// ListContactsResponse), items is the slice used for CSV export (may be nil
// to extract from data), filename is the CSV download name. Replaces the
// per-handler:
//
//	if httpx.RequestFormat(c) == httpx.FormatCSV {
//	    return httpx.ExportCSV(c, fiber.StatusOK, "x.csv", items)
//	}
//	return httpx.CreateSuccessResponseWithMeta(...)
func WriteListResponse(c fiber.Ctx, message string, data any, items any, meta Meta, filename string) error {
	if NegotiatedResponseFormat(c) == FormatCSV {
		exportItems := items
		if exportItems == nil {
			if extracted, ok := csvItems(data); ok {
				exportItems = extracted
			}
		}
		if exportItems != nil {
			if filename == "" {
				filename = "export.csv"
			}
			populated := populateMeta(c, &meta)
			setCSVListHeaders(c, populated, paginationLinks(c, populated.Pagination))
			return ExportCSV(c, http.StatusOK, filename, exportItems)
		}
	}
	return CreateSuccessResponseWithMeta(c, message, data, meta)
}

func writeError(c fiber.Ctx, status int, errorCode, message, detail string, fieldErrors *[]FieldError) error {
	envelope := ErrorResponse{
		Success:     false,
		Message:     message,
		ErrorCode:   errorCode,
		Detail:      detail,
		FieldErrors: fieldErrors,
	}
	if NegotiatedResponseFormat(c) == FormatXML {
		return c.Status(status).XML(envelope)
	}
	return c.Status(status).JSON(envelope)
}

func errDetail(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func SetResponseVersion(version string) {
	if version != "" {
		responseVersion = version
	}
}

func populateMeta(c fiber.Ctx, meta *Meta) *Meta {
	meta.Timestamp = time.Now().UTC().Format(time.RFC3339)
	meta.RequestID = requestid.FromContext(c)
	meta.Version = responseVersion
	meta.Elapsed = requestElapsed(c)
	return meta
}

func isCSVList(meta *Meta, data any) bool {
	if meta != nil && meta.Pagination != nil {
		return true
	}
	value := reflect.ValueOf(data)
	return value.IsValid() && (value.Kind() == reflect.Slice || value.Kind() == reflect.Array)
}

func csvItems(data any) (any, bool) {
	value := reflect.ValueOf(data)
	if !value.IsValid() {
		return nil, false
	}
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		return data, true
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if field.PkgPath != "" {
				continue
			}
			kind := value.Field(i).Kind()
			if kind == reflect.Slice || kind == reflect.Array {
				return value.Field(i).Interface(), true
			}
		}
	}
	return nil, false
}

func exportCSVWithMeta(c fiber.Ctx, status int, items any, meta *Meta, links []Link) error {
	setCSVListHeaders(c, meta, links)
	return ExportCSV(c, status, "export.csv", items)
}

func setCSVListHeaders(c fiber.Ctx, meta *Meta, links []Link) {
	if meta != nil && meta.Pagination != nil {
		pagination := meta.Pagination
		c.Set("X-Pagination-Page", strconv.Itoa(pagination.Page))
		c.Set("X-Pagination-Per-Page", strconv.Itoa(pagination.PerPage))
		c.Set("X-Pagination-Total", strconv.Itoa(pagination.Total))
		c.Set("X-Pagination-Total-Pages", strconv.Itoa(pagination.TotalPages))
	}
	if len(links) > 0 {
		parts := make([]string, 0, len(links))
		for _, link := range links {
			parts = append(parts, "<"+link.Href+`>; rel="`+link.Rel+`"`)
		}
		c.Set(fiber.HeaderLink, strings.Join(parts, ", "))
	}
}
