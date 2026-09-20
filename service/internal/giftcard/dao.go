package giftcard

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type GiftCardDAO interface {
	dao.CRUD[GiftCard]
	CreateTx(ctx context.Context, tx *gorm.DB, entity *GiftCard) (*GiftCard, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, entity *GiftCard) (*GiftCard, error)
	ListDueForExpiry(ctx context.Context, organizationID uint64, asOf time.Time) ([]*GiftCard, error)
}

type giftCardDAO struct {
	dao.Base[GiftCard]
	db *gorm.DB
}

func NewGiftCardDAO(db *gorm.DB) GiftCardDAO {
	return giftCardDAO{Base: dao.NewBase[GiftCard](db), db: db}
}

func (d giftCardDAO) CreateTx(ctx context.Context, tx *gorm.DB, entity *GiftCard) (*GiftCard, error) {
	if err := tx.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (d giftCardDAO) UpdateTx(ctx context.Context, tx *gorm.DB, entity *GiftCard) (*GiftCard, error) {
	if err := tx.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (d giftCardDAO) ListDueForExpiry(ctx context.Context, organizationID uint64, asOf time.Time) ([]*GiftCard, error) {
	var cards []GiftCard
	err := d.db.WithContext(ctx).
		Where("organization_id = ? AND state = ? AND balance > 0 AND expiry_date IS NOT NULL AND expiry_date < ? AND deleted_at IS NULL", organizationID, GiftCardStateActive, asOf).
		Order("expiry_date").
		Find(&cards).Error
	if err != nil {
		return nil, err
	}
	result := make([]*GiftCard, 0, len(cards))
	for i := range cards {
		result = append(result, &cards[i])
	}
	return result, nil
}

type GiftCardTransactionDAO interface {
	dao.CRUD[GiftCardTransaction]
	CreateTx(ctx context.Context, tx *gorm.DB, entity *GiftCardTransaction) (*GiftCardTransaction, error)
	ListByGiftCard(ctx context.Context, giftCardID uint64) ([]*GiftCardTransaction, error)
}

type giftCardTransactionDAO struct {
	dao.Base[GiftCardTransaction]
	db *gorm.DB
}

func NewGiftCardTransactionDAO(db *gorm.DB) GiftCardTransactionDAO {
	return giftCardTransactionDAO{Base: dao.NewBase[GiftCardTransaction](db), db: db}
}

func (d giftCardTransactionDAO) CreateTx(ctx context.Context, tx *gorm.DB, entity *GiftCardTransaction) (*GiftCardTransaction, error) {
	if err := tx.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (d giftCardTransactionDAO) ListByGiftCard(ctx context.Context, giftCardID uint64) ([]*GiftCardTransaction, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "gift_card_id", Operator: query.Equal, Value: giftCardID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
