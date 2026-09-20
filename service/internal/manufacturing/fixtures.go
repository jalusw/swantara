package manufacturing

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func RecipeFixture(opts ...func(*Recipe) *Recipe) *Recipe {
	now := time.Now()
	b := &Recipe{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		ItemID:         uint64(gofakeit.Number(1, 10000)),
		Code:           helper.Ptr(gofakeit.AppName()),
		Qty:            gofakeit.Float64Range(1, 10000),
		UnitID:         helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Type:           RecipeTypeManufacture,
		Version:        gofakeit.Number(1, 10),
		Active:         gofakeit.Bool(),
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func RecipeLineFixture(opts ...func(*RecipeLine) *RecipeLine) *RecipeLine {
	now := time.Now()
	bl := &RecipeLine{
		Base:             model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		RecipeID:         uint64(gofakeit.Number(1, 10000)),
		ComponentID:      uint64(gofakeit.Number(1, 10000)),
		Qty:              gofakeit.Float64Range(1, 10000),
		UnitID:           helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		ScrapPct:         gofakeit.Float64Range(0, 50),
		ProductionStepID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
	}
	for _, opt := range opts {
		opt(bl)
	}
	return bl
}

func ProductionOrderFixture(opts ...func(*ProductionOrder) *ProductionOrder) *ProductionOrder {
	now := time.Now()
	productionOrder := &ProductionOrder{
		Base:              model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID:    helper.Ptr(uint64(gofakeit.Number(1, 100))),
		Name:              helper.Ptr(gofakeit.AppName()),
		ItemID:            uint64(gofakeit.Number(1, 10000)),
		RecipeID:          helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		QtyToProduce:      gofakeit.Float64Range(1, 10000),
		QtyProduced:       gofakeit.Float64Range(0, 1000),
		UnitID:            helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		SrcLocationID:     helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		DstLocationID:     helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		State:             ProductionOrderStateDraft,
		DatePlannedStart:  helper.Ptr(time.Now()),
		DatePlannedFinish: helper.Ptr(time.Now().AddDate(0, 0, 7)),
		DateStart:         helper.Ptr(time.Now()),
		DateFinished:      helper.Ptr(time.Now().AddDate(0, 0, 7)),
		Origin:            helper.Ptr(gofakeit.AppName()),
		Priority:          gofakeit.Number(0, 3),
	}
	for _, opt := range opts {
		opt(productionOrder)
	}
	return productionOrder
}

func ConsumedMaterialFixture(opts ...func(*ConsumedMaterial) *ConsumedMaterial) *ConsumedMaterial {
	now := time.Now()
	mc := &ConsumedMaterial{
		Base:              model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ProductionOrderID: uint64(gofakeit.Number(1, 10000)),
		ItemID:            uint64(gofakeit.Number(1, 10000)),
		QtyPlanned:        gofakeit.Float64Range(1, 10000),
		QtyConsumed:       gofakeit.Float64Range(0, 1000),
		UnitID:            helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		StockMovementID:   helper.Ptr(uint64(gofakeit.Number(1, 10000))),
	}
	for _, opt := range opts {
		opt(mc)
	}
	return mc
}

func ShopTaskFixture(opts ...func(*ShopTask) *ShopTask) *ShopTask {
	now := time.Now()
	wo := &ShopTask{
		Base:              model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID:    helper.Ptr(uint64(gofakeit.Number(1, 100))),
		ProductionOrderID: uint64(gofakeit.Number(1, 10000)),
		ProductionStepID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		WorkCenterID:      uint64(gofakeit.Number(1, 10000)),
		Name:              helper.Ptr(gofakeit.AppName()),
		State:             ShopTaskStateDraft,
		Sequence:          gofakeit.Number(1, 100),
		PlannedStart:      helper.Ptr(time.Now()),
		PlannedFinish:     helper.Ptr(time.Now().AddDate(0, 0, 3)),
		DateStart:         helper.Ptr(time.Now()),
		DateFinished:      helper.Ptr(time.Now().AddDate(0, 0, 3)),
		PlannedMinutes:    gofakeit.Float64Range(1, 480),
		ActualMinutes:     gofakeit.Float64Range(0, 480),
	}
	for _, opt := range opts {
		opt(wo)
	}
	return wo
}

func OutsideProcessingOrderFixture(opts ...func(*OutsideProcessingOrder) *OutsideProcessingOrder) *OutsideProcessingOrder {
	now := time.Now()
	so := &OutsideProcessingOrder{
		Base:              model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ProductionOrderID: uint64(gofakeit.Number(1, 10000)),
		SupplierID:        uint64(gofakeit.Number(1, 10000)),
		PurchaseOrderID:   helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		State:             OutsideProcessingStateDraft,
	}
	for _, opt := range opts {
		opt(so)
	}
	return so
}

func ProductionStepFixture(opts ...func(*ProductionStep) *ProductionStep) *ProductionStep {
	now := time.Now()
	ro := &ProductionStep{
		Base:         model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		RecipeID:     helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		WorkCenterID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Name:         helper.Ptr(gofakeit.AppName()),
		Sequence:     gofakeit.Number(1, 100),
		SetupMinutes: gofakeit.Float64Range(0, 120),
		TimeMinutes:  gofakeit.Float64Range(1, 480),
	}
	for _, opt := range opts {
		opt(ro)
	}
	return ro
}
