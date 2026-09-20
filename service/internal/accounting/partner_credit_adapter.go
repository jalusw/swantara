package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
)

type contactCreditAdapter struct {
	customers contacts.CustomerProfileDAO
}

func NewContactCreditAdapter(customers contacts.CustomerProfileDAO) ContactCreditLimiter {
	return contactCreditAdapter{customers: customers}
}

func (a contactCreditAdapter) CreditLimit(ctx context.Context, contactID uint64) (*float64, error) {
	record, err := a.customers.FindByContact(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	return record.CreditLimit, nil
}
