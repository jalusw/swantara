package organization

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type OrgDAO interface {
	dao.CRUD[reference.Organization]
}

type CurrencyDAO interface {
	Search(ctx context.Context, field string, value any) (*reference.Currency, error)
}

type MemberProvisioner interface {
	ProvisionOwner(ctx context.Context, organizationID, ownerUserID uint64) error
}

type PaymentTermSeeder interface {
	SeedForOrganization(ctx context.Context, organizationID uint64) error
}

type ModuleStore interface {
	dao.CRUD[reference.OrganizationModule]
}
