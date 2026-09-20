package asset

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type FixedAssetDAOMock struct {
	dao.CRUDMock[FixedAsset]
	UpdateTxFunc func(ctx context.Context, tx *gorm.DB, entity *FixedAsset) (*FixedAsset, error)
}

func (m FixedAssetDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, entity *FixedAsset) (*FixedAsset, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}

type AssetDepreciationLineDAOMock struct {
	dao.CRUDMock[AssetDepreciationLine]
	ListByAssetFunc func(ctx context.Context, assetID uint64) ([]*AssetDepreciationLine, error)
	CreateTxFunc    func(ctx context.Context, tx *gorm.DB, entity *AssetDepreciationLine) (*AssetDepreciationLine, error)
	UpdateTxFunc    func(ctx context.Context, tx *gorm.DB, entity *AssetDepreciationLine) (*AssetDepreciationLine, error)
}

func (m AssetDepreciationLineDAOMock) ListByAsset(ctx context.Context, assetID uint64) ([]*AssetDepreciationLine, error) {
	if m.ListByAssetFunc != nil {
		return m.ListByAssetFunc(ctx, assetID)
	}
	return []*AssetDepreciationLine{}, nil
}

func (m AssetDepreciationLineDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, entity *AssetDepreciationLine) (*AssetDepreciationLine, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}

func (m AssetDepreciationLineDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, entity *AssetDepreciationLine) (*AssetDepreciationLine, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}
