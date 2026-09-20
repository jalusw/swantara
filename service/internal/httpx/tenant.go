package httpx

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

const OrganizationHeader = "Organization-ID"

func ResolveOrganizationID(c fiber.Ctx) (uint64, bool) {
	for _, key := range []string{"organization_id", "id"} {
		if param := c.Params(key); param != "" {
			organizationID, err := strconv.ParseUint(param, 10, 64)
			if err == nil && organizationID != 0 {
				return organizationID, true
			}
		}
	}

	if header := c.Get(OrganizationHeader); header != "" {
		organizationID, err := strconv.ParseUint(header, 10, 64)
		if err == nil && organizationID != 0 {
			return organizationID, true
		}
	}

	return 0, false
}

func TenantOrganizationID(c fiber.Ctx, fallback *uint64) *uint64 {
	if organizationID, ok := CallerOrganizationID(c); ok {
		id := organizationID
		return &id
	}
	return fallback
}

func ForceTenantFilter(c fiber.Ctx, parsedQuery *query.Query) error {
	organizationID, ok := CallerOrganizationID(c)
	if !ok {
		return ErrMissingOrganizationInContext
	}
	parsedQuery.Filters = append(parsedQuery.Filters, query.Filter{
		Field:    "organization_id",
		Operator: query.Equal,
		Value:    organizationID,
	})
	return nil
}

func OwnsTenant(c fiber.Ctx, organizationID *uint64) bool {
	tenantID, ok := CallerOrganizationID(c)
	if !ok {
		return false
	}
	return organizationID != nil && *organizationID == tenantID
}
