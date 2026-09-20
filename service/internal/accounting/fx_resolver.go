package accounting

import (
	"context"
	"encoding/json"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

const (
	configFxGainAccountID = "accounting.fx_gain_account_id"
	configFxLossAccountID = "accounting.fx_loss_account_id"
)

type configLister interface {
	List(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error)
}

type fxAccountResolver struct {
	configs  configLister
	accounts AccountLookup
}

func NewFxAccountResolver(configs configLister, accounts AccountLookup) FxAccountResolver {
	return fxAccountResolver{configs: configs, accounts: accounts}
}

func (r fxAccountResolver) FxGainAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	if id, err := r.read(ctx, organizationID, configFxGainAccountID); err == nil && id != 0 {
		return id, nil
	}
	if id, err := r.read(ctx, organizationID, "reporting.fx_gain_loss_account_id"); err == nil && id != 0 {
		return id, nil
	}
	return r.fallbackByType(ctx, organizationID, AccountTypeIncome)
}

func (r fxAccountResolver) FxLossAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	if id, err := r.read(ctx, organizationID, configFxLossAccountID); err == nil && id != 0 {
		return id, nil
	}
	if id, err := r.read(ctx, organizationID, "reporting.fx_gain_loss_account_id"); err == nil && id != 0 {
		return id, nil
	}
	return r.fallbackByType(ctx, organizationID, AccountTypeExpense)
}

func (r fxAccountResolver) read(ctx context.Context, organizationID uint64, key string) (uint64, error) {
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

func (r fxAccountResolver) fallbackByType(ctx context.Context, organizationID uint64, accountType string) (uint64, error) {
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
