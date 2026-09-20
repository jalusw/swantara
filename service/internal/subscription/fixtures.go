package subscription

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func SubscriptionFixture(opts ...func(*Subscription) *Subscription) *Subscription {
	now := time.Now()
	sub := &Subscription{
		Base:            model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Name:            gofakeit.AppName(),
		ContactID:       helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		PlanID:          helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		PriceBookID:     helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		CurrencyCode:    helper.Ptr("IDR"),
		DateStart:       helper.Ptr(time.Now()),
		NextInvoiceDate: helper.Ptr(time.Now().AddDate(0, 1, 0)),
		DateEnd:         helper.Ptr(time.Now().AddDate(1, 0, 0)),
		State:           SubscriptionStateDraft,
		MRR:             gofakeit.Float64Range(1, 10000),
	}
	for _, opt := range opts {
		opt(sub)
	}
	return sub
}

func SubscriptionLineFixture(opts ...func(*SubscriptionLine) *SubscriptionLine) *SubscriptionLine {
	now := time.Now()
	line := &SubscriptionLine{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		SubscriptionID: uint64(gofakeit.Number(1, 10000)),
		ItemID:         helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Qty:            gofakeit.Float64Range(1, 100),
		UnitPrice:      gofakeit.Float64Range(1, 10000),
		DiscountPct:    gofakeit.Float64Range(0, 100),
	}
	for _, opt := range opts {
		opt(line)
	}
	return line
}
