package httpx

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
)

const (
	RequiredCode  = "REQUIRED"
	EmailCode     = "EMAIL"
	E164Code      = "E164"
	MinLengthCode = "MIN_LENGTH"
)

func BindAndValidate(c fiber.Ctx, request any) bool {
	if !checkRequestFormatAllowed(c) {
		return false
	}
	if err := BindBody(c, request); err != nil {
		if writeErr := CreateBadRequestResponse(c, "", err); writeErr != nil {
			RequestLog(c).Error("failed to write bind error response", "error", writeErr)
		}
		return false
	}

	if err := ValidateRequest(request); err != nil {
		if writeErr := CreateUnprocessableEntityErrorResponse(c, "", err); writeErr != nil {
			RequestLog(c).Error("failed to write validation error response", "error", writeErr)
		}
		return false
	}

	return true
}

// BindAndValidateWithFormats opts the handler in to the given request body
// formats before binding. JSON is the default.
func BindAndValidateWithFormats(c fiber.Ctx, request any, formats ...Format) bool {
	AllowRequestFormats(c, formats...)
	return BindAndValidate(c, request)
}

func checkRequestFormatAllowed(c fiber.Ctx) bool {
	if len(c.Body()) == 0 || isFormContentType(c) {
		return true
	}
	bodyFormat := RequestBodyFormat(c)
	if bodyFormat == "" {
		if writeErr := CreateUnsupportedMediaTypeResponse(c, "Unsupported media type.", nil); writeErr != nil {
			RequestLog(c).Error("failed to write unsupported media type response", "error", writeErr)
		}
		return false
	}
	if !IsRequestFormatAllowed(c, bodyFormat) {
		RequestLog(c).Warn("unsupported request format", "format", string(bodyFormat))
		if writeErr := CreateUnsupportedMediaTypeResponse(c, "Format "+string(bodyFormat)+" is not available for this endpoint.", nil); writeErr != nil {
			RequestLog(c).Error("failed to write unsupported media type response", "error", writeErr)
		}
		return false
	}
	return true
}

func ClientIP(c fiber.Ctx) string {
	return c.IP()
}

func UserAgent(c fiber.Ctx) string {
	ua := c.Get("User-Agent")
	forwardedUA := c.Get("X-Forwarded-User-Agent")
	if forwardedUA != "" {
		return forwardedUA
	}
	return ua
}

func DeviceName(c fiber.Ctx) string {
	return helper.ParseDeviceName(UserAgent(c))
}

func Browser(c fiber.Ctx) string {
	return helper.ParseBrowser(UserAgent(c))
}

func OS(c fiber.Ctx) string {
	return helper.ParseOS(UserAgent(c))
}

var requestValidator = newRequestValidator()

func newRequestValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}

func ValidateRequest(data any) *[]FieldError {
	if err := requestValidator.Struct(data); err != nil {
		var validationErrors validator.ValidationErrors
		if !errors.As(err, &validationErrors) {
			return &[]FieldError{
				{
					Code:    "validation_error",
					Message: err.Error(),
				},
			}
		}
		errs := make([]FieldError, 0, len(validationErrors))
		for _, e := range validationErrors {
			errs = append(errs, buildFieldError(e))
		}
		return &errs
	}

	return nil
}

func buildFieldError(e validator.FieldError) FieldError {
	field := e.Field()
	code := strings.ToUpper(e.Tag())
	message := fmt.Sprintf("%s is invalid.", field)

	switch e.Tag() {
	case "required":
		code = RequiredCode
		message = fmt.Sprintf("%s is required.", field)
	case "required_if", "required_unless", "required_with", "required_without":
		code = RequiredCode
		message = fmt.Sprintf("%s is required.", field)
	case "email":
		code = EmailCode
		message = fmt.Sprintf("%s must be a valid email.", field)
	case "e164":
		code = E164Code
		message = fmt.Sprintf("%s must be a valid phone number.", field)
	case "min":
		code = MinLengthCode
		message = fmt.Sprintf("%s must be at least %s characters.", field, e.Param())
	case "max":
		message = fmt.Sprintf("%s must be at most %s characters.", field, e.Param())
	case "gt":
		message = fmt.Sprintf("%s must be greater than %s.", field, e.Param())
	case "gte":
		message = fmt.Sprintf("%s must be greater than or equal to %s.", field, e.Param())
	case "lt":
		message = fmt.Sprintf("%s must be less than %s.", field, e.Param())
	case "lte":
		message = fmt.Sprintf("%s must be less than or equal to %s.", field, e.Param())
	case "oneof":
		message = fmt.Sprintf("%s must be one of: %s.", field, e.Param())
	}

	return FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	}
}

func requestElapsed(c fiber.Ctx) string {
	if start, ok := c.Locals(requestStartKey).(time.Time); ok {
		return time.Since(start).Round(time.Microsecond).String()
	}
	return ""
}
