package giftcard

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func GiftCardFixture(opts ...func(*GiftCard) *GiftCard) *GiftCard {
	now := time.Now()
	gc := &GiftCard{
		Base:              model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID:    helper.Ptr(uint64(gofakeit.Number(1, 100))),
		Code:              gofakeit.AppName(),
		ContactID:         helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		InitialAmount:     gofakeit.Float64Range(1, 10000),
		Balance:           gofakeit.Float64Range(1, 10000),
		CurrencyCode:      "USD",
		ExpiryDate:        helper.Ptr(time.Now().AddDate(0, 1, 0)),
		State:             GiftCardStateActive,
		IssuedFromOrderID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
	}
	for _, opt := range opts {
		opt(gc)
	}
	return gc
}

func GiftCardTransactionFixture(opts ...func(*GiftCardTransaction) *GiftCardTransaction) *GiftCardTransaction {
	now := time.Now()
	tx := &GiftCardTransaction{
		Base:       model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		GiftCardID: uint64(gofakeit.Number(1, 10000)),
		Type:       TransactionIssue,
		Amount:     gofakeit.Float64Range(1, 10000),
		OrderType:  "sale_order",
		OrderID:    uint64(gofakeit.Number(1, 10000)),
		EntryID:    helper.Ptr(uint64(gofakeit.Number(1, 10000))),
	}
	for _, opt := range opts {
		opt(tx)
	}
	return tx
}

func CouponFixture(opts ...func(*Coupon) *Coupon) *Coupon {
	now := time.Now()
	c := &Coupon{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		Code:           gofakeit.AppName(),
		DiscountType:   CouponDiscountPercent,
		DiscountValue:  gofakeit.Float64Range(1, 100),
		PriceRuleID:    helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		UsageLimit:     helper.Ptr(gofakeit.Number(1, 100)),
		UsedCount:      gofakeit.Number(0, 10),
		ExpiryDate:     helper.Ptr(time.Now().AddDate(0, 1, 0)),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
