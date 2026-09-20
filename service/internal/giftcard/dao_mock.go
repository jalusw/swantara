package giftcard

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type GiftCardDAOMock struct {
	dao.CRUDMock[GiftCard]
	CreateTxFunc         func(ctx context.Context, tx *gorm.DB, entity *GiftCard) (*GiftCard, error)
	UpdateTxFunc         func(ctx context.Context, tx *gorm.DB, entity *GiftCard) (*GiftCard, error)
	ListDueForExpiryFunc func(ctx context.Context, organizationID uint64, asOf time.Time) ([]*GiftCard, error)
}

func (m GiftCardDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, entity *GiftCard) (*GiftCard, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}

func (m GiftCardDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, entity *GiftCard) (*GiftCard, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}

func (m GiftCardDAOMock) ListDueForExpiry(ctx context.Context, organizationID uint64, asOf time.Time) ([]*GiftCard, error) {
	if m.ListDueForExpiryFunc != nil {
		return m.ListDueForExpiryFunc(ctx, organizationID, asOf)
	}
	return []*GiftCard{}, nil
}

type GiftCardTransactionDAOMock struct {
	dao.CRUDMock[GiftCardTransaction]
	CreateTxFunc       func(ctx context.Context, tx *gorm.DB, entity *GiftCardTransaction) (*GiftCardTransaction, error)
	ListByGiftCardFunc func(ctx context.Context, giftCardID uint64) ([]*GiftCardTransaction, error)
}

func (m GiftCardTransactionDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, entity *GiftCardTransaction) (*GiftCardTransaction, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}

func (m GiftCardTransactionDAOMock) ListByGiftCard(ctx context.Context, giftCardID uint64) ([]*GiftCardTransaction, error) {
	if m.ListByGiftCardFunc != nil {
		return m.ListByGiftCardFunc(ctx, giftCardID)
	}
	return []*GiftCardTransaction{}, nil
}

type CouponDAOMock struct {
	dao.CRUDMock[Coupon]
	RedeemTxFunc func(ctx context.Context, tx *gorm.DB, couponID uint64) (*Coupon, error)
}

func (m CouponDAOMock) RedeemTx(ctx context.Context, tx *gorm.DB, couponID uint64) (*Coupon, error) {
	if m.RedeemTxFunc != nil {
		return m.RedeemTxFunc(ctx, tx, couponID)
	}
	return nil, nil
}
