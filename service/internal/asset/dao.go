package asset

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type FixedAssetDAO interface {
	dao.CRUD[FixedAsset]
	UpdateTx(ctx context.Context, tx *gorm.DB, entity *FixedAsset) (*FixedAsset, error)
}

type fixedAssetDAO struct {
	dao.Base[FixedAsset]
	db *gorm.DB
}

func NewFixedAssetDAO(db *gorm.DB) FixedAssetDAO {
	return fixedAssetDAO{Base: dao.NewBase[FixedAsset](db), db: db}
}

func (d fixedAssetDAO) UpdateTx(ctx context.Context, tx *gorm.DB, entity *FixedAsset) (*FixedAsset, error) {
	if err := tx.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

type AssetDepreciationLineDAO interface {
	dao.CRUD[AssetDepreciationLine]
	ListByAsset(ctx context.Context, assetID uint64) ([]*AssetDepreciationLine, error)
	CreateTx(ctx context.Context, tx *gorm.DB, entity *AssetDepreciationLine) (*AssetDepreciationLine, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, entity *AssetDepreciationLine) (*AssetDepreciationLine, error)
}

type assetDepreciationLineDAO struct {
	dao.Base[AssetDepreciationLine]
	db *gorm.DB
}

func NewAssetDepreciationLineDAO(db *gorm.DB) AssetDepreciationLineDAO {
	return assetDepreciationLineDAO{Base: dao.NewBase[AssetDepreciationLine](db), db: db}
}

func (d assetDepreciationLineDAO) ListByAsset(ctx context.Context, assetID uint64) ([]*AssetDepreciationLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "asset_id", Operator: query.Equal, Value: assetID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d assetDepreciationLineDAO) CreateTx(ctx context.Context, tx *gorm.DB, entity *AssetDepreciationLine) (*AssetDepreciationLine, error) {
	if err := tx.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (d assetDepreciationLineDAO) UpdateTx(ctx context.Context, tx *gorm.DB, entity *AssetDepreciationLine) (*AssetDepreciationLine, error) {
	if err := tx.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}
