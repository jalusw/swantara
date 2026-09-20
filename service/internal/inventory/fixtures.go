package inventory

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func ShipmentFixture(opts ...func(*Shipment) *Shipment) *Shipment {
	now := time.Now()
	sp := &Shipment{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		Name:           helper.Ptr(gofakeit.AppName()),
		Type:           ShipmentTypeIncoming,
		ContactID:      helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		SrcLocationID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		DstLocationID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		State:          ShipmentStateDraft,
		ScheduledDate:  helper.Ptr(time.Now()),
		DateDone:       helper.Ptr(time.Now().AddDate(0, 0, 1)),
		Origin:         helper.Ptr(gofakeit.AppName()),
		CarrierID:      helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		TrackingRef:    helper.Ptr(gofakeit.AppName()),
	}
	for _, opt := range opts {
		opt(sp)
	}
	return sp
}

func StockMovementFixture(opts ...func(*StockMovement) *StockMovement) *StockMovement {
	now := time.Now()
	sm := &StockMovement{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		ShipmentID:     helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		ItemID:         uint64(gofakeit.Number(1, 10000)),
		Qty:            gofakeit.Float64Range(1, 10000),
		UnitID:         helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		SrcLocationID:  uint64(gofakeit.Number(1, 10000)),
		DstLocationID:  uint64(gofakeit.Number(1, 10000)),
		BatchID:        helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		State:          MovementStateDraft,
		UnitCost:       helper.Ptr(gofakeit.Float64Range(1, 10000)),
		OriginType:     helper.Ptr("purchase_order"),
		OriginID:       helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		ScheduledDate:  helper.Ptr(time.Now()),
		DateDone:       helper.Ptr(time.Now().AddDate(0, 0, 1)),
	}
	for _, opt := range opts {
		opt(sm)
	}
	return sm
}

func StockBalanceFixture(opts ...func(*StockBalance) *StockBalance) *StockBalance {
	now := time.Now()
	sq := &StockBalance{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		ItemID:         uint64(gofakeit.Number(1, 10000)),
		LocationID:     uint64(gofakeit.Number(1, 10000)),
		BatchID:        helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Quantity:       gofakeit.Float64Range(1, 10000),
		ReservedQty:    gofakeit.Float64Range(0, 1000),
	}
	for _, opt := range opts {
		opt(sq)
	}
	return sq
}

func BatchFixture(opts ...func(*Batch) *Batch) *Batch {
	now := time.Now()
	sl := &Batch{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ItemID:         uint64(gofakeit.Number(1, 10000)),
		Name:           gofakeit.AppName(),
		Ref:            helper.Ptr(gofakeit.AppName()),
		ExpiryDate:     helper.Ptr(time.Now().AddDate(0, 6, 0)),
		BestBeforeDate: helper.Ptr(time.Now().AddDate(0, 3, 0)),
	}
	for _, opt := range opts {
		opt(sl)
	}
	return sl
}
