package returns

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func RMAFixture(opts ...func(*RMA) *RMA) *RMA {
	now := gofakeit.Date()
	r := &RMA{
		Base:            model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Name:            helper.Ptr(gofakeit.AppName()),
		Type:            TypeCustomerReturn,
		ContactID:       uint64(gofakeit.Number(1, 10000)),
		OriginOrderType: helper.Ptr("sale_order"),
		OriginOrderID:   helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Reason:          helper.Ptr(gofakeit.Sentence(3)),
		State:           StateDraft,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func RMALineFixture(opts ...func(*RMALine) *RMALine) *RMALine {
	now := gofakeit.Date()
	l := &RMALine{
		Base:            model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		RMAID:           uint64(gofakeit.Number(1, 10000)),
		ItemID:          uint64(gofakeit.Number(1, 10000)),
		Qty:             gofakeit.Float64Range(1, 100),
		BatchID:         helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Disposition:     DispositionRestock,
		StockMovementID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		CreditNoteID:    helper.Ptr(uint64(gofakeit.Number(1, 10000))),
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}
