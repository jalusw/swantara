package accounting

import (
	"context"
	"encoding/json"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

const (
	configCustomerAdvanceAccountID = "accounting.customer_advance_account_id"
	configSupplierAdvanceAccountID = "accounting.supplier_advance_account_id"
)

type AdvanceAccountResolver interface {
	CustomerAdvanceAccountID(ctx context.Context, organizationID uint64) (uint64, error)
	SupplierAdvanceAccountID(ctx context.Context, organizationID uint64) (uint64, error)
}

type advanceAccountResolver struct {
	configs  configLister
	accounts AccountLookup
}

func NewAdvanceAccountResolver(configs configLister, accounts AccountLookup) AdvanceAccountResolver {
	return advanceAccountResolver{configs: configs, accounts: accounts}
}

func (r advanceAccountResolver) CustomerAdvanceAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	if id, err := r.read(ctx, organizationID, configCustomerAdvanceAccountID); err == nil && id != 0 {
		return id, nil
	}
	return r.fallbackByType(ctx, organizationID, AccountTypeReceivable)
}

func (r advanceAccountResolver) SupplierAdvanceAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	if id, err := r.read(ctx, organizationID, configSupplierAdvanceAccountID); err == nil && id != 0 {
		return id, nil
	}
	return r.fallbackByType(ctx, organizationID, AccountTypePayable)
}

func (r advanceAccountResolver) read(ctx context.Context, organizationID uint64, key string) (uint64, error) {
	page, err := r.configs.List(ctx, &query.Query{Filters: []query.Filter{{Field: "key", Operator: query.Equal, Value: key}}})
	if err != nil {
		return 0, err
	}
	for _, cfg := range page.Items {
		if cfg.Key != key {
			continue
		}
		if cfg.OrganizationID != nil && *cfg.OrganizationID != organizationID {
			continue
		}
		if len(cfg.Value) == 0 {
			continue
		}
		var v uint64
		if err := json.Unmarshal(cfg.Value, &v); err != nil {
			continue
		}
		if v != 0 {
			return v, nil
		}
	}
	return 0, ErrFxAccountNotFound
}

func (r advanceAccountResolver) fallbackByType(ctx context.Context, organizationID uint64, accountType string) (uint64, error) {
	if r.accounts == nil {
		return 0, ErrFxAccountNotFound
	}
	page, err := r.accounts.List(ctx, &query.Query{Filters: []query.Filter{{Field: "type", Operator: query.Equal, Value: accountType}}})
	if err != nil {
		return 0, err
	}
	for _, acc := range page.Items {
		if acc.OrganizationID == organizationID && acc.Active {
			return acc.ID, nil
		}
	}
	return 0, ErrFxAccountNotFound
}
