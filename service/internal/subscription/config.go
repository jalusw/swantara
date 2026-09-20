package subscription

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/systemconfig"
)

const (
	configSubscriptionJournalID           = "subscription.journal_id"
	configSubscriptionDeferredRevenueAcct = "subscription.deferred_revenue_account_id"
)

type SubscriptionConfigSource interface {
	JournalID(ctx context.Context, organizationID uint64) (uint64, error)
	DeferredRevenueAccountID(ctx context.Context, organizationID uint64) (uint64, error)
}

type subscriptionConfigSource struct {
	configs systemconfig.Lookup
}

func NewSubscriptionConfigSource(configs systemconfig.Lookup) SubscriptionConfigSource {
	return subscriptionConfigSource{configs: configs}
}

func (s subscriptionConfigSource) JournalID(ctx context.Context, organizationID uint64) (uint64, error) {
	var value uint64
	if err := systemconfig.Read(ctx, s.configs, organizationID, configSubscriptionJournalID, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (s subscriptionConfigSource) DeferredRevenueAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	var value uint64
	if err := systemconfig.Read(ctx, s.configs, organizationID, configSubscriptionDeferredRevenueAcct, &value); err != nil {
		return 0, err
	}
	return value, nil
}
