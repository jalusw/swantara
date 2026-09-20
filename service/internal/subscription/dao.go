package subscription

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type SubscriptionDAO interface {
	dao.CRUD[Subscription]
	ListDue(ctx context.Context, asOf time.Time) ([]*Subscription, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, subscription *Subscription) (*Subscription, error)
}

type subscriptionDAO struct {
	dao.Base[Subscription]
	db *gorm.DB
}

func NewSubscriptionDAO(db *gorm.DB) SubscriptionDAO {
	return subscriptionDAO{Base: dao.NewBase[Subscription](db), db: db}
}

func (d subscriptionDAO) ListDue(ctx context.Context, asOf time.Time) ([]*Subscription, error) {
	var entities []Subscription
	if err := d.db.WithContext(ctx).
		Where("state = ? AND next_invoice_date IS NOT NULL AND next_invoice_date <= ?", SubscriptionStateActive, asOf).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	items := make([]*Subscription, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

func (d subscriptionDAO) UpdateTx(ctx context.Context, tx *gorm.DB, subscription *Subscription) (*Subscription, error) {
	if err := tx.WithContext(ctx).Save(subscription).Error; err != nil {
		return nil, err
	}
	return subscription, nil
}

type SubscriptionLineDAO interface {
	dao.CRUD[SubscriptionLine]
	ListBySubscription(ctx context.Context, subscriptionID uint64) ([]*SubscriptionLine, error)
}

type subscriptionLineDAO struct {
	dao.Base[SubscriptionLine]
	db *gorm.DB
}

func NewSubscriptionLineDAO(db *gorm.DB) SubscriptionLineDAO {
	return subscriptionLineDAO{Base: dao.NewBase[SubscriptionLine](db), db: db}
}

func (d subscriptionLineDAO) ListBySubscription(ctx context.Context, subscriptionID uint64) ([]*SubscriptionLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "subscription_id", Operator: query.Equal, Value: subscriptionID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
