package handler

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type handlerCurrencySearchMock struct {
	search func(ctx context.Context, field string, value any) (*reference.Currency, error)
}

func (m handlerCurrencySearchMock) Search(ctx context.Context, field string, value any) (*reference.Currency, error) {
	if m.search != nil {
		return m.search(ctx, field, value)
	}
	return nil, nil
}

type handlerMemberProvisionerMock struct {
	provision func(ctx context.Context, organizationID, ownerUserID uint64) error
}

func (m handlerMemberProvisionerMock) ProvisionOwner(ctx context.Context, organizationID, ownerUserID uint64) error {
	if m.provision != nil {
		return m.provision(ctx, organizationID, ownerUserID)
	}
	return nil
}
