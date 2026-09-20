package asset

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func FixedAssetFixture(opts ...func(*FixedAsset) *FixedAsset) *FixedAsset {
	now := time.Now()
	asset := &FixedAsset{
		Base:            model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID:  uint64(gofakeit.Number(1, 10000)),
		Name:            gofakeit.AppName(),
		CategoryID:      uint64(gofakeit.Number(1, 10000)),
		PurchaseValue:   gofakeit.Float64Range(1, 10000),
		SalvageValue:    gofakeit.Float64Range(1, 1000),
		AcquisitionDate: helper.Ptr(time.Now()),
		InServiceDate:   helper.Ptr(time.Now()),
		State:           AssetStateDraft,
	}
	for _, opt := range opts {
		opt(asset)
	}
	return asset
}

func AssetDepreciationLineFixture(opts ...func(*AssetDepreciationLine) *AssetDepreciationLine) *AssetDepreciationLine {
	now := time.Now()
	line := &AssetDepreciationLine{
		Base:             model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		AssetID:          uint64(gofakeit.Number(1, 10000)),
		Sequence:         gofakeit.Number(1, 100),
		DepreciationDate: now,
		Amount:           gofakeit.Float64Range(1, 10000),
		Accumulated:      gofakeit.Float64Range(1, 10000),
		RemainingValue:   gofakeit.Float64Range(1, 10000),
		Posted:           false,
	}
	for _, opt := range opts {
		opt(line)
	}
	return line
}
