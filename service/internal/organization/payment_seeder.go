package organization

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type paymentTermSeeder struct {
	terms reference.PaymentTermDAO
}

func NewPaymentTermSeeder(terms reference.PaymentTermDAO) PaymentTermSeeder {
	return paymentTermSeeder{terms: terms}
}

func (s paymentTermSeeder) SeedForOrganization(ctx context.Context, organizationID uint64) error {
	return reference.SeedDefaultPaymentTerms(ctx, s.terms, organizationID)
}
