package httpx

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
)

var (
	ErrIdempotencyKeyInFlight = errors.New("idempotency request is still in progress")
	ErrIdempotencyKeyReused   = errors.New("idempotency key was already used for a different operation")
)

type IdempotencyEntry struct {
	StatusCode int
	Response   []byte
}

type IdempotencyReader interface {
	Claim(ctx context.Context, organizationID *uint64, key, resource string) (*IdempotencyEntry, error)
	Complete(ctx context.Context, organizationID *uint64, key string, statusCode int, response []byte) error
	Release(ctx context.Context, organizationID *uint64, key string) error
}

func IdempotencyGuard(reader IdempotencyReader, resource string) fiber.Handler {
	if reader == nil {
		return func(c fiber.Ctx) error { return c.Next() }
	}
	return func(c fiber.Ctx) error {
		key := c.Get("Idempotency-Key")
		if key == "" {
			return c.Next()
		}

		organizationID := TenantOrganizationID(c, nil)
		if organizationID == nil {
			RequestLog(c).Warn("idempotency guard skipped without organization", "key", key, "resource", resource)
			return c.Next()
		}

		entry, err := reader.Claim(c.Context(), organizationID, key, resource)
		if err != nil {
			switch {
			case errors.Is(err, ErrIdempotencyKeyReused):
				return CreateUnprocessableEntityErrorResponse(c, "Idempotency key was already used for a different operation.", nil)
			case errors.Is(err, ErrIdempotencyKeyInFlight):
				return CreateConflictResponse(c, "A request with this idempotency key is currently in progress.", nil)
			default:
				RequestLog(c).Error("idempotency claim failed", "key", key, "resource", resource, "error", err)
				return CreateInternalServerErrorResponse(c, "Failed to check idempotency.", err)
			}
		}
		if entry != nil {
			c.Status(entry.StatusCode)
			return c.Send(entry.Response)
		}

		if err := c.Next(); err != nil {
			if releaseErr := reader.Release(c.Context(), organizationID, key); releaseErr != nil {
				RequestLog(c).Warn("idempotency release failed", "key", key, "resource", resource, "error", releaseErr)
			}
			return err
		}

		status := c.Response().StatusCode()
		if status >= 200 && status < 300 {
			if err := reader.Complete(c.Context(), organizationID, key, status, c.Response().Body()); err != nil {
				RequestLog(c).Error("idempotency record failed", "key", key, "resource", resource, "error", err)
			}
			return nil
		}

		if releaseErr := reader.Release(c.Context(), organizationID, key); releaseErr != nil {
			RequestLog(c).Warn("idempotency release failed", "key", key, "resource", resource, "error", releaseErr)
		}
		return nil
	}
}
