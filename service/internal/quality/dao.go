package quality

import (
	"context"
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type QualityPointDAO interface {
	dao.CRUD[reference.QualityPoint]
	ListByItem(ctx context.Context, itemID uint64) ([]*reference.QualityPoint, error)
}

type qualityPointDAO struct {
	dao.Base[reference.QualityPoint]
	db *gorm.DB
}

func NewQualityPointDAO(db *gorm.DB) QualityPointDAO {
	return qualityPointDAO{Base: dao.NewBase[reference.QualityPoint](db), db: db}
}

func (d qualityPointDAO) ListByItem(ctx context.Context, itemID uint64) ([]*reference.QualityPoint, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "item_id", Operator: query.Equal, Value: itemID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type QualityCheckDAO interface {
	dao.CRUD[QualityCheck]
	ListByShipment(ctx context.Context, shipmentID uint64) ([]*QualityCheck, error)
	ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[QualityCheck], error)
	FindInOrg(ctx context.Context, id, organizationID uint64) (*QualityCheck, error)
}

type qualityCheckDAO struct {
	dao.Base[QualityCheck]
	db *gorm.DB
}

func NewQualityCheckDAO(db *gorm.DB) QualityCheckDAO {
	return qualityCheckDAO{Base: dao.NewBase[QualityCheck](db), db: db}
}

func (d qualityCheckDAO) ListByShipment(ctx context.Context, shipmentID uint64) ([]*QualityCheck, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "shipment_id", Operator: query.Equal, Value: shipmentID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d qualityCheckDAO) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[QualityCheck], error) {
	var count int64
	var entities []QualityCheck

	join := "JOIN item_variants ON item_variants.id = quality_checks.item_id " +
		"JOIN items ON items.id = item_variants.item_id " +
		"AND items.organization_id = ?"

	countTx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		countTx = query.ApplyFilters(countTx, q.Filters)
	}
	if err := countTx.Count(&count).Error; err != nil {
		return nil, err
	}

	tx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		tx = query.ApplyQuery(tx, q)
	}
	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}

	items := make([]*QualityCheck, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}

	return &query.Page[QualityCheck]{Items: items, Count: count}, nil
}

func (d qualityCheckDAO) FindInOrg(ctx context.Context, id, organizationID uint64) (*QualityCheck, error) {
	var check QualityCheck
	err := d.db.WithContext(ctx).Model(&QualityCheck{}).
		Joins("JOIN item_variants ON item_variants.id = quality_checks.item_id "+
			"JOIN items ON items.id = item_variants.item_id "+
			"AND items.organization_id = ?", organizationID).
		Where("quality_checks.id = ?", id).
		Take(&check).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &check, nil
}

type QualityAlertDAO interface {
	dao.CRUD[QualityAlert]
	ListByCheck(ctx context.Context, checkID uint64) ([]*QualityAlert, error)
	ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[QualityAlert], error)
	FindInOrg(ctx context.Context, id, organizationID uint64) (*QualityAlert, error)
}

type qualityAlertDAO struct {
	dao.Base[QualityAlert]
	db *gorm.DB
}

func NewQualityAlertDAO(db *gorm.DB) QualityAlertDAO {
	return qualityAlertDAO{Base: dao.NewBase[QualityAlert](db), db: db}
}

func (d qualityAlertDAO) ListByCheck(ctx context.Context, checkID uint64) ([]*QualityAlert, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "check_id", Operator: query.Equal, Value: checkID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d qualityAlertDAO) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[QualityAlert], error) {
	var count int64
	var entities []QualityAlert

	join := "JOIN item_variants ON item_variants.id = quality_alerts.item_id " +
		"JOIN items ON items.id = item_variants.item_id " +
		"AND items.organization_id = ?"

	countTx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		countTx = query.ApplyFilters(countTx, q.Filters)
	}
	if err := countTx.Count(&count).Error; err != nil {
		return nil, err
	}

	tx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		tx = query.ApplyQuery(tx, q)
	}
	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}

	items := make([]*QualityAlert, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}

	return &query.Page[QualityAlert]{Items: items, Count: count}, nil
}

func (d qualityAlertDAO) FindInOrg(ctx context.Context, id, organizationID uint64) (*QualityAlert, error) {
	var alert QualityAlert
	err := d.db.WithContext(ctx).Model(&QualityAlert{}).
		Joins("JOIN item_variants ON item_variants.id = quality_alerts.item_id "+
			"JOIN items ON items.id = item_variants.item_id "+
			"AND items.organization_id = ?", organizationID).
		Where("quality_alerts.id = ?", id).
		Take(&alert).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &alert, nil
}
