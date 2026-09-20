package accounting

import (
	"context"
	"encoding/json"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

const (
	configBankFeeAccountID      = "accounting.bank_fee_account_id"
	configBankInterestAccountID = "accounting.bank_interest_account_id"
)

type BankChargeResolver interface {
	FeeAccountID(ctx context.Context, organizationID uint64) (uint64, error)
	InterestAccountID(ctx context.Context, organizationID uint64) (uint64, error)
}

type bankChargeResolver struct {
	configs  configLister
	accounts AccountLookup
}

func NewBankChargeResolver(configs configLister, accounts AccountLookup) BankChargeResolver {
	return bankChargeResolver{configs: configs, accounts: accounts}
}

func (r bankChargeResolver) FeeAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	if id, err := r.read(ctx, organizationID, configBankFeeAccountID); err == nil && id != 0 {
		return id, nil
	}
	return r.fallbackByType(ctx, organizationID, AccountTypeExpense)
}

func (r bankChargeResolver) InterestAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	if id, err := r.read(ctx, organizationID, configBankInterestAccountID); err == nil && id != 0 {
		return id, nil
	}
	return r.fallbackByType(ctx, organizationID, AccountTypeIncome)
}

func (r bankChargeResolver) read(ctx context.Context, organizationID uint64, key string) (uint64, error) {
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

func (r bankChargeResolver) fallbackByType(ctx context.Context, organizationID uint64, accountType string) (uint64, error) {
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
